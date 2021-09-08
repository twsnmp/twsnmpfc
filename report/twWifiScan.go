package report

import (
	"log"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
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

//type=APInfo,ssid=%s,bssid=%s,rssi=%s,Channel=%s,info=%s,count=%d,change=%d,ft=%s,lt=%s
func checkAPInfoReport(h string, m map[string]string) {
	bssid, ok := m["bssid"]
	if !ok || bssid == "" {
		return
	}
	rssi := getNumberFromTWLog(m["rssi"])
	id := h + ":" + bssid
	now := time.Now().Unix()
	e := datastore.GetWifiAP(id)
	if e != nil {
		e.Count++
		if e.SSID != m["ssid"] || e.Channel != m["Channel"] || e.Info != m["info"] {
			e.Change++
		}
		e.SSID = m["ssid"]
		e.Channel = m["Channel"]
		e.Info = m["info"]
		e.LastTime = now
		e.RSSI = append(e.RSSI, datastore.RSSIEnt{Value: int(rssi), Time: now})
		if len(e.RSSI) > MAX_DATA_SIZE {
			e.RSSI = e.RSSI[1:]
		}
		return
	}
	datastore.AddWifiAP(&datastore.WifiAPEnt{
		ID:    id,
		Host:  h,
		BSSID: bssid,
		Count: 1,
		RSSI: []datastore.RSSIEnt{
			{Value: int(rssi), Time: now},
		},
		SSID:      m["ssid"],
		Channel:   m["Channel"],
		Info:      m["info"],
		LastTime:  now,
		FirstTime: now,
	})
}
