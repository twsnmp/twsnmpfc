package report

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	etherTypeCount = 0
	ipToMacCount   = 0
	ntpCount       = 0
	dhcpCount      = 0
	dnsCount       = 0
	radiusCount    = 0
	tlsCount       = 0
	statsCount     = 0
	otherCount     = 0
)

func ReportTWPCAP(l map[string]interface{}) {
	twpcapReportCh <- l
}

func checkTWPCAPReport(l map[string]interface{}) {
	h, ok := l["hostname"].(string)
	if !ok {
		log.Printf("twpcap no hostname %v", l)
		return
	}
	m, ok := l["content"].(string)
	if !ok {
		log.Printf("twpcap no content %v", l)
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
		log.Printf("twpcap no type %v", twpcapMap)
		return
	}
	switch t {
	case "IPToMAC":
		checkIPTOMACReport(twpcapMap)
		ipToMacCount++
	case "EtherType":
		checkEtherTypeReport(h, twpcapMap)
		etherTypeCount++
	case "DNS":
		checkDNSReport(h, twpcapMap)
		dnsCount++
	case "DHCP":
		checkDHCPReport(twpcapMap)
		dhcpCount++
	case "NTP":
		checkNTPReport(twpcapMap)
		ntpCount++
	case "RADIUS":
		checkRADIUSReport(twpcapMap)
		radiusCount++
	case "TLSFlow":
		checkTLSFlowReport(twpcapMap)
		tlsCount++
	case "Stats":
		statsCount++
	default:
		log.Printf("twpcap unkown type %v", twpcapMap)
		otherCount++
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
	lts, ok := twpcap["lt"]
	if !ok {
		return
	}
	lt, err := time.Parse(time.RFC3339, lts)
	if err != nil {
		log.Printf("twpcap report err=%v", err)
		return
	}
	mac = normMACAddr(mac)
	// Device Report
	if !strings.Contains(ip, ":") {
		updateDeviceReport(mac, ip, lt.UnixNano())
	}

	// IP Report
	checkIPReport(ip, mac, lt.UnixNano())
}

// Ethernet type report
func checkEtherTypeReport(h string, twpcap map[string]string) {
	now := time.Now().UnixNano()
	for k, v := range twpcap {
		if strings.HasPrefix(k, "0x") {
			c := getNumberFromTWPCAPLog(v)
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
	"0x0800": "IPv4",
	"0x0806": "ARP",
	"0x0842": "WakeOnLAN",
	"0x8035": "RARP",
	"0x86dd": "IPv6",
	"0x8899": "RRCP",
	"0x88cc": "LLDP",
	"0x8100": "VLAN",
	"0x9100": "VLAN DT",
	"0x8847": "MPLS U",
	"0x8848": "MPLS M",
	"0x8863": "PPPoE DS",
	"0x8864": "PPPoE SS",
	"0x888e": "802.1X",
	"0x88a2": "ATAoE",
	"0x9000": "EConfTest",
}

func getEtherTypeName(t string) string {
	if n, ok := etherTypeMap[t]; ok {
		return n
	}
	return fmt.Sprintf("Other(%s)", t)
}

func getNumberFromTWPCAPLog(s string) int64 {
	if c, err := strconv.ParseInt(s, 10, 64); err == nil {
		return c
	}
	return 0
}

func getTimeFromTWPCAPLog(s string) int64 {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UnixNano()
	}
	return time.Now().UnixNano()
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
		e.Count = getNumberFromTWPCAPLog(twpcap["count"])
		e.Change = getNumberFromTWPCAPLog(twpcap["change"])
		e.LastTime = getTimeFromTWPCAPLog(twpcap["lt"])
		e.FirstTime = getTimeFromTWPCAPLog(twpcap["ft"])
		return
	}
	datastore.AddDNSQ(&datastore.DNSQEnt{
		ID:        id,
		Host:      h,
		Type:      t,
		Server:    sv,
		Name:      n,
		Count:     getNumberFromTWPCAPLog(twpcap["count"]),
		Change:    getNumberFromTWPCAPLog(twpcap["change"]),
		LastTime:  getTimeFromTWPCAPLog(twpcap["lt"]),
		FirstTime: getTimeFromTWPCAPLog(twpcap["ft"]),
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
		e.Accept = getNumberFromTWPCAPLog(twpcap["accept"])
		e.Reject = getNumberFromTWPCAPLog(twpcap["reject"])
		e.Request = getNumberFromTWPCAPLog(twpcap["request"])
		e.Challenge = getNumberFromTWPCAPLog(twpcap["challenge"])
		e.Count = getNumberFromTWPCAPLog(twpcap["count"])
		e.LastTime = getTimeFromTWPCAPLog(twpcap["lt"])
		e.FirstTime = getTimeFromTWPCAPLog(twpcap["ft"])
		e.UpdateTime = time.Now().UnixNano()
		setRADIUSFlowPenalty(e)
		return
	}
	e = &datastore.RADIUSFlowEnt{
		ID:         id,
		Client:     cl,
		Server:     sv,
		Accept:     getNumberFromTWPCAPLog(twpcap["accept"]),
		Request:    getNumberFromTWPCAPLog(twpcap["req"]),
		Challenge:  getNumberFromTWPCAPLog(twpcap["challenge"]),
		Reject:     getNumberFromTWPCAPLog(twpcap["reject"]),
		Count:      getNumberFromTWPCAPLog(twpcap["count"]),
		LastTime:   getTimeFromTWPCAPLog(twpcap["lt"]),
		FirstTime:  getTimeFromTWPCAPLog(twpcap["ft"]),
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
		f.Count = getNumberFromTWPCAPLog(twpcap["count"])
		f.FirstTime = getTimeFromTWPCAPLog(twpcap["ft"])
		f.LastTime = getTimeFromTWPCAPLog(twpcap["lt"])
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
		FirstTime:  getTimeFromTWPCAPLog(twpcap["ft"]),
		LastTime:   getTimeFromTWPCAPLog(twpcap["lt"]),
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
	if strings.Contains(f.Version, "1.0") || strings.Contains(f.Version, "1.1") || strings.Contains(f.Version, "SSL") {
		f.Penalty++
	}
	// DNSで解決できない場合
	if f.ServerName == f.Server {
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
