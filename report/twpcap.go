package report

import (
	"log"
	"strings"
	"time"
)

func ReportTWPCAP(log map[string]interface{}) {
	twpcapReportCh <- log
}

func checkTWPCAPReport(log map[string]interface{}) {
	h, ok := log["hostname"].(string)
	if !ok {
		return
	}
	m, ok := log["content"].(string)
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
		checkDHCPReport(h, twpcapMap)
	case "NTP":
		checkNTPReport(h, twpcapMap)
	case "RADIUS":
		checkRADIUSReport(h, twpcapMap)
	case "TLSFlow":
		checkTLSFlowReport(h, twpcapMap)
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

func checkEtherTypeReport(h string, d map[string]string) {

}

func checkDNSReport(h string, d map[string]string) {

}
func checkDHCPReport(h string, d map[string]string) {

}

func checkRADIUSReport(h string, d map[string]string) {

}

func checkNTPReport(h string, d map[string]string) {

}

func checkTLSFlowReport(h string, d map[string]string) {

}
