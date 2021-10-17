package report

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func ReportTWPCAP(l map[string]interface{}) {
	twpcapReportCh <- l
}

func checkTWPCAPReport(l map[string]interface{}) {
	h, ok := l["hostname"].(string)
	if !ok {
		return
	}
	m, ok := l["content"].(string)
	if !ok {
		return
	}
	kvs := strings.Split(m, ",")
	var twpcapMap = make(map[string]string)
	for _, kv := range kvs {
		a := strings.SplitN(kv, "=", 2)
		if len(a) == 2 {
			twpcapMap[a[0]] = a[1]
		}
	}
	t, ok := twpcapMap["type"]
	if !ok {
		return
	}
	switch t {
	case "IPToMAC":
		checkIPTOMACReport(twpcapMap)
	case "EtherType":
		checkEtherTypeReport(h, twpcapMap)
	case "DNS":
		checkDNSReport(h, twpcapMap)
	case "DHCP":
		checkDHCPReport(twpcapMap)
	case "NTP":
		checkNTPReport(twpcapMap)
	case "RADIUS":
		checkRADIUSReport(twpcapMap)
	case "TLSFlow":
		checkTLSFlowReport(twpcapMap)
	case "Stats":
		checkStats(h, "twpcap", twpcapMap)
	case "Monitor":
		checkMonitor(h, "twpcap", twpcapMap)
	default:
		log.Printf("twpcap unkown type=%v", t)
	}
}

func checkIPTOMACReport(twpcap map[string]string) {
	mac, ok := twpcap["mac"]
	if !ok {
		return
	}
	ip, ok := twpcap["ip"]
	if !ok {
		return
	}
	lt := getTimeFromTWLog("lt")
	mac = normMACAddr(mac)
	// Device Report
	if !strings.Contains(ip, ":") {
		updateDeviceReport(mac, ip, lt)
	}

	// IP Report
	checkIPReport(ip, mac, lt)
}

// Ethernet type report
func checkEtherTypeReport(h string, twpcap map[string]string) {
	now := time.Now().UnixNano()
	for k, v := range twpcap {
		if strings.HasPrefix(k, "0x") {
			c := getNumberFromTWLog(v)
			id := h + ":" + k
			e := datastore.GetEtherType(id)
			if e != nil {
				e.Count += c
				e.LastTime = now
				continue
			}
			datastore.AddEtherType(&datastore.EtherTypeEnt{
				ID:        id,
				Host:      h,
				Type:      k,
				Count:     c,
				Name:      getEtherTypeName(k),
				FirstTime: now,
				LastTime:  now,
			})
		}
	}
}

var etherTypeMap = map[string]string{
	"0x0000": "LLC",
	"0x0800": "IPv4",
	"0x0806": "ARP",
	"0x0842": "WakeOnLAN",
	"0x8035": "RARP",
	"0x86dd": "IPv6",
	"0x8899": "RRCP",
	"0x88cc": "LLDP",
	"0x8100": "VLAN",
	"0x9100": "VLAN DT",
	"0x8847": "MPLS Unicat",
	"0x8848": "MPLS Multicast",
	"0x8863": "PPPoE Discovery",
	"0x8864": "PPPoE Session",
	"0x888e": "802.1X",
	"0x88a2": "ATAoE",
	"0x9000": "Ethernet Conf Test",
	"0x87d2": "Aironet DDP",
	"0x890d": "802.11MP",
	"0x2000": "Cisco Discovery",
	"0x01a2": "Nortel Discovery",
	"0x6558": "Transparent Ethernet Bridging",
	"0x880b": "PPP",
	"0x88be": "ERSPAN",
	"0x88a8": "QinQ",
}

func getEtherTypeName(t string) string {
	if n, ok := etherTypeMap[t]; ok {
		return n
	}
	return fmt.Sprintf("Other(%s)", t)
}

func checkDNSReport(h string, twpcap map[string]string) {
	t, ok := twpcap["DNSType"]
	if !ok {
		return
	}
	n, ok := twpcap["Name"]
	if !ok {
		return
	}
	sv, ok := twpcap["sv"]
	if !ok {
		return
	}
	id := h + ":" + sv + ":" + t + ":" + n
	e := datastore.GetDNSQ(id)
	if e != nil {
		e.Count = getNumberFromTWLog(twpcap["count"])
		e.Change = getNumberFromTWLog(twpcap["change"])
		e.LastTime = getTimeFromTWLog(twpcap["lt"])
		e.FirstTime = getTimeFromTWLog(twpcap["ft"])
		return
	}
	datastore.AddDNSQ(&datastore.DNSQEnt{
		ID:        id,
		Host:      h,
		Type:      t,
		Server:    sv,
		Name:      n,
		Count:     getNumberFromTWLog(twpcap["count"]),
		Change:    getNumberFromTWLog(twpcap["change"]),
		LastTime:  getTimeFromTWLog(twpcap["lt"]),
		FirstTime: getTimeFromTWLog(twpcap["ft"]),
	})
}

func checkDHCPReport(twpcap map[string]string) {
	sv, ok := twpcap["sv"]
	if !ok {
		return
	}
	e := datastore.GetServer(sv)
	if e != nil {
		e.DHCPInfo = fmt.Sprintf("count=%s,offer=%s,ack=%s,nak=%s",
			twpcap["count"],
			twpcap["offer"],
			twpcap["ack"],
			twpcap["nak"],
		)
		return
	}
	checkServerReport(sv, "bootps/udp", 0, time.Now().UnixNano())
}

func checkNTPReport(twpcap map[string]string) {
	sv, ok := twpcap["sv"]
	if !ok {
		return
	}
	e := datastore.GetServer(sv)
	if e != nil {
		e.NTPInfo = fmt.Sprintf("count=%s,change=%s,version=%s,stratum=%s,refid=%s",
			twpcap["count"],
			twpcap["change"],
			twpcap["version"],
			twpcap["stratum"],
			twpcap["refid"],
		)
		return
	}
	checkServerReport(sv, "ntp/udp", 0, time.Now().UnixNano())
}

func checkRADIUSReport(twpcap map[string]string) {
	sv, ok := twpcap["sv"]
	if !ok {
		return
	}
	cl, ok := twpcap["cl"]
	if !ok {
		return
	}
	id := cl + ":" + sv
	e := datastore.GetRADIUSFlow(id)
	if e != nil {
		e.Accept = getNumberFromTWLog(twpcap["accept"])
		e.Reject = getNumberFromTWLog(twpcap["reject"])
		e.Request = getNumberFromTWLog(twpcap["req"])
		e.Challenge = getNumberFromTWLog(twpcap["challenge"])
		e.Count = getNumberFromTWLog(twpcap["count"])
		e.LastTime = getTimeFromTWLog(twpcap["lt"])
		e.FirstTime = getTimeFromTWLog(twpcap["ft"])
		e.UpdateTime = time.Now().UnixNano()
		setRADIUSFlowPenalty(e)
		return
	}
	e = &datastore.RADIUSFlowEnt{
		ID:         id,
		Client:     cl,
		Server:     sv,
		Accept:     getNumberFromTWLog(twpcap["accept"]),
		Request:    getNumberFromTWLog(twpcap["req"]),
		Challenge:  getNumberFromTWLog(twpcap["challenge"]),
		Reject:     getNumberFromTWLog(twpcap["reject"]),
		Count:      getNumberFromTWLog(twpcap["count"]),
		LastTime:   getTimeFromTWLog(twpcap["lt"]),
		FirstTime:  getTimeFromTWLog(twpcap["ft"]),
		UpdateTime: time.Now().UnixNano(),
	}
	e.ClientName, e.ClientNodeID = findNodeInfoFromIP(cl)
	e.ServerName, e.ServerNodeID = findNodeInfoFromIP(sv)
	setRADIUSFlowPenalty(e)
	datastore.AddRADIUSFlow(e)
}

func checkTLSFlowReport(twpcap map[string]string) {
	sv, ok := twpcap["sv"]
	if !ok {
		return
	}
	cl, ok := twpcap["cl"]
	if !ok {
		return
	}
	service, ok := twpcap["serv"]
	if !ok {
		return
	}
	if service == "HTTPS" {
		checkHTTPSServer(sv, twpcap)
	}
	id := cl + ":" + sv + ":" + service
	f := datastore.GetTLSFlow(id)
	if f != nil {
		if f.ServerLoc == "" {
			f.ServerLoc = datastore.GetLoc(f.Server)
		}
		if f.ClientLoc == "" {
			f.ClientLoc = datastore.GetLoc(f.Client)
		}
		f.Cipher = twpcap["cipher"]
		f.Version = twpcap["maxver"]
		f.Count = getNumberFromTWLog(twpcap["count"])
		f.FirstTime = getTimeFromTWLog(twpcap["ft"])
		f.LastTime = getTimeFromTWLog(twpcap["lt"])
		f.UpdateTime = time.Now().UnixNano()
		return
	}
	f = &datastore.TLSFlowEnt{
		ID:         id,
		Client:     cl,
		Server:     sv,
		Service:    service,
		Count:      1,
		Version:    twpcap["maxver"],
		Cipher:     twpcap["cipher"],
		ServerLoc:  datastore.GetLoc(sv),
		ClientLoc:  datastore.GetLoc(cl),
		FirstTime:  getTimeFromTWLog(twpcap["ft"]),
		LastTime:   getTimeFromTWLog(twpcap["lt"]),
		UpdateTime: time.Now().UnixNano(),
	}
	f.ClientName, f.ClientNodeID = findNodeInfoFromIP(cl)
	f.ServerName, f.ServerNodeID = findNodeInfoFromIP(sv)
	setTLSFlowPenalty(f)
	datastore.AddTLSFlow(f)
}

func checkHTTPSServer(sv string, twpcap map[string]string) {
	e := datastore.GetServer(sv)
	if e != nil {
		e.TLSInfo = fmt.Sprintf("version=%s,cipher=%s",
			twpcap["maxver"],
			twpcap["cipher"],
		)
	}
}

func setTLSFlowPenalty(f *datastore.TLSFlowEnt) {
	f.Penalty = 0
	if !isSafeCountry(f.ServerLoc) {
		f.Penalty++
	}
	if !isSafeCountry(f.ClientLoc) {
		f.Penalty++
	}
	if strings.Contains(f.Version, "1.2") {
		f.Penalty++
	} else if strings.Contains(f.Version, "1.1") {
		f.Penalty += 2
	} else if strings.Contains(f.Version, "1.0") {
		f.Penalty += 3
	} else if strings.Contains(f.Version, "SSL") {
		f.Penalty += 4
	}
	// DNSで解決できない場合
	if f.ServerName == f.Server {
		f.Penalty++
	}
	if f.ClientName == f.Client {
		f.Penalty++
	}
}

func setRADIUSFlowPenalty(e *datastore.RADIUSFlowEnt) {
	e.Penalty = 0
	if e.Reject > 0 {
		e.Penalty++
		if e.Accept < e.Reject {
			e.Penalty++
		}
	}
	// DNSで解決できない場合
	if e.ServerName == e.Server {
		e.Penalty++
	}
}

func ResetRADIUSFlowsScore() {
	datastore.ForEachRADIUSFlows(func(f *datastore.RADIUSFlowEnt) bool {
		f.Penalty = 0
		f.ClientName, f.ClientNodeID = findNodeInfoFromIP(f.Client)
		f.ServerName, f.ServerNodeID = findNodeInfoFromIP(f.Server)
		setRADIUSFlowPenalty(f)
		f.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcRADIUSFlowScore()
}

func ResetTLSFlowsScore() {
	datastore.ForEachTLSFlows(func(f *datastore.TLSFlowEnt) bool {
		f.Penalty = 0
		f.ClientName, f.ClientNodeID = findNodeInfoFromIP(f.Client)
		f.ServerName, f.ServerNodeID = findNodeInfoFromIP(f.Server)
		setTLSFlowPenalty(f)
		f.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcTLSFlowScore()
}

func checkOldEtherType(delOld int64) {
	ids := []string{}
	datastore.ForEachEtherType(func(e *datastore.EtherTypeEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("ether", ids)
		log.Printf("delete etherType=%d", len(ids))
	}
}

func checkOldDNSQ(delOld int64) {
	ids := []string{}
	datastore.ForEachDNSQ(func(e *datastore.DNSQEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("dns", ids)
		log.Printf("delete DNSQ=%d", len(ids))
	}
}

func checkOldRadiusFlow(safeOld, delOld int64) {
	ids := []string{}
	datastore.ForEachRADIUSFlows(func(i *datastore.RADIUSFlowEnt) bool {
		if i.LastTime < safeOld {
			if i.LastTime < delOld || (i.Score > 50.0 && i.LastTime == i.FirstTime) {
				ids = append(ids, i.ID)
			}
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("radius", ids)
		log.Printf("report delete radiusFlow=%d", len(ids))
	}
}

func checkOldTLSFlow(safeOld, delOld int64) {
	ids := []string{}
	datastore.ForEachTLSFlows(func(i *datastore.TLSFlowEnt) bool {
		if i.LastTime < safeOld {
			if i.LastTime < delOld || (i.Score > 50.0 && i.LastTime == i.FirstTime) {
				ids = append(ids, i.ID)
			}
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("tls", ids)
		log.Printf("report delete tlsFlow=%d", len(ids))
	}
}

func calcRADIUSFlowScore() {
	var xs []float64
	datastore.ForEachRADIUSFlows(func(e *datastore.RADIUSFlowEnt) bool {
		if e.Penalty > 100 {
			e.Penalty = 100
		}
		xs = append(xs, float64(100-e.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachRADIUSFlows(func(e *datastore.RADIUSFlowEnt) bool {
		if sd != 0 {
			e.Score = ((10 * (float64(100-e.Penalty) - m) / sd) + 50)
		} else {
			e.Score = 50.0
		}
		e.ValidScore = true
		return true
	})
}

func calcTLSFlowScore() {
	var xs []float64
	datastore.ForEachTLSFlows(func(e *datastore.TLSFlowEnt) bool {
		if e.Penalty > 100 {
			e.Penalty = 100
		}
		xs = append(xs, float64(100-e.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachTLSFlows(func(e *datastore.TLSFlowEnt) bool {
		if sd != 0 {
			e.Score = ((10 * (float64(100-e.Penalty) - m) / sd) + 50)
		} else {
			e.Score = 50.0
		}
		e.ValidScore = true
		return true
	})
}
