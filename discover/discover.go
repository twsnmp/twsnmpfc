// Package discover : 自動発見
package discover

/* discover.go: 自動発見の処理
自動発見は、PINGを実行して、応答があるノードに関してSNMPの応答があるか確認する
*/

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/signalsciences/ipv4"
	"github.com/twsnmp/gosnmp"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
	"github.com/twsnmp/twsnmpfc/polling"
)

// GRID : 自動発見時にノードを配置する間隔
const GRID = 90

var (
	Stat DiscoverStat
	Stop bool
	X    int
	Y    int
)

type DiscoverStat struct {
	Running   bool
	Total     uint32
	Sent      uint32
	Found     uint32
	Snmp      uint32
	Web       uint32
	Mail      uint32
	SSH       uint32
	File      uint32
	RDP       uint32
	LDAP      uint32
	StartTime int64
	Now       int64
}

type discoverInfoEnt struct {
	IP          string
	HostName    string
	SysName     string
	SysObjectID string
	IfIndexList []string
	ServerList  map[string]bool
	X           int
	Y           int
}

// StopDiscover : 自動発見を停止する
func StopDiscover() {
	for Stat.Running {
		Stop = true
		time.Sleep(time.Millisecond * 100)
	}
}

func StartDiscover() error {
	if Stat.Running {
		return fmt.Errorf("discover already runnning")
	}
	sip, err := ipv4.FromDots(datastore.DiscoverConf.StartIP)
	if err != nil {
		return fmt.Errorf("discover start ip err=%v", err)
	}
	eip, err := ipv4.FromDots(datastore.DiscoverConf.EndIP)
	if err != nil {
		return fmt.Errorf("discover end ip err=%v", err)
	}
	if sip > eip {
		return fmt.Errorf("discover start ip > end ip")
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf("自動発見開始 %s - %s", datastore.DiscoverConf.StartIP, datastore.DiscoverConf.EndIP),
	})
	Stop = false
	Stat.Total = eip - sip + 1
	Stat.Sent = 0
	Stat.Found = 0
	Stat.Snmp = 0
	Stat.Web = 0
	Stat.Mail = 0
	Stat.SSH = 0
	Stat.File = 0
	Stat.RDP = 0
	Stat.Running = true
	Stat.StartTime = time.Now().Unix()
	Stat.Now = Stat.StartTime
	X = (1 + datastore.DiscoverConf.X/GRID) * GRID
	Y = (1 + datastore.DiscoverConf.Y/GRID) * GRID
	var mu sync.Mutex
	sem := make(chan bool, 20)
	go func() {
		for ; sip <= eip && !Stop; sip++ {
			sem <- true
			Stat.Sent++
			Stat.Now = time.Now().Unix()
			go func(ip uint32) {
				defer func() {
					<-sem
				}()
				ipstr := ipv4.ToDots(ip)
				if datastore.FindNodeFromIP(ipstr) != nil {
					return
				}
				r := ping.DoPing(ipstr, datastore.DiscoverConf.Timeout, datastore.DiscoverConf.Retry, 64)
				if r.Stat == ping.PingOK {
					dent := discoverInfoEnt{
						IP:          ipstr,
						IfIndexList: []string{},
						ServerList:  make(map[string]bool),
					}
					r := &net.Resolver{}
					ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*50)
					defer cancel()
					if names, err := r.LookupAddr(ctx, ipstr); err == nil && len(names) > 0 {
						dent.HostName = names[0]
					}
					getSnmpInfo(ipstr, &dent)
					checkServer(&dent)
					mu.Lock()
					dent.X = X
					dent.Y = Y
					Stat.Found++
					X += GRID
					if X > GRID*10 {
						X = GRID
						Y += GRID
					}
					if dent.SysName != "" {
						Stat.Snmp++
					}
					if dent.ServerList["http"] || dent.ServerList["https"] {
						Stat.Web++
					}
					if dent.ServerList["cifs"] || dent.ServerList["nfs"] {
						Stat.File++
					}
					if dent.ServerList["rdp"] || dent.ServerList["vnc"] {
						Stat.RDP++
					}
					if dent.ServerList["ldap"] || dent.ServerList["ldaps"] || dent.ServerList["kerberos"] {
						Stat.LDAP++
					}
					if dent.ServerList["smtp"] || dent.ServerList["imap"] || dent.ServerList["pop3"] {
						Stat.Mail++
					}
					if dent.ServerList["ssh"] {
						Stat.SSH++
					}
					addFoundNode(&dent)
					mu.Unlock()
				}
			}(sip)
		}
		for len(sem) > 0 {
			time.Sleep(time.Millisecond * 10)
		}
		Stat.Running = false
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("自動発見終了 %s - %s", datastore.DiscoverConf.StartIP, datastore.DiscoverConf.EndIP),
		})
	}()
	return nil
}

func ClearStat() {
	if Stat.Running {
		return
	}
	Stat.Total = 0
	Stat.Sent = 0
	Stat.Found = 0
	Stat.Snmp = 0
	Stat.Web = 0
	Stat.Mail = 0
	Stat.SSH = 0
	Stat.File = 0
	Stat.RDP = 0
	Stat.StartTime = 0
	Stat.Now = 0
}

func getSnmpInfo(t string, dent *discoverInfoEnt) {
	agent := &gosnmp.GoSNMP{
		Target:             t,
		Port:               161,
		Transport:          "udp",
		Community:          datastore.MapConf.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(datastore.DiscoverConf.Timeout) * time.Second,
		Retries:            datastore.DiscoverConf.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	switch datastore.MapConf.SnmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	defer agent.Conn.Close()
	oids := []string{datastore.MIBDB.NameToOID("sysName"), datastore.MIBDB.NameToOID("sysObjectID")}
	result, err := agent.GetNext(oids)
	if err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	for _, variable := range result.Variables {
		if datastore.MIBDB.OIDToName(variable.Name) == "sysName.0" {
			dent.SysName = getMIBStringVal(variable.Value)
		} else if datastore.MIBDB.OIDToName(variable.Name) == "sysObjectID.0" {
			dent.SysObjectID = getMIBStringVal(variable.Value)
		}
	}
	_ = agent.Walk(datastore.MIBDB.NameToOID("ifType"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) == 2 &&
			a[0] == "ifType" &&
			gosnmp.ToBigInt(variable.Value).Int64() == 6 {
			dent.IfIndexList = append(dent.IfIndexList, a[1])
		}
		return nil
	})
}

func addFoundNode(dent *discoverInfoEnt) {
	n := datastore.NodeEnt{
		Name:  dent.HostName,
		IP:    dent.IP,
		Icon:  "desktop",
		X:     dent.X,
		Y:     dent.Y,
		Descr: time.Now().Format("2006/01/02") + "に発見",
	}
	if n.Name == "" {
		if dent.SysName != "" {
			n.Name = dent.SysName
		} else {
			n.Name = dent.IP
		}
	}
	if dent.SysObjectID != "" {
		n.SnmpMode = datastore.MapConf.SnmpMode
		n.User = datastore.MapConf.SnmpUser
		n.Password = datastore.MapConf.SnmpPassword
		n.Community = datastore.MapConf.Community
		n.Icon = "hdd"
		n.Descr += " / snmp対応"
	}
	if len(dent.ServerList) > 0 {
		n.Descr += " / "
		sl := []string{}
		for s := range dent.ServerList {
			sl = append(sl, s)
		}
		n.Descr += strings.Join(sl, ",") + "対応"
	}
	if err := datastore.AddNode(&n); err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "discover",
		Level:    "info",
		NodeID:   n.ID,
		NodeName: n.Name,
		Event:    "自動発見により追加",
	})
	if len(datastore.DiscoverConf.AutoAddPollings) < 1 {
		return
	}
	if datastore.DiscoverConf.AutoAddPollings[0] == "basic" {
		addBasicPolling(dent, &n)
		return
	}
	autoAddPollings(&n)
}

func autoAddPollings(n *datastore.NodeEnt) {
	for _, id := range datastore.DiscoverConf.AutoAddPollings {
		pt := datastore.GetPollingTemplate(id)
		if pt == nil {
			log.Printf("add polling template not found id=%s", id)
			continue
		}
		if pt.AutoMode == "disable" {
			continue
		}
		if pt.AutoMode != "" {
			// インデックスの展開などを行う並列で処理する
			go polling.AutoAddPolling(n, pt)
			continue
		}
		p := new(datastore.PollingEnt)
		p.Name = pt.Name
		p.NodeID = n.ID
		p.Type = pt.Type
		p.Params = pt.Params
		p.Mode = pt.Mode
		p.Script = pt.Script
		p.Extractor = pt.Extractor
		p.Filter = pt.Filter
		p.Level = pt.Level
		p.PollInt = datastore.MapConf.PollInt
		p.Timeout = datastore.MapConf.Timeout
		p.Retry = datastore.MapConf.Timeout
		p.LogMode = 0
		p.NextTime = 0
		p.State = "unknown"
		if err := datastore.AddPolling(p); err != nil {
			log.Printf("discover err=%v", err)
			return
		}
	}
}

func addBasicPolling(dent *discoverInfoEnt, n *datastore.NodeEnt) {
	p := &datastore.PollingEnt{
		NodeID:  n.ID,
		Name:    "PING監視",
		Type:    "ping",
		Level:   "low",
		State:   "unknown",
		PollInt: datastore.MapConf.PollInt,
		Timeout: datastore.MapConf.Timeout,
		Retry:   datastore.MapConf.Retry,
	}
	if err := datastore.AddPolling(p); err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	for s := range dent.ServerList {
		name := ""
		ptype := ""
		params := ""
		mode := ""
		level := "off"
		switch s {
		case "http":
			name = "HTTPサーバー監視"
			ptype = "http"
			params = "http://" + n.IP
		case "https":
			name = "HTTPSサーバー監視"
			ptype = "http"
			mode = "https"
			params = "https://" + n.IP
			level = "low"
		case "smtp":
			name = "SMTPサーバー監視"
			ptype = "tcp"
			params = "25"
			level = "low"
		case "pop3":
			name = "POP3サーバー監視"
			ptype = "tcp"
			params = "110"
		case "imap":
			name = "IMAPサーバー監視"
			ptype = "tcp"
			params = "143"
			level = "low"
		case "ssh":
			name = "SSHサーバー監視"
			ptype = "tcp"
			params = "22"
		case "cifs":
			name = "ファイル共有(CIFS)サーバー監視"
			ptype = "tcp"
			params = "445"
		case "nfs":
			name = "ファイル共有(NFS)サーバー監視"
			ptype = "tcp"
			params = "2049"
		case "vnc":
			name = "画面共有(VNC)サーバー監視"
			ptype = "tcp"
			params = "5900"
		case "rdp":
			name = "RDPサーバー監視"
			ptype = "tcp"
			params = "3389"
		case "kerberos":
			name = "AD(kerberos)サーバー監視"
			ptype = "tcp"
			params = "88"
		case "ldap":
			name = "LDAPサーバー監視"
			ptype = "tcp"
			params = "389"
		case "ldaps":
			name = "LDAPSサーバー監視"
			ptype = "tcp"
			params = "88"
		default:
			continue
		}
		p = &datastore.PollingEnt{
			NodeID:  n.ID,
			Name:    name,
			Type:    ptype,
			Mode:    mode,
			Params:  params,
			Level:   level,
			State:   "unknown",
			PollInt: datastore.MapConf.PollInt,
			Timeout: datastore.MapConf.Timeout,
			Retry:   datastore.MapConf.Retry,
		}
		if err := datastore.AddPolling(p); err != nil {
			log.Printf("discover err=%v", err)
			return
		}
	}
	if dent.SysObjectID == "" {
		return
	}
	p = &datastore.PollingEnt{
		NodeID:  n.ID,
		Name:    "sysUptime監視",
		Type:    "snmp",
		Mode:    "sysUpTime",
		Level:   "off",
		State:   "unknown",
		PollInt: datastore.MapConf.PollInt,
		Timeout: datastore.MapConf.Timeout,
		Retry:   datastore.MapConf.Retry,
	}
	if err := datastore.AddPolling(p); err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	for _, i := range dent.IfIndexList {
		p = &datastore.PollingEnt{
			NodeID:  n.ID,
			Type:    "snmp",
			Name:    "IF " + i + "監視",
			Mode:    "ifOperStatus",
			Params:  i,
			Level:   "off",
			State:   "unknown",
			PollInt: datastore.MapConf.PollInt,
			Timeout: datastore.MapConf.Timeout,
			Retry:   datastore.MapConf.Retry,
		}
		if err := datastore.AddPolling(p); err != nil {
			log.Printf("discover err=%v", err)
			return
		}
	}
}

// サーバーの確認
func checkServer(dent *discoverInfoEnt) {
	checkList := map[string]string{
		"http":     "80",
		"https":    "443",
		"pop3":     "110",
		"imap":     "143",
		"smtp":     "25",
		"ssh":      "22",
		"cifs":     "445",
		"nfs":      "2049",
		"vnc":      "5900",
		"rdp":      "3389",
		"ldap":     "389",
		"ldaps":    "636",
		"kerberos": "88",
	}
	for s, p := range checkList {
		if doTCPConnect(dent.IP + ":" + p) {
			dent.ServerList[s] = true
		}
	}
}

func doTCPConnect(dst string) bool {
	conn, err := net.DialTimeout("tcp", dst, time.Duration(datastore.DiscoverConf.Timeout)*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
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
