package backend

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
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
		case "hrStorageAllocationFailures", "hrPrinterStatus", "hrPrinterDetectedErrorState", "hrFSStorageIndex":
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
		case "hrSWRunIndex", "hrSWRunID", "hrSWInstalledDate", "hrSWInstalledType", "hrSWInstalledName", "hrSWInstalledID":
		case "hrSWOSIndex", "hrSWInstalledLastChange", "hrSWInstalledLastUpdateTime", "hrSWInstalledIndex":
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
		if len(v) == 11 {
			return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d.%02d%c%02d%02d",
				(int(v[0])*256 + int(v[1])), v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10])
		} else if len(v) == 8 {
			return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d.%02d",
				(int(v[0])*256 + int(v[1])), v[2], v[3], v[4], v[5], v[6], v[7])
		}
		log.Printf("invalid  date and time format %v", v)
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return "Unknown"
}

type PortEnt struct {
	Port    int
	Address string
	Process string
	Descr   string
}

func GetPortList(n *datastore.NodeEnt) ([]*PortEnt, []*PortEnt) {
	tcpPorts := []*PortEnt{}
	udpPorts := []*PortEnt{}
	processNameMap := make(map[int]string)
	agent := getSNMPAgent(n)
	if agent == nil {
		return tcpPorts, udpPorts
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("GetPortList err=%v", err)
		return tcpPorts, udpPorts
	}
	defer agent.Conn.Close()
	_ = agent.Walk(datastore.MIBDB.NameToOID("hrSWRunName"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) != 2 {
			return nil
		}
		processNameMap[getInt(a[1])] = getMIBStringVal(variable.Value)
		return nil
	})
	_ = agent.Walk(datastore.MIBDB.NameToOID("udpEndpointProcess"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) < 2 {
			return nil
		}
		addr := getLocalAddr(a)
		if addr == "" {
			return nil
		}
		var port int
		if strings.Contains(addr, ":") {
			port = getInt(a[19])
		} else {
			port = getInt(a[7])
		}
		rport := getInt(a[len(a)-2])
		descr, ok := datastore.GetServiceName(17, port)
		if !ok && rport > 0 {
			raddr := getRemoteAddr(a)
			sv, ok := datastore.GetServiceName(17, rport)
			if ok {
				descr = fmt.Sprintf("%s -> %s", sv, raddr)
			}
		}
		pid := int(gosnmp.ToBigInt(variable.Value).Int64())
		process, ok := processNameMap[pid]
		if !ok {
			process = fmt.Sprintf("%d", pid)
		}
		udpPorts = append(udpPorts, &PortEnt{
			Port:    port,
			Descr:   descr,
			Address: addr,
			Process: process,
		})
		return nil
	})
	_ = agent.Walk(datastore.MIBDB.NameToOID("tcpListenerProcess"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) < 2 {
			return nil
		}
		port := getInt(a[len(a)-1])
		descr, _ := datastore.GetServiceName(6, port)
		pid := int(gosnmp.ToBigInt(variable.Value).Int64())
		process, ok := processNameMap[pid]
		if !ok {
			process = fmt.Sprintf("%d", pid)
		}
		tcpPorts = append(tcpPorts, &PortEnt{
			Port:    port,
			Descr:   descr,
			Address: getLocalAddr(a),
			Process: process,
		})
		return nil
	})
	return tcpPorts, udpPorts
}

func getInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}

func getLocalAddr(a []string) string {
	switch a[1] {
	case "1":
		// IPv4
		return strings.Join(a[3:4+3], ".")
	case "2":
		// IPv6
		r := ""
		for i := 3; i < 16+3; i++ {
			if i != 3 {
				r += ":"
			}
			r += fmt.Sprintf("%02x", getInt(a[i]))
		}
		return r
	}
	return ""
}

func getRemoteAddr(a []string) string {
	switch a[1] {
	case "1":
		// IPv4
		return strings.Join(a[10:4+10], ".")
	case "2":
		// IPv6
		r := ""
		for i := 22; i < 16+22; i++ {
			if i != 22 {
				r += ":"
			}
			r += fmt.Sprintf("%02x", getInt(a[i]))
		}
		return r
	}
	return ""
}

type RMONEnt struct {
	ProtocolDir map[int]string
	SysUptime   int64
	LocalTime   int64
	MIBs        map[string]map[string]string
}

var protocolDirCache = make(map[string]map[int]string)

func getRMONProtocolDir(id string, agent *gosnmp.GoSNMP) map[int]string {
	if e, ok := protocolDirCache[id]; ok {
		return e
	}
	ret := make(map[int]string)
	indexMap := make(map[string]int)
	_ = agent.Walk(datastore.MIBDB.NameToOID("protocolDirLocalIndex"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		indexMap[a[1]] = int(gosnmp.ToBigInt(variable.Value).Int64())
		return nil
	})
	_ = agent.Walk(datastore.MIBDB.NameToOID("protocolDirDescr"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		if i, ok := indexMap[a[1]]; ok {
			ret[i] = getMIBStringVal(variable.Value)
		}
		return nil
	})
	protocolDirCache[id] = ret
	return ret
}

func getSysUptime(agent *gosnmp.GoSNMP) int64 {
	oids := []string{datastore.MIBDB.NameToOID("sysUpTime.0")}
	result, err := agent.Get(oids)
	if err != nil {
		return 0
	}
	var uptime int64
	for _, variable := range result.Variables {
		if variable.Name == datastore.MIBDB.NameToOID("sysUpTime.0") {
			uptime = int64(gosnmp.ToBigInt(variable.Value).Uint64())
			break
		}
	}
	return uptime
}

func GetRMON(n *datastore.NodeEnt, t string) *RMONEnt {
	ret := &RMONEnt{
		ProtocolDir: make(map[int]string),
		MIBs:        make(map[string]map[string]string),
	}
	agent := getSNMPAgent(n)
	if agent == nil {
		log.Printf("GetRMON agent == nil")
		return ret
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("GetRMON err=%v", err)
		return ret
	}
	defer agent.Conn.Close()
	if t == "protocolDistStatsTable" ||
		t == "alHostTable" ||
		t == "alMatrixSDTable" {
		ret.ProtocolDir = getRMONProtocolDir(n.ID, agent)
	}
	ret.SysUptime = getSysUptime(agent)
	ret.LocalTime = time.Now().Unix()
	_ = agent.Walk(datastore.MIBDB.NameToOID(t), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		name := a[0]
		index := a[1]
		value := ""
		vendor := ""
		switch variable.Type {
		case gosnmp.OctetString:
			mi := datastore.FindMIBInfo(name)
			if mi != nil {
				switch mi.Type {
				case "PhysAddress", "OctetString":
					a, ok := variable.Value.([]uint8)
					if !ok {
						a = []uint8(getMIBStringVal(variable.Value))
					}
					mac := []string{}
					for _, m := range a {
						mac = append(mac, fmt.Sprintf("%02X", m&0x00ff))
					}
					value = strings.Join(mac, ":")
					vendor = datastore.FindVendor(value)
				case "BITS":
					a, ok := variable.Value.([]uint8)
					if !ok {
						a = []uint8(getMIBStringVal(variable.Value))
					}
					hex := []string{}
					ap := []string{}
					for _, m := range a {
						hex = append(hex, fmt.Sprintf("%02X", m&0x00ff))
					}
					value = strings.Join(hex, " ")
					if len(ap) > 0 {
						value += " " + strings.Join(ap, " ")
					}
				case "DisplayString":
					value = getMIBStringVal(variable.Value)
				default:
					value = getMIBStringVal(variable.Value)
				}
			} else {
				value = getMIBStringVal(variable.Value)
			}
		case gosnmp.ObjectIdentifier:
			value = datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
		case gosnmp.TimeTicks:
			t := gosnmp.ToBigInt(variable.Value).Uint64()
			value = fmt.Sprintf("%d", t)
		default:
			v := int(gosnmp.ToBigInt(variable.Value).Uint64())
			value = fmt.Sprintf("%d", v)
		}
		if _, ok := ret.MIBs[a[1]]; !ok {
			ret.MIBs[index] = make(map[string]string)
		}
		ret.MIBs[index][name] = value
		if vendor != "" {
			ret.MIBs[index]["Vendor"] = vendor
		}
		return nil
	})
	return ret
}
