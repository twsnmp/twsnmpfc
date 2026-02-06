package backend

import (
	"context"
	"fmt"
	"log"
	"strconv"
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
						go checkNetwork(n)
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
	findLineFromFDB(n, ret)
	return ret
}

func findLineFromFDB(n *datastore.NetworkEnt, ret *FindNeighborNetworksAndLinesResp) {
	list := datastore.GetFDBTable(n.ID)
	if list == nil {
		return
	}
	for _, e := range *list {
		node := datastore.FindNodeFromMAC(e.MAC)
		if node == nil {
			continue
		}
		pid := ""
		pcmp := fmt.Sprintf("ifOperStatus.%d", e.IfIndex)
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
		idx := fmt.Sprintf("%d", e.IfIndex)
		for _, lp := range n.Ports {
			if lp.Index == idx {
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
}

func checkNetwork(n *datastore.NetworkEnt) {
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
	checkNetworkPortState(n, agent)
	if n.ArpWatch {
		checkNetworkArpWatch(agent)
	}
	if n.PortWatch {
		checkNetworkIfPorts(n, agent)
		checkNetworkFDBTable(n, agent)
	}
}

func checkNetworkPortState(n *datastore.NetworkEnt, agent *gosnmp.GoSNMP) {
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
}

func getSysUpTime(agent *gosnmp.GoSNMP) (uint64, error) {
	r, err := agent.Get([]string{datastore.MIBDB.NameToOID("sysUpTime.0")})
	if err != nil {
		return 0, err
	}
	if len(r.Variables) < 1 {
		return 0, fmt.Errorf("cant not get sysuptime")
	}
	return gosnmp.ToBigInt(r.Variables[0].Value).Uint64(), nil
}

func checkNetworkIfPorts(n *datastore.NetworkEnt, agent *gosnmp.GoSNMP) {
	// sysUpTimeを取得する
	sysUpTime, err := getSysUpTime(agent)
	if err != nil {
		log.Println(err)
		return
	}
	ifIndexToIfPortEntMap := make(map[int]*datastore.IfPortEnt)
	now := time.Now().UnixNano()
	err = agent.Walk(datastore.MIBDB.NameToOID("ifTable"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		idx, err := strconv.Atoi(a[1])
		if err != nil {
			return nil
		}
		switch a[0] {
		case "ifIndex":
			ifIndexToIfPortEntMap[idx] = &datastore.IfPortEnt{
				IfIndex:         idx,
				FirstCheckTime:  now,
				LastCheckTime:   now,
				LastChangedTime: now,
			}
		case "ifDescr":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.Descr = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "ifType":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.Type = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "ifPhysAddress":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.MAC = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "ifMtu":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.Mtu = gosnmp.ToBigInt(variable.Value).Uint64()
			}
		case "ifSpeed":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.Speed = gosnmp.ToBigInt(variable.Value).Uint64()
			}
		case "ifInOctets":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.InOctets = gosnmp.ToBigInt(variable.Value).Uint64()
			}
		case "ifOutOctets":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.OutOctets = gosnmp.ToBigInt(variable.Value).Uint64()
			}
		case "ifLastChange":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				lt := gosnmp.ToBigInt(variable.Value).Uint64()
				p.LastChangedTime = now - int64((sysUpTime-lt)*10*1000*1000)
			}
		case "ifAdminStatus":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.AdminStatus = int(gosnmp.ToBigInt(variable.Value).Uint64())
			}
		case "ifOperStatus":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.OperStatus = int(gosnmp.ToBigInt(variable.Value).Uint64())
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("get ifTable err=%v", err)
		return
	}
	err = agent.Walk(datastore.MIBDB.NameToOID("ifXTable"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		idx, err := strconv.Atoi(a[1])
		if err != nil {
			return nil
		}
		switch a[0] {
		case "ifName":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.Name = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "ifHCInOctets":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.InOctets = gosnmp.ToBigInt(variable.Value).Uint64()
			}
		case "ifHCOutOctets":
			if p, ok := ifIndexToIfPortEntMap[idx]; ok {
				p.OutOctets = gosnmp.ToBigInt(variable.Value).Uint64()
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("get ifXTable err=%v", err)
	}
	old := datastore.GetIfPortTable(n.ID)
	if old != nil {
		for _, e := range *old {
			if p, ok := ifIndexToIfPortEntMap[e.IfIndex]; ok {
				p.FirstCheckTime = e.FirstCheckTime
				diff := p.LastCheckTime - e.LastCheckTime
				if diff > 0 {
					p.InBPS = (1000 * 1000 * 1000 * (p.InOctets - e.InOctets)) / uint64(diff)
					p.OutBPS = (1000 * 1000 * 1000 * (p.OutOctets - e.OutOctets)) / uint64(diff)
				}
				p.Changed = e.Changed
				if p.LastChangedTime > e.LastChangedTime+2*1000*1000*1000 || p.LastChangedTime < e.LastChangedTime-2*1000*1000*1000 {
					log.Printf("lastChange %+v != %+v ", p, e)
					p.Changed++
				} else {
					p.LastChangedTime = e.LastChangedTime
				}
			}
		}
	}
	list := []datastore.IfPortEnt{}
	for _, p := range ifIndexToIfPortEntMap {
		list = append(list, *p)
	}
	datastore.UpdateIfPortTable(n.ID, &list)
}

func checkNetworkFDBTable(n *datastore.NetworkEnt, agent *gosnmp.GoSNMP) {
	// ブリッジのポートからifIndexに変換するテーブルの作成
	portToIFIndexMap := make(map[int]int)
	err := agent.Walk(datastore.MIBDB.NameToOID("dot1dBasePortIfIndex"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		idx, err := strconv.Atoi(a[1])
		if err != nil {
			return nil
		}
		switch a[0] {
		case "dot1dBasePortIfIndex":
			portToIFIndexMap[idx] = int(gosnmp.ToBigInt(variable.Value).Int64())
		}
		return nil
	})
	if err != nil {
		log.Printf("get dot1dBasePortIfIndex  err=%v", err)
		return
	}
	list := []datastore.FDBTableEnt{}
	macToIndexMap := make(map[string]int)
	err = agent.Walk(datastore.MIBDB.NameToOID("dot1qTpFdbPort"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) != 1+1+6 {
			return nil
		}
		vlan, err := strconv.Atoi(a[1])
		if err != nil {
			return nil
		}
		mac, err := indexToMacAddress(a[2:])
		if err != nil {
			log.Println(err)
			return nil
		}
		now := time.Now().UnixNano()
		switch a[0] {
		case "dot1qTpFdbPort":
			port := int(gosnmp.ToBigInt(variable.Value).Int64())
			if idx, ok := portToIFIndexMap[port]; ok {
				list = append(list, datastore.FDBTableEnt{
					MAC:             mac,
					VLanID:          vlan,
					Port:            port,
					IfIndex:         idx,
					FirstCheckTime:  now,
					LastCheckTime:   now,
					Changed:         0,
					LastChangedTime: now,
				})
				macToIndexMap[mac] = len(list) - 1
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("get dot1qTpFdbPort err=%v", err)
		return
	}
	old := datastore.GetFDBTable(n.ID)
	if old != nil {
		for _, e := range *old {
			if i, ok := macToIndexMap[e.MAC]; ok {
				list[i].FirstCheckTime = e.FirstCheckTime
				list[i].Changed = e.Changed
				if list[i].Port == e.Port {
					list[i].LastChangedTime = e.LastChangedTime
				} else {
					mac := e.MAC
					if nn := datastore.FindNodeFromMAC(mac); nn != nil {
						mac = fmt.Sprintf("%s(%s)", nn.Name, mac)
					}
					list[i].Changed++
					// 別のポートに移動した
					datastore.AddEventLog(&datastore.EventLogEnt{
						NodeName: n.Name,
						NodeID:   n.ID,
						Level:    "warn",
						Type:     "FDBWatch",
						Event:    fmt.Sprintf("%sは%dポートから%dに移動しました", mac, e.Port, list[i].Port),
					})
				}
				delete(macToIndexMap, e.MAC)
			} else {
				// ポートから切断された
				mac := e.MAC
				if nn := datastore.FindNodeFromMAC(mac); nn != nil {
					mac = fmt.Sprintf("%s(%s)", nn.Name, mac)
				}
				datastore.AddEventLog(&datastore.EventLogEnt{
					NodeName: n.Name,
					NodeID:   n.ID,
					Level:    "warn",
					Type:     "FDBWatch",
					Event:    fmt.Sprintf("%sは%dポートから切断されました", mac, e.Port),
				})
			}
		}
	}
	for mac, i := range macToIndexMap {
		//新規接続のMACアドレス
		if nn := datastore.FindNodeFromMAC(mac); nn != nil {
			mac = fmt.Sprintf("%s(%s)", nn.Name, mac)
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			NodeName: n.Name,
			NodeID:   n.ID,
			Level:    "info",
			Type:     "FDBWatch",
			Event:    fmt.Sprintf("%sが%dポートに接続されています", mac, list[i].Port),
		})
	}
	datastore.UpdateFDBTable(n.ID, &list)
}

func indexToMacAddress(a []string) (string, error) {
	ret := []string{}
	for _, s := range a {
		if i, err := strconv.Atoi(s); err == nil {
			ret = append(ret, fmt.Sprintf("%02X", i))
		} else {
			return "", err
		}
	}
	return strings.Join(ret, ":"), nil
}

func checkNetworkArpWatch(agent *gosnmp.GoSNMP) {
	log.Printf("arp watch")
	arpMap := make(map[string]string)
	err := agent.Walk(datastore.MIBDB.NameToOID("ipNetToMediaPhysAddress"), func(variable gosnmp.SnmpPDU) error {
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
