package report

import (
	"log"
	"strings"
)

func ReportTWBuleScan(l map[string]interface{}) {
	twBlueScanCh <- l
}

func checkTWBlueScanReport(l map[string]interface{}) {
	h, ok := l["hostname"].(string)
	if !ok {
		log.Printf("twBlueScan no hostname %v", l)
		return
	}
	c, ok := l["content"].(string)
	if !ok {
		log.Printf("twBlueScan no content %v", l)
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
		log.Printf("twBlueScan no type %v", m)
		return
	}
	switch t {
	case "Device":
		checkBlueDeviceReport(h, m)
	case "OMRONEnv":
		checkOMRONEnvReport(h, m)
	case "Stats":
		checkStats(h, "twBlueScan", m)
	case "Monitor":
		checkMonitor(h, "twBlueScan", m)
	default:
		log.Printf("twBlueScan unkown type %v", m)
	}
}

func checkBlueDeviceReport(h string, m map[string]string) {

}

func checkOMRONEnvReport(h string, m map[string]string) {

}
