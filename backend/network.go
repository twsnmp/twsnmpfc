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
)

func networkBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start network backend")
	checkNetworkMap := make(map[string]int64)
	now := time.Now().Unix()
	j := 0
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
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
		case <-timer.C:
			now = time.Now().Unix()
			j = 0
			datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
				if len(n.Ports) < 1 && n.Error == "" {
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
					go checkNetworkPortState(n)
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
		n.Error = fmt.Sprintf("SNMP Connect err=%s", err)
		log.Printf("getNetworkPorts err=%v", err)
		return
	}
	defer agent.Conn.Close()
	setName := n.Name == ""
	setDescr := n.Descr == ""
	x := 0
	y := 0
	// LLDP-MIBの対応をチェック
	err = agent.Walk(datastore.MIBDB.NameToOID("lldpLocalSystemData"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		switch a[0] {
		case "lldpLocSysName":
			if setName {
				n.Name = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpLocSysDescr":
			if setDescr {
				n.Descr = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpLocSysCapEnabled":
			if setDescr {
				n.Descr += " " + datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpLocPortDesc":
			n.Ports = append(n.Ports, datastore.PortEnt{
				Name:    datastore.GetMIBValueString(a[0], &variable, false),
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
		} else {
			log.Println(err)
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
		n.Error = fmt.Sprintf("SNMP get err=%v", err)
		log.Println(err)
		return
	}
	for _, index := range ifIndexs {
		name := fmt.Sprintf("ifDescr.%s", index)
		oid := datastore.MIBDB.NameToOID(name)
		r, err := agent.Get([]string{oid})
		if err != nil {
			log.Println(err)
			continue
		}
		for _, variable := range r.Variables {
			if datastore.MIBDB.OIDToName(variable.Name) == name {
				n.Ports = append(n.Ports, datastore.PortEnt{
					Name:    datastore.GetMIBValueString(name, &variable, false),
					X:       x,
					Y:       y,
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

func checkNetworkPortState(n *datastore.NetworkEnt) {
	agent := getSNMPAgentForNetwork(n)
	if agent == nil {
		n.Error = "SNMPのパラメータエラー"
		return
	}
	err := agent.Connect()
	if err != nil {
		n.Error = fmt.Sprintf("SNMP Connect err=%s", err)
		log.Printf("checkNetworkPortState err=%v", err)
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
			n.Error = fmt.Sprintf("SNMP get %s err=%s", p.Polling, err)
			log.Printf("checkNetworkPortState err=%v", err)
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
