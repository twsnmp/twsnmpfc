// Package discover : 自動発見
package discover

/* discover.go: 自動発見の処理
自動発見は、PINGを実行して、応答があるノードに関してSNMPの応答があるか確認する
*/

import (
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
)

// GRID : 自動発見時にノードを配置する間隔
const GRID = 90

// Discover : 自動発見の
type Discover struct {
	ds   *datastore.DataStore
	ping *ping.Ping
	Stat DiscoverStat
	Stop bool
	X    int
	Y    int
}

type DiscoverStat struct {
	Running   bool
	Total     uint32
	Sent      uint32
	Found     uint32
	Snmp      uint32
	Progress  uint32
	StartTime int64
	EndTime   int64
}

type discoverInfoEnt struct {
	IP          string
	HostName    string
	SysName     string
	SysObjectID string
	IfIndexList []string
	X           int
	Y           int
}

func NewDiscover(ds *datastore.DataStore, ping *ping.Ping) *Discover {
	d := &Discover{
		ds:   ds,
		ping: ping,
	}
	return d
}

// StopDiscover : 自動発見を停止する
func (d *Discover) StopDiscover() {
	for d.Stat.Running {
		d.Stop = true
		time.Sleep(time.Millisecond * 100)
	}
}

func (d *Discover) StartDiscover() error {
	if d.Stat.Running {
		return fmt.Errorf("discover already runnning")
	}
	sip, err := ipv4.FromDots(d.ds.DiscoverConf.StartIP)
	if err != nil {
		return fmt.Errorf("discover start ip err=%v", err)
	}
	eip, err := ipv4.FromDots(d.ds.DiscoverConf.EndIP)
	if err != nil {
		return fmt.Errorf("discover end ip err=%v", err)
	}
	if sip > eip {
		return fmt.Errorf("discover start ip > end ip")
	}
	d.ds.AddEventLog(datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf("自動発見開始 %s - %s", d.ds.DiscoverConf.StartIP, d.ds.DiscoverConf.EndIP),
	})
	d.Stop = false
	d.Stat.Total = eip - sip + 1
	d.Stat.Sent = 0
	d.Stat.Found = 0
	d.Stat.Snmp = 0
	d.Stat.Running = true
	d.Stat.StartTime = time.Now().UnixNano()
	d.Stat.EndTime = 0
	d.X = (1 + d.ds.DiscoverConf.X/GRID) * GRID
	d.Y = (1 + d.ds.DiscoverConf.Y/GRID) * GRID
	var mu sync.Mutex
	sem := make(chan bool, 20)
	go func() {
		for ; sip <= eip && !d.Stop; sip++ {
			sem <- true
			d.Stat.Sent++
			d.Stat.Progress = (100 * d.Stat.Sent) / d.Stat.Total
			go func(ip uint32) {
				defer func() {
					<-sem
				}()
				ipstr := ipv4.ToDots(ip)
				if d.ds.FindNodeFromIP(ipstr) != nil {
					return
				}
				r := d.ping.DoPing(ipstr, d.ds.DiscoverConf.Timeout, d.ds.DiscoverConf.Retry, 64)
				if r.Stat == ping.PingOK {
					dent := discoverInfoEnt{
						IP:          ipstr,
						IfIndexList: []string{},
					}
					if names, err := net.LookupAddr(ipstr); err == nil && len(names) > 0 {
						dent.HostName = names[0]
					}
					d.discoverGetSnmpInfo(ipstr, &dent)
					mu.Lock()
					dent.X = d.X
					dent.Y = d.Y
					d.Stat.Found++
					d.X += GRID
					if d.X > GRID*10 {
						d.X = GRID
						d.Y += GRID
					}
					if dent.SysName != "" {
						d.Stat.Snmp++
					}
					d.addFoundNode(dent)
					mu.Unlock()
				}
			}(sip)
		}
		for len(sem) > 0 {
			time.Sleep(time.Millisecond * 10)
		}
		d.Stat.Running = false
		d.Stat.EndTime = time.Now().UnixNano()
		d.ds.AddEventLog(datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("自動発見終了 %s - %s", d.ds.DiscoverConf.StartIP, d.ds.DiscoverConf.EndIP),
		})
	}()
	return nil
}

func (d *Discover) discoverGetSnmpInfo(t string, dent *discoverInfoEnt) {
	agent := &gosnmp.GoSNMP{
		Target:             t,
		Port:               161,
		Transport:          "udp",
		Community:          d.ds.MapConf.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(2) * time.Second,
		Retries:            1,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	if d.ds.MapConf.SnmpMode != "" {
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		if d.ds.MapConf.SnmpMode == "v3auth" {
			agent.MsgFlags = gosnmp.AuthNoPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 d.ds.MapConf.SnmpUser,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: d.ds.MapConf.SnmpPassword,
			}
		} else {
			agent.MsgFlags = gosnmp.AuthPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 d.ds.MapConf.SnmpUser,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: d.ds.MapConf.SnmpPassword,
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        d.ds.MapConf.SnmpPassword,
			}
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("discoverGetSnmpInfo err=%v", err)
		return
	}
	defer agent.Conn.Close()
	oids := []string{d.ds.MIBDB.NameToOID("sysName"), d.ds.MIBDB.NameToOID("sysObjectID")}
	result, err := agent.GetNext(oids)
	if err != nil {
		log.Printf("discoverGetSnmpInfo err=%v", err)
		return
	}
	for _, variable := range result.Variables {
		if d.ds.MIBDB.OIDToName(variable.Name) == "sysName.0" {
			dent.SysName = variable.Value.(string)
		} else if d.ds.MIBDB.OIDToName(variable.Name) == "sysObjectID.0" {
			dent.SysObjectID = variable.Value.(string)
		}
	}
	_ = agent.Walk(d.ds.MIBDB.NameToOID("ifType"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(d.ds.MIBDB.OIDToName(variable.Name), ".")
		if len(a) == 2 &&
			a[0] == "ifType" &&
			gosnmp.ToBigInt(variable.Value).Int64() == 6 {
			dent.IfIndexList = append(dent.IfIndexList, a[1])
		}
		return nil
	})
}

func (d *Discover) addFoundNode(dent discoverInfoEnt) {
	n := datastore.NodeEnt{
		Name:  dent.HostName,
		IP:    dent.IP,
		Icon:  "desktop",
		X:     dent.X,
		Y:     dent.Y,
		Descr: "自動登録:" + time.Now().Format(time.RFC3339),
	}
	if n.Name == "" {
		if dent.SysName != "" {
			n.Name = dent.SysName
		} else {
			n.Name = dent.IP
		}
	}
	if dent.SysObjectID != "" {
		n.SnmpMode = d.ds.MapConf.SnmpMode
		n.User = d.ds.MapConf.SnmpUser
		n.Password = d.ds.MapConf.SnmpPassword
		n.Community = d.ds.MapConf.Community
		n.Icon = "hdd"
	}
	if err := d.ds.AddNode(&n); err != nil {
		log.Printf("discover AddNode err=%v", err)
		return
	}
	d.ds.AddEventLog(datastore.EventLogEnt{
		Type:     "discover",
		Level:    "info",
		NodeID:   n.ID,
		NodeName: n.Name,
		Event:    "自動発見により追加",
	})
	p := &datastore.PollingEnt{
		NodeID:  n.ID,
		Name:    "PING監視",
		Type:    "ping",
		Level:   "low",
		State:   "unknown",
		PollInt: d.ds.MapConf.PollInt,
		Timeout: d.ds.MapConf.Timeout,
		Retry:   d.ds.MapConf.Retry,
	}
	if err := d.ds.AddPolling(p); err != nil {
		log.Printf("discover AddPolling err=%v", err)
		return
	}
	if dent.SysObjectID == "" {
		return
	}
	p = &datastore.PollingEnt{
		NodeID:  n.ID,
		Name:    "sysUptime監視",
		Type:    "snmp",
		Polling: "sysUpTime",
		Level:   "low",
		State:   "unknown",
		PollInt: d.ds.MapConf.PollInt,
		Timeout: d.ds.MapConf.Timeout,
		Retry:   d.ds.MapConf.Retry,
	}
	if err := d.ds.AddPolling(p); err != nil {
		log.Printf("discover AddPolling err=%v", err)
		return
	}
	for _, i := range dent.IfIndexList {
		p = &datastore.PollingEnt{
			NodeID:  n.ID,
			Type:    "snmp",
			Name:    "IF " + i + "監視",
			Polling: "ifOperStatus." + i,
			Level:   "low",
			State:   "unknown",
			PollInt: d.ds.MapConf.PollInt,
			Timeout: d.ds.MapConf.Timeout,
			Retry:   d.ds.MapConf.Retry,
		}
		if err := d.ds.AddPolling(p); err != nil {
			log.Printf("discover AddPolling err=%v", err)
			return
		}
	}
}
