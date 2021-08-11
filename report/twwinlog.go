package report

import (
	"log"
	"strings"
)

var (
	otherWinLogCount = 0
)

func ReportTwWinLog(l map[string]interface{}) {
	twWinLogCh <- l
}

func checkTWWinLogReport(l map[string]interface{}) {
	h, ok := l["hostname"].(string)
	if !ok {
		log.Printf("twwinlog no hostname %v", l)
		return
	}
	msg, ok := l["content"].(string)
	if !ok {
		log.Printf("twwinlog no content %v", l)
		return
	}
	kvs := strings.Split(msg, ",")
	var m = make(map[string]string)
	for _, kv := range kvs {
		a := strings.SplitN(kv, "=", 2)
		if len(a) == 2 {
			m[a[0]] = a[1]
		}
	}
	t, ok := m["type"]
	if !ok {
		log.Printf("twwinlog no type %v", m)
		return
	}
	switch t {
	case "Stats":
		checkStats(h, "twwinlog", m)
	case "Monitor":
		checkMonitor(h, "twwinlog", m)
	default:
		log.Printf("twwinlog unkown type %v", m)
		otherWinLogCount++
	}
}
