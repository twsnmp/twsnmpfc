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

type HrSystem struct {
	Index   int
	Name    string
	Value   string
	Polling string
}

type HrStorage struct {
	Index string
	Type  string
	Descr string
	Size  int64
	Used  int64
	Unit  int64
}

type HrDevice struct {
	Index  string
	Type   string
	Descr  string
	Status string
	Errors string
}

type HrFileSystem struct {
	Index    string
	Type     string
	Mount    string
	Remote   string
	Bootable int64
	Access   int64
}

type HrProcess struct {
	PID    string
	Name   string
	Type   string
	Status string
	Path   string
	Param  string
	CPU    int64
	Mem    int64
}

type HostResourceEnt struct {
	System     []*HrSystem
	Storage    []*HrStorage
	Device     []*HrDevice
	FileSystem []*HrFileSystem
	Process    []*HrProcess
}

func GetHostResource(n *datastore.NodeEnt) *HostResourceEnt {
	hr := new(HostResourceEnt)
	hr.System = []*HrSystem{}
	hr.Storage = []*HrStorage{}
	hr.Device = []*HrDevice{}
	hr.FileSystem = []*HrFileSystem{}
	hr.Process = []*HrProcess{}
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
	storageMap := make(map[string]*HrStorage)
	deviceMap := make(map[string]*HrDevice)
	fsMap := make(map[string]*HrFileSystem)
	procMap := make(map[string]*HrProcess)
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
				Name:    "システム時刻",
				Value:   getDateAndTime(variable.Value),
				Index:   2,
				Polling: "hrSystemDate",
			})
		case "hrSystemInitialLoadDevice":
			hr.System = append(hr.System, &HrSystem{
				Name:  "起動デバイス",
				Value: getMIBStringVal(variable.Value),
				Index: 3,
			})
		case "hrSystemInitialLoadParameters":
			hr.System = append(hr.System, &HrSystem{
				Name:  "起動パラメータ",
				Value: getMIBStringVal(variable.Value),
				Index: 4,
			})
		case "hrSystemNumUsers":
			hr.System = append(hr.System, &HrSystem{
				Name:    "システムユーザ数",
				Value:   fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index:   5,
				Polling: "hrSystemNumUsers",
			})
		case "hrSystemProcesses":
			hr.System = append(hr.System, &HrSystem{
				Name:    "システムプロセス数",
				Value:   fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index:   6,
				Polling: "hrSystemProcesses",
			})
		case "hrSystemMaxProcesses":
			hr.System = append(hr.System, &HrSystem{
				Name:  "最大プロセス数",
				Value: fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index: 7,
			})
		case "hrMemorySize":
			hr.System = append(hr.System, &HrSystem{
				Name:  "メモリサイズ",
				Value: fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index: 8,
			})
		case "hrProcessorLoad":
			hr.System = append(hr.System, &HrSystem{
				Name:    fmt.Sprintf("CPU%d使用率", nCPU),
				Value:   fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64()),
				Index:   8 + nCPU,
				Polling: "hrProcessorLoad." + a[1],
			})
			nCPU++
		case "hrStorageIndex":
			// Skip
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
		case "hrStorageAllocationUnits":
			if s, ok := storageMap[a[1]]; !ok {
				storageMap[a[1]] = &HrStorage{
					Unit: gosnmp.ToBigInt(variable.Value).Int64(),
				}
			} else {
				s.Unit = gosnmp.ToBigInt(variable.Value).Int64()
			}
		case "hrDeviceIndex":
			// Skip
		case "hrDeviceType":
			if d, ok := deviceMap[a[1]]; !ok {
				deviceMap[a[1]] = &HrDevice{
					Type: datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value)),
				}
			} else {
				d.Type = datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
			}
		case "hrDeviceDescr":
			if d, ok := deviceMap[a[1]]; !ok {
				deviceMap[a[1]] = &HrDevice{
					Descr: getMIBStringVal(variable.Value),
				}
			} else {
				d.Descr = getMIBStringVal(variable.Value)
			}
		case "hrDeviceID", "hrProcessorFrwID", "hrNetworkIfIndex", "hrDiskStorageAccess", "hrDiskStorageMedia":
		case "hrDiskStorageRemoveble", "hrDiskStorageCapacity", "hrPartitionIndex", "hrPartitionLabel", "hrPartitionID":
		case "hrPartitionSize", "hrPartitionFSIndex":
			// Skip
		case "hrDeviceStatus":
			if d, ok := deviceMap[a[1]]; !ok {
				deviceMap[a[1]] = &HrDevice{
					Status: getDeviceStatusName(gosnmp.ToBigInt(variable.Value).Int64()),
				}
			} else {
				d.Status = getDeviceStatusName(gosnmp.ToBigInt(variable.Value).Int64())
			}
		case "hrDeviceErrors":
			if d, ok := deviceMap[a[1]]; !ok {
				deviceMap[a[1]] = &HrDevice{
					Errors: getMIBStringVal(variable.Value),
				}
			} else {
				d.Errors = getMIBStringVal(variable.Value)
			}
		case "hrFSIndex":
			// Skip
		case "hrFSMountPoint":
			if f, ok := fsMap[a[1]]; !ok {
				fsMap[a[1]] = &HrFileSystem{
					Mount: getMIBStringVal(variable.Value),
				}
			} else {
				f.Mount = getMIBStringVal(variable.Value)
			}
		case "hrFSRemoteMountPoint":
			if f, ok := fsMap[a[1]]; !ok {
				fsMap[a[1]] = &HrFileSystem{
					Remote: getMIBStringVal(variable.Value),
				}
			} else {
				f.Remote = getMIBStringVal(variable.Value)
			}
		case "hrFSType":
			if f, ok := fsMap[a[1]]; !ok {
				fsMap[a[1]] = &HrFileSystem{
					Type: datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value)),
				}
			} else {
				f.Type = datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
			}
		case "hrFSAccess":
			if f, ok := fsMap[a[1]]; !ok {
				fsMap[a[1]] = &HrFileSystem{
					Access: gosnmp.ToBigInt(variable.Value).Int64(),
				}
			} else {
				f.Access = gosnmp.ToBigInt(variable.Value).Int64()
			}
		case "hrFSBootable":
			if f, ok := fsMap[a[1]]; !ok {
				fsMap[a[1]] = &HrFileSystem{
					Bootable: gosnmp.ToBigInt(variable.Value).Int64(),
				}
			} else {
				f.Bootable = gosnmp.ToBigInt(variable.Value).Int64()
			}
		case "hrFSLastFullBackupDate", "hrFSLastPartialBackupDate":
		case "hrSWRunIndex", "hrSWRunID":
			// Skip
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
		case "hrSWRunPath":
			if p, ok := procMap[a[1]]; !ok {
				procMap[a[1]] = &HrProcess{
					Path: getMIBStringVal(variable.Value),
				}
			} else {
				p.Path = getMIBStringVal(variable.Value)
			}
		case "hrSWRunParameters":
			if p, ok := procMap[a[1]]; !ok {
				procMap[a[1]] = &HrProcess{
					Param: getMIBStringVal(variable.Value),
				}
			} else {
				p.Param = getMIBStringVal(variable.Value)
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
		default:
			log.Printf("%v", a)
		}
		return nil
	})
	for i, s := range storageMap {
		s.Index = i
		if s.Unit > 0 {
			s.Size *= s.Unit
			s.Used *= s.Unit
		}
		hr.Storage = append(hr.Storage, s)
	}
	for pid, p := range procMap {
		p.PID = pid
		hr.Process = append(hr.Process, p)
	}
	for i, d := range deviceMap {
		d.Index = i
		hr.Device = append(hr.Device, d)
	}
	for i, f := range fsMap {
		f.Index = i
		hr.FileSystem = append(hr.FileSystem, f)
	}
	return hr
}

func getDeviceStatusName(s int64) string {
	switch s {
	case 2:
		return "Running"
	case 3:
		return "Warning"
	case 4:
		return "Testing"
	case 5:
		return "Down"
	}
	return "Unknown"
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
			return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d%c%02d%02d",
				(int(v[0])*256 + int(v[1])), v[2], v[3], v[4], v[5], v[6], v[8], v[9], v[10])
		}
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return "Unknown"
}
