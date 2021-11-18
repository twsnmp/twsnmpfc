package backend

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/twsnmp/gosnmp"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type VPanelPortEnt struct {
	Index        int64
	State        string
	Name         string
	Speed        int64
	OutPacktes   int64
	OutBytes     int64
	OutError     int64
	InPacktes    int64
	InBytes      int64
	InError      int64
	Type         int64
	Admin        int64
	Oper         int64
	MAC          string
	pollingIndex string
}

// GetVPanelPowerInfo : パネルの電源状態を取得
func GetVPanelPowerInfo(n *datastore.NodeEnt) bool {
	// まずはノードの状態を反映
	state := n.State
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == n.ID && p.Type == "ping" {
			// PINGの状態を反映
			state = p.State
			return false
		}
		return true
	})
	return state == "normal" || state == "repair"
}

// GetVPanelPorts : パネルに表示するポートの情報を取得する
// 優先順位は
// 1.ポーリングの設定
// 2.SNMPから取得
// 3.ラインの設定
func GetVPanelPorts(n *datastore.NodeEnt) []*VPanelPortEnt {
	// ポーリングから取得
	if ports := getPortsFromPolling(n); len(ports) > 0 {
		return ports
	}
	// SNMPで取得
	if ports := getPortsBySNMP(n); len(ports) > 0 {
		return ports
	}
	// ラインから取得
	return getPortsFromLine(n)
}

func getPortsFromPolling(n *datastore.NodeEnt) []*VPanelPortEnt {
	ports := []*VPanelPortEnt{}
	traffPollings := make(map[string]*datastore.PollingEnt)
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == n.ID && p.Type == "snmp" {
			if p.Mode == "ifOperStatus" && strings.Contains(p.Script, ":") {
				a := strings.Split(p.Script, ":")
				if len(a) != 2 {
					return true
				}
				i, err := strconv.ParseInt(a[0], 10, 64)
				if err != nil {
					return true
				}
				state := "down"
				switch p.State {
				case "normal", "reapir":
					state = "up"
				case "unknown":
					state = "off"
				}
				ports = append(ports, &VPanelPortEnt{
					Index:        i,
					Name:         a[1],
					pollingIndex: p.Params,
					State:        state,
					Type:         6,
				})
			} else if p.Mode == "traffic" {
				traffPollings[p.Params] = p
			}
		}
		return true
	})
	for _, e := range ports {
		if p, ok := traffPollings[e.pollingIndex]; ok {
			e.InBytes = getTraffData("bytes", p)
			e.InPacktes = getTraffData("packets", p)
			e.InError = getTraffData("errors", p)
			e.OutBytes = getTraffData("outBytes", p)
			e.OutPacktes = getTraffData("outPackets", p)
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Index < ports[j].Index
	})
	return ports
}

func getTraffData(k string, p *datastore.PollingEnt) int64 {
	if d, ok := p.Result[k]; ok {
		if v, ok := d.(float64); ok {
			return int64(v)
		}
	}
	return 0
}

func getPortsFromLine(n *datastore.NodeEnt) []*VPanelPortEnt {
	ports := []*VPanelPortEnt{}
	max := int64(0)
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 != n.ID && l.NodeID2 != n.ID {
			return true
		}
		name := ""
		i := int64(0)
		a := strings.Split(l.Info, ":")
		if len(a) == 2 {
			i, _ = strconv.ParseInt(a[0], 10, 64)
			if max < i {
				max = i
			}
			name = a[1]
		} else if l.NodeID1 == n.ID {
			np := datastore.GetNode(l.NodeID2)
			if np == nil {
				return true
			}
			name = np.Name
		} else {
			np := datastore.GetNode(l.NodeID1)
			if np == nil {
				return true
			}
			name = np.Name
		}
		s := l.State1
		if l.NodeID1 == n.ID {
			s = l.State2
		}
		state := "down"
		switch s {
		case "normal", "reapir":
			state = "up"
		case "unknown":
			state = "off"
		}
		ports = append(ports, &VPanelPortEnt{
			Index: i,
			Name:  name,
			State: state,
			Type:  6,
		})
		return true
	})
	max++
	for _, e := range ports {
		if e.Index == 0 {
			e.Index = max
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Index == ports[j].Index {
			return strings.Compare(ports[i].Name, ports[j].Name) < 0
		}
		return ports[i].Index < ports[j].Index
	})
	for i, e := range ports {
		e.Index = int64(i + 1)
	}
	return ports
}

func getPortsBySNMP(n *datastore.NodeEnt) []*VPanelPortEnt {
	ports := []*VPanelPortEnt{}
	agent := &gosnmp.GoSNMP{
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(datastore.MapConf.Timeout) * time.Second,
		Retries:            datastore.MapConf.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	switch n.SnmpMode {
	case "v3auth":
		if n.User == "" || n.Password == "" {
			return ports
		}
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
		}
	case "v3authpriv":
		if n.User == "" || n.Password == "" {
			return ports
		}
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
		if n.User == "" || n.Password == "" {
			return ports
		}
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
	default:
		if n.Community == "" {
			return ports
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("getPortsBySNMP err=%v", err)
		return ports
	}
	defer agent.Conn.Close()
	ifMap := make(map[string]*VPanelPortEnt)
	_ = agent.Walk(datastore.MIBDB.NameToOID("ifTable"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) != 2 {
			return nil
		}
		e, ok := ifMap[a[1]]
		if !ok {
			ifMap[a[1]] = new(VPanelPortEnt)
			e = ifMap[a[1]]
		}
		switch a[0] {
		case "ifDescr":
			e.Name = getMIBStringVal(variable.Value)
		case "ifType":
			e.Type = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifSpeed":
			e.Speed = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifIndex":
			e.Index = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifPhysAddress":
			mac := getMIBStringVal(variable.Value)
			if len(mac) > 5 {
				e.MAC = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
			}
		case "ifAdminStatus":
			e.Admin = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifOperStatus":
			e.Oper = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifInOctets":
			e.InBytes = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifInUcastPkts", "ifInNUcastPkts", "ifInUnknownProtos":
			e.InPacktes += gosnmp.ToBigInt(variable.Value).Int64()
		case "ifInErrors":
			e.InError = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifOutOctets":
			e.OutBytes = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifOutUcastPkts", "ifOutNUcastPkts":
			e.OutPacktes += gosnmp.ToBigInt(variable.Value).Int64()
		case "ifOutErrors":
			e.OutError = gosnmp.ToBigInt(variable.Value).Int64()
		}
		return nil
	})
	for _, e := range ifMap {
		if e.Oper == 1 {
			e.State = "up"
		} else if e.Admin == 1 {
			e.State = "down"
		} else {
			e.State = "off"
		}
		ports = append(ports, e)
	}
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Index < ports[j].Index
	})
	return ports
}

func getMIBStringVal(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case []uint8:
		return string(v)
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return ""
}
