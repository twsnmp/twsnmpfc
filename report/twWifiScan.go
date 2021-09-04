package report

import (
	"log"
	"strings"
)

func ReportTWWifiScan(l map[string]interface{}) {
	twWifiScanCh <- l
}

func checkTWWifiScanReport(l map[string]interface{}) {
	h, ok := l["hostname"].(string)
	if !ok {
		log.Printf("twWifiScan no hostname %v", l)
		return
	}
	c, ok := l["content"].(string)
	if !ok {
		log.Printf("twWifiScan no content %v", l)
		return
	}
	kvs := strings.Split(c, ",")
	var m = make(map[string]string)
	for _, kv := range kvs {
		a := strings.SplitN(kv, "=", 2)
		if len(a) == 2 {
			m[a[0]] = a[1]
		}
	}
	t, ok := m["type"]
	if !ok {
		log.Printf("twWifiScan no type %v", m)
		return
	}
	switch t {
	case "APInfo":
		checkAPInfoReport(h, m)
	case "Stats":
		checkStats(h, "twWifiScan", m)
	case "Monitor":
		checkMonitor(h, "twWifiScan", m)
	default:
		log.Printf("twWifiScan unkown type %v", m)
	}
}

func checkAPInfoReport(h string, m map[string]string) {

}
