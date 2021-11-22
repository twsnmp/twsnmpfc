package backend

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/twsnmp/gosnmp"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type HrProcess struct {
	PID    string
	Name   string
	Type   string
	Status string
	CPU    int64
	Mem    int64
}

type HrStorage struct {
	Type  string
	Descr string
	Size  int64
	Used  int64
}

type HrSystem struct {
	Index int
	Name  string
	Value string
}

type HostResourceEnt struct {
	System  []*HrSystem
	Process []*HrProcess
	Storage []*HrStorage
}

func GetHostResource(n *datastore.NodeEnt) *HostResourceEnt {
	hr := new(HostResourceEnt)
	hr.System = []*HrSystem{}
	hr.Process = []*HrProcess{}
	hr.Storage = []*HrStorage{}
	agent := getSNMPAgent(n)
	if agent == nil {
		return hr
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("getPortsBySNMP err=%v", err)
		return hr
	}
	defer agent.Conn.Close()
	nCPU := 1
	procMap := make(map[string]*HrProcess)
	storageMap := make(map[string]*HrStorage)
	_ = agent.Walk(datastore.MIBDB.NameToOID("host"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		switch a[0] {
		case "hrSystemUptime":
			hr.System = append(hr.System, &HrSystem{
				Name:  "システム稼働時間",
				Value: getTimeTickStr(gosnmp.ToBigInt(variable.Value).Int64()),
				Index: 1,
			})
		case "hrSystemDate":
			hr.System = append(hr.System, &HrSystem{
				Name:  "システム稼働時間",
				Value: getDateAndTime(variable.Value),
				Index: 2,
			})
		case "hrSystemNumUsers":
			hr.System = append(hr.System, &HrSystem{
				Name:  "システムユーザ数",
				Value: fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index: 3,
			})
		case "hrSystemProcesses":
			hr.System = append(hr.System, &HrSystem{
				Name:  "システムプロセス数",
				Value: fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index: 4,
			})
		case "hrMemorySize":
			hr.System = append(hr.System, &HrSystem{
				Name:  "メモリサイズ",
				Value: fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index: 5,
			})
		case "hrProcessorLoad":
			hr.System = append(hr.System, &HrSystem{
				Name:  fmt.Sprintf("CPU%d使用率", nCPU),
				Value: fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index: 5 + nCPU,
			})
			nCPU++
		case "hrStorageType":
			if s, ok := storageMap[a[1]]; !ok {
				storageMap[a[1]] = &HrStorage{
					Type: datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value)),
				}
			} else {
				s.Type = datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
			}
		case "hrStorageDescr":
			if s, ok := storageMap[a[1]]; !ok {
				storageMap[a[1]] = &HrStorage{
					Descr: getMIBStringVal(variable.Value),
				}
			} else {
				s.Descr = getMIBStringVal(variable.Value)
			}
		case "hrStorageSize":
			if s, ok := storageMap[a[1]]; !ok {
				storageMap[a[1]] = &HrStorage{
					Size: gosnmp.ToBigInt(variable.Value).Int64(),
				}
			} else {
				s.Size = gosnmp.ToBigInt(variable.Value).Int64()
			}
		case "hrStorageUsed":
			if s, ok := storageMap[a[1]]; !ok {
				storageMap[a[1]] = &HrStorage{
					Used: gosnmp.ToBigInt(variable.Value).Int64(),
				}
			} else {
				s.Used = gosnmp.ToBigInt(variable.Value).Int64()
			}
		case "hrSWRunName":
			if p, ok := procMap[a[1]]; !ok {
				procMap[a[1]] = &HrProcess{
					Name: getMIBStringVal(variable.Value),
				}
			} else {
				p.Name = getMIBStringVal(variable.Value)
			}
		case "hrSWRunType":
			if p, ok := procMap[a[1]]; !ok {
				procMap[a[1]] = &HrProcess{
					Type: getSWRunTypeName(gosnmp.ToBigInt(variable.Value).Int64()),
				}
			} else {
				p.Type = getSWRunTypeName(gosnmp.ToBigInt(variable.Value).Int64())
			}
		case "hrSWRunStatus":
			if p, ok := procMap[a[1]]; !ok {
				procMap[a[1]] = &HrProcess{
					Status: getSWRunStatusName(gosnmp.ToBigInt(variable.Value).Int64()),
				}
			} else {
				p.Status = getSWRunStatusName(gosnmp.ToBigInt(variable.Value).Int64())
			}
		case "hrSWRunPerfCPU":
			if p, ok := procMap[a[1]]; !ok {
				procMap[a[1]] = &HrProcess{
					CPU: gosnmp.ToBigInt(variable.Value).Int64(),
				}
			} else {
				p.CPU = gosnmp.ToBigInt(variable.Value).Int64()
			}
		case "hrSWRunPerfMem":
			if p, ok := procMap[a[1]]; !ok {
				procMap[a[1]] = &HrProcess{
					Mem: gosnmp.ToBigInt(variable.Value).Int64(),
				}
			} else {
				p.Mem = gosnmp.ToBigInt(variable.Value).Int64()
			}
		}
		return nil
	})
	for _, s := range storageMap {
		hr.Storage = append(hr.Storage, s)
	}
	for pid, p := range procMap {
		p.PID = pid
		hr.Process = append(hr.Process, p)
	}
	return hr
}

func getSWRunStatusName(s int64) string {
	switch s {
	case 1:
		return "Running"
	case 2:
		return "Runnable"
	case 3:
		return "NotRunnable"
	case 4:
		return "Invalid"
	}
	return "Unknown"
}

func getSWRunTypeName(t int64) string {
	switch t {
	case 2:
		return "OS"
	case 3:
		return "Driver"
	case 4:
		return "Application"
	}
	return "Unknown"
}

func getPortsBySNMP(n *datastore.NodeEnt) []*VPanelPortEnt {
	ports := []*VPanelPortEnt{}
	agent := getSNMPAgent(n)
	if agent == nil {
		return ports
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
	// ifXTableからも取得する
	ifXTable := false
	_ = agent.Walk(datastore.MIBDB.NameToOID("ifXTable"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) != 2 {
			return nil
		}
		e, ok := ifMap[a[1]]
		if !ok {
			return nil
		}
		if !ifXTable {
			// Reset Counter
			for _, e := range ifMap {
				e.InBytes = 0
				e.InPacktes = 0
				e.OutBytes = 0
				e.OutPacktes = 0
			}
			ifXTable = true
		}
		switch a[0] {
		case "ifName":
			e.Name = getMIBStringVal(variable.Value)
		case "ifHighSpeed":
			e.Speed = gosnmp.ToBigInt(variable.Value).Int64() * 1000 * 1000
		case "ifHCInOctets":
			e.InBytes = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifHCInMulticastPkts", "ifHCInBroadcastPkts", "ifHCInUcastPkts":
			e.InPacktes += gosnmp.ToBigInt(variable.Value).Int64()
		case "ifHCOutOctets":
			e.OutBytes = gosnmp.ToBigInt(variable.Value).Int64()
		case "ifHCOutUcastPkts", "ifHCOutMulticastPkts", "ifHCOutBroadcastPkts":
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

func getSNMPAgent(n *datastore.NodeEnt) *gosnmp.GoSNMP {
	if strings.HasPrefix(n.SnmpMode, "v3") && n.User == "" {
		return nil
	} else if n.Community == "" {
		return nil
	}
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

func getTimeTickStr(t int64) string {
	ft := float64(t) / 100
	if ft > 3600*24 {
		return fmt.Sprintf("%.2f日(%d)", ft/(3600*24), t)
	} else if ft > 3600 {
		return fmt.Sprintf("%.2f時間(%d)", ft/(3600), t)
	}
	return fmt.Sprintf("%.2f秒(%d)", ft, t)
}

func getDateAndTime(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case []uint8:
		if len(v) > 6 {
			return fmt.Sprintf("%d/%d/%d %d:%d:%d",
				(int(v[0])*256 + int(v[1])), v[2], v[3], v[4], v[5], v[6])
		}
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return "Unknown"
}
