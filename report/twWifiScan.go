package report

import (
	"log"
	"sort"
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
		return
	}
	c, ok := l["content"].(string)
	if !ok {
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
		log.Printf("twwifiscan unknown type=%v", t)
	}
}

// type=APInfo,ssid=%s,bssid=%s,rssi=%s,Channel=%s,info=%s,count=%d,change=%d,ft=%s,lt=%s
func checkAPInfoReport(h string, m map[string]string) {
	bssid, ok := m["bssid"]
	if !ok || bssid == "" {
		return
	}
	rssi := getNumberFromTWLog(m["rssi"])
	id := h + ":" + bssid
	now := time.Now().UnixNano()
	e := datastore.GetWifiAP(id)
	if e != nil {
		e.Count++
		if e.SSID != m["ssid"] || e.Channel != m["Channel"] || e.Info != m["info"] {
			e.Change++
		}
		if e.Vendor == "" {
			e.Vendor = datastore.FindVendor(bssid)
		}
		e.SSID = m["ssid"]
		e.Channel = m["Channel"]
		e.Info = m["info"]
		e.LastTime = now
		e.RSSI = append(e.RSSI, datastore.RSSIEnt{Value: int(rssi), Time: now})
		if len(e.RSSI) > MaxDataSize {
			e.RSSI = e.RSSI[1:]
		}
		return
	}
	datastore.AddWifiAP(&datastore.WifiAPEnt{
		ID:     id,
		Host:   h,
		BSSID:  bssid,
		Vendor: datastore.FindVendor(bssid),
		Count:  1,
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

func checkOldWifiAP() {
	ids := []string{}
	list := []*datastore.WifiAPEnt{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachWifiAP(func(e *datastore.WifiAPEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		} else {
			list = append(list, e)
		}
		return true
	})
	if datastore.ReportConf.Limit < len(list) {
		sort.Slice(list, func(i, j int) bool {
			if list[i].LastTime == list[j].LastTime {
				return list[i].Count < list[j].Count
			}
			return list[i].LastTime < list[j].LastTime
		})
		for i := 0; i < len(list)-datastore.ReportConf.Limit; i++ {
			ids = append(ids, list[i].ID)
		}
	}
	if len(ids) > 0 {
		datastore.DeleteReport("wifiAP", ids)
	}
}
