package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
	"github.com/twsnmp/twsnmpfc/report"
)

var checkNetworkCh = make(chan string)

func networkBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start network backend")
	checkNetworkMap := make(map[string]int64)
	now := time.Now().Unix()
	j := 0
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		n.Error = ""
		if len(n.Ports) > 0 {
			checkNetworkMap[n.ID] = now + int64(j)
			j++
			for i := range n.Ports {
				n.Ports[i].State = "unknown"
			}
		}
		return true
	})
	timer := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop network backend")
			return
		case id := <-checkNetworkCh:
			delete(checkNetworkMap, id)
		case <-timer.C:
			now = time.Now().Unix()
			j = 0
			datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
				if n.Error != "" && len(n.Ports) < 1 {
					return true
				}
				if len(n.Ports) < 1 {
					log.Printf("get network port=%s", n.IP)
					go getNetworkPorts(n)
				} else if _, ok := checkNetworkMap[n.ID]; !ok {
					checkNetworkMap[n.ID] = now + int64(j)
					j++
				}
				return true
			})
			for id, t := range checkNetworkMap {
				n := datastore.GetNetwork(id)
				if n == nil {
					delete(checkNetworkMap, id)
					continue
				}
				if t < now {
					if n.Unmanaged {
						go checkUnmanagedNetworkPortState(n)
					} else {
						go checkNetworkPortState(n)
					}
					checkNetworkMap[id] = now + 60
				}
			}
		}
	}
}

func getNetworkPorts(n *datastore.NetworkEnt) {
	agent := getSNMPAgentForNetwork(n)
	if agent == nil {
		n.Error = "SNMPのパラメータエラー"
		return
	}
	err := agent.Connect()
	if err != nil {
		n.Error = fmt.Sprintf("SNMPアクセスエラー err=%s", err)
		return
	}
	defer agent.Conn.Close()
	setName := n.Name == ""
	setDescr := n.Descr == ""
	portMap := make(map[string]string)
	x := 0
	y := 0
	// LLDP-MIBの対応をチェック
	err = agent.Walk(datastore.MIBDB.NameToOID("lldpLocalSystemData"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		switch a[0] {
		case "lldpLocChassisId":
			n.SystemID = datastore.GetMIBValueString(a[0], &variable, true)
		case "lldpLocSysName":
			if setName {
				n.Name = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpLocSysDesc":
			if setDescr {
				n.Descr = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpLocSysCapEnabled":
			if setDescr {
				n.Descr += " " + datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpLocPortId":
			portMap[a[1]] = datastore.GetMIBValueString(a[0], &variable, true)
		case "lldpLocPortDesc":
			id, ok := portMap[a[1]]
			if !ok {
				id = a[1]
			}
			name := datastore.GetMIBValueString(a[0], &variable, false)
			if name == "" {
				name = id
			}
			n.Ports = append(n.Ports, datastore.PortEnt{
				Name:    name,
				ID:      id,
				Index:   a[1],
				X:       x,
				Y:       y,
				Polling: fmt.Sprintf("ifOperStatus.%s", a[1]),
			})
			x++
			if x > 24 {
				y++
				x = 0
			}
		}
		return nil
	})
	if err == nil && len(n.Ports) > 0 {
		n.LLDP = true
		datastore.UpdateNetwork(n)
		log.Printf("found network %+v", n)
		return
	}
	// LLDP-MIBの未対応の場合
	if setName || setDescr {
		if r, err := agent.Get([]string{
			datastore.MIBDB.NameToOID("sysName.0"),
			datastore.MIBDB.NameToOID("sysDescr.0"),
		}); err == nil {
			for _, variable := range r.Variables {
				a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
				if len(a) != 2 {
					continue
				}
				switch a[0] {
				case "sysName":
					if setName {
						n.Name = datastore.GetMIBValueString(a[0], &variable, false)
					}
				case "sysDescr":
					if setDescr {
						n.Descr = datastore.GetMIBValueString(a[0], &variable, false)
					}
				}
			}
		}
	}
	ifIndexs := []string{}
	err = agent.Walk(datastore.MIBDB.NameToOID("ifType"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		// Ethernet ONLY
		if getMIBStringVal(variable.Value) == "6" {
			ifIndexs = append(ifIndexs, a[1])
		}
		return nil
	})
	if err != nil {
		n.Error = fmt.Sprintf("SNMP取得エラー err=%v", err)
		return
	}
	for _, index := range ifIndexs {
		name := fmt.Sprintf("ifName.%s", index)
		oid := datastore.MIBDB.NameToOID(name)
		r, err := agent.Get([]string{oid})
		if err != nil {
			name = fmt.Sprintf("ifDescr.%s", index)
			oid = datastore.MIBDB.NameToOID(name)
			r, err = agent.Get([]string{oid})
			if err != nil {
				continue
			}
		}
		for _, variable := range r.Variables {
			if datastore.MIBDB.OIDToName(variable.Name) == name {
				pn := datastore.GetMIBValueString(name, &variable, true)
				if pn == "" {
					pn = "#" + index
				}
				n.Ports = append(n.Ports, datastore.PortEnt{
					Name:    pn,
					X:       x,
					Y:       y,
					ID:      index,
					Index:   index,
					Polling: fmt.Sprintf("ifOperStatus.%s", index),
				})
				x++
				if x > 24 {
					y++
					x = 0
				}
			}
		}
	}
	if len(n.Ports) > 0 {
		log.Printf("found network %+v", n)
		datastore.UpdateNetwork(n)
	}
}

type FindNeighborNetworksAndLinesResp struct {
	Networks []*datastore.NetworkEnt
	Lines    []*datastore.LineEnt
}

// FindNeighborNetworksAndLines 接続可能なLineと隣接する未登録のネットワークを検索する
func FindNeighborNetworksAndLines(n *datastore.NetworkEnt) *FindNeighborNetworksAndLinesResp {
	ret := &FindNeighborNetworksAndLinesResp{
		Networks: []*datastore.NetworkEnt{},
		Lines:    []*datastore.LineEnt{},
	}
	agent := getSNMPAgentForNetwork(n)
	if agent == nil {
		n.Error = "SNMPパラメータエラー"
		return ret
	}
	err := agent.Connect()
	if err != nil {
		n.Error = fmt.Sprintf("SNMP接続エラー err=%s", err)
		return ret
	}
	defer agent.Conn.Close()
	remoteMap := make(map[string]*datastore.NetworkEnt)
	// LLDP-MIBのlldpRemoteSystemsDataから隣接するNetworkを探す
	agent.Walk(datastore.MIBDB.NameToOID("lldpRemoteSystemsData"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		switch a[0] {
		case "lldpRemChassisId":
			remoteMap[a[1]] = &datastore.NetworkEnt{
				SystemID: datastore.GetMIBValueString(a[0], &variable, true),
			}
		case "lldpRemPortId":
			if rn, ok := remoteMap[a[1]]; ok {
				b := strings.Split(a[1], ".")
				if len(b) < 2 {
					return nil
				}
				id := datastore.GetMIBValueString(a[0], &variable, true)
				rn.Ports = append(rn.Ports, datastore.PortEnt{
					ID:    id,
					Index: b[1],
					Name:  id,
					X:     len(n.Ports),
				})
			}
		case "lldpRemSysName":
			if rn, ok := remoteMap[a[1]]; ok {
				rn.Name = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpRemSysDesc":
			if rn, ok := remoteMap[a[1]]; ok {
				rn.Descr = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpRemSysCapEnabled":
			if rn, ok := remoteMap[a[1]]; ok {
				rn.Descr += " " + datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpRemManAddrIfId":
			b := strings.Split(a[1], ".")
			if len(b) == 3+2+4 {
				if rn, ok := remoteMap[strings.Join(b[:3], ".")]; ok {
					rn.IP = strings.Join(b[5:], ".")
				}
			}
		}
		return nil
	})
	// 見つけた隣接ネットワークを確認する
	for _, rn := range remoteMap {
		rnr := datastore.FindNetwork(rn.SystemID, rn.IP)
		if rnr == nil {
			// 未登録
			rn.SnmpMode = n.SnmpMode
			rn.Community = n.Community
			rn.Password = n.Password
			rn.User = n.User
			rn.HPorts = n.HPorts
			rn.Ports = []datastore.PortEnt{}
			rn.Y = n.Y + n.H
			rn.X = n.X
			ret.Networks = append(ret.Networks, rn)
		} else {
			// 登録済みならラインの候補に
			for _, rp := range rnr.Ports {
				for _, frp := range rn.Ports {
					log.Printf("rp=%+v frp=%+v", rp, frp)
					if frp.ID == rp.ID {
						for _, lp := range n.Ports {
							if lp.Index == frp.Index {
								l := &datastore.LineEnt{
									NodeID1:    fmt.Sprintf("NET:%s", n.ID),
									PollingID1: lp.ID,
									NodeID2:    fmt.Sprintf("NET:%s", rnr.ID),
									PollingID2: rp.ID,
									Width:      2,
								}
								if !datastore.HasLine(l, true) {
									ret.Lines = append(ret.Lines, l)
								}
							}
						}
					}
				}
			}
		}
	}
	// ARPテーブルから接続先を探す
	arpMap := make(map[string]string)
	err = agent.Walk(datastore.MIBDB.NameToOID("ipNetToMediaPhysAddress"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		switch a[0] {
		case "ipNetToMediaPhysAddress":
			arpMap[a[1]] = datastore.GetMIBValueString(a[0], &variable, true)
		}
		return nil
	})
	if err != nil {
		// ipNetToMediaPhysAddress 未対応
		agent.Walk(datastore.MIBDB.NameToOID("atPhysAddress"), func(variable gosnmp.SnmpPDU) error {
			a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
			if len(a) != 2 {
				return nil
			}
			switch a[0] {
			case "atPhysAddress":
				arpMap[a[1]] = datastore.GetMIBValueString(a[0], &variable, true)
			}
			return nil
		})
	}
	for index, mac := range arpMap {
		a := strings.Split(index, ".")
		if len(a) < 1+4 {
			continue
		}
		ip := strings.Join(a[len(a)-4:], ".")
		node := datastore.FindNodeFromIP(ip)
		if node == nil {
			node = datastore.FindNodeFromMAC(mac)
		}
		if node == nil {
			continue
		}
		pid := ""
		pcmp := fmt.Sprintf("ifOperStatus.%s", a[0])
		datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
			if p.NodeID == node.ID {
				if p.Type == "snmp" && p.Params == pcmp {
					pid = p.ID
					return false
				}
				if pid == "" {
					pid = p.ID
				} else if p.Type == "ping" {
					pid = p.ID
				}
			}
			return true
		})
		for _, lp := range n.Ports {
			if lp.Index == a[0] {
				l := &datastore.LineEnt{
					NodeID1:    fmt.Sprintf("NET:%s", n.ID),
					PollingID1: lp.ID,
					NodeID2:    node.ID,
					PollingID2: pid,
					Width:      2,
				}
				if !datastore.HasLine(l, false) {
					ret.Lines = append(ret.Lines, l)
				}
			}
		}
	}
	return ret
}

func checkNetworkPortState(n *datastore.NetworkEnt) {
	agent := getSNMPAgentForNetwork(n)
	if agent == nil {
		n.Error = "SNMPパラメータエラー"
		return
	}
	err := agent.Connect()
	if err != nil {
		n.Error = fmt.Sprintf("SNMPアクセス err=%s", err)
		return
	}
	defer agent.Conn.Close()

	for i, p := range n.Ports {
		n.Ports[i].State = "unknown"
		a := strings.SplitN(p.Polling, ":", 2)
		c := "1"
		if len(a) == 2 {
			c = a[1]
		}
		r, err := agent.Get([]string{datastore.MIBDB.NameToOID(a[0])})
		if err != nil {
			n.Error = fmt.Sprintf("SNMP%s取得 err=%s", p.Polling, err)
			continue
		}
		for _, variable := range r.Variables {
			if datastore.MIBDB.OIDToName(variable.Name) == a[0] {
				if getMIBStringVal(variable.Value) == c {
					n.Ports[i].State = "up"
				} else {
					n.Ports[i].State = "down"
				}
			}
		}
	}
	if !n.ArpWatch {
		return
	}
	// ARP監視
	arpMap := make(map[string]string)
	err = agent.Walk(datastore.MIBDB.NameToOID("ipNetToMediaPhysAddress"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		switch a[0] {
		case "ipNetToMediaPhysAddress":
			arpMap[a[1]] = datastore.GetMIBValueString(a[0], &variable, false)
		}
		return nil
	})
	if err != nil {
		// ipNetToMediaPhysAddress 未対応
		agent.Walk(datastore.MIBDB.NameToOID("atPhysAddress"), func(variable gosnmp.SnmpPDU) error {
			a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
			if len(a) != 2 {
				return nil
			}
			switch a[0] {
			case "atPhysAddress":
				arpMap[a[1]] = datastore.GetMIBValueString(a[0], &variable, false)
			}
			return nil
		})
	}
	for index, mac := range arpMap {
		a := strings.Split(index, ".")
		if len(a) < 1+4 {
			continue
		}
		ip := strings.Join(a[len(a)-4:], ".")
		report.ReportDevice(mac, ip, time.Now().UnixNano())
	}
}

func getSNMPAgentForNetwork(n *datastore.NetworkEnt) *gosnmp.GoSNMP {
	if strings.HasPrefix(n.SnmpMode, "v3") && n.User == "" {
		return nil
	} else if n.Community == "" {
		return nil
	}
	agent := &gosnmp.GoSNMP{
		Target:    n.IP,
		Port:      161,
		Transport: "udp",
		Community: n.Community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(datastore.MapConf.Timeout) * time.Second,
		Retries:   datastore.MapConf.Retry,
		MaxOids:   gosnmp.MaxOids,
	}
	switch n.SnmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        n.Password,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        n.Password,
		}
	}
	return agent
}

func CheckNetwork(id string) {
	n := datastore.GetNetwork(id)
	if n == nil {
		return
	}
	// Clear Error
	n.Error = ""
	checkNetworkCh <- id
}

func CheckAllNetworks() {
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		if n.Error != "" {
			n.Error = ""
			checkNetworkCh <- n.ID
		}
		return true
	})
}

func checkUnmanagedNetworkPortState(n *datastore.NetworkEnt) {
	if n.IP != "" {
		r := ping.DoPing(n.IP, datastore.MapConf.Timeout, datastore.MapConf.Retry, 64, 64)
		if r.Stat != ping.PingOK {
			n.Error = "Ping No Responce"
			for i := range n.Ports {
				n.Ports[i].State = "down"
			}
			return
		}
	}
	for i := range n.Ports {
		n.Ports[i].State = "unknown"
	}
	id := "NET:" + n.ID
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 == id {
			for i := range n.Ports {
				if n.Ports[i].ID == l.PollingID1 {
					n.Ports[i].State = getPortState(l.NodeID2, l.PollingID2)
				}
			}
		} else if l.NodeID2 == id {
			for i := range n.Ports {
				if n.Ports[i].ID == l.PollingID2 {
					n.Ports[i].State = getPortState(l.NodeID1, l.PollingID1)
				}
			}
		}
		return true
	})
}

func getPortState(nodeID, pollingID string) string {
	st := "unknown"
	if strings.HasPrefix(nodeID, "NET:") {
		n := datastore.GetNetwork(nodeID)
		if n != nil {
			if n.Unmanaged {
				// UnmanagedのHUB同士は接続するだけでUP
				st = "up"
			} else {
				// 相手の状態を取得
				for _, p := range n.Ports {
					if p.ID == pollingID {
						st = p.State
						break
					}
				}
			}
		}
	} else {
		poll := datastore.GetPolling(pollingID)
		if poll != nil {
			st = poll.State
		}
	}
	return st
}
