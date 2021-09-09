package report

import (
	"log"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

const MAX_DATA_SIZE = 12 * 24 * 7

func ReportTWBuleScan(l map[string]interface{}) {
	twBlueScanCh <- l
}

func checkTWBlueScanReport(l map[string]interface{}) {
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
	case "Device":
		checkBlueDeviceReport(h, m)
	case "OMRONEnv":
		checkOMRONEnvReport(h, m)
	case "Stats":
		checkStats(h, "twBlueScan", m)
	case "Monitor":
		checkMonitor(h, "twBlueScan", m)
	default:
		log.Printf("twbluescan unkown type=%v", t)
	}
}

// type=Device,address=%s,name=%s,rssi=%d,addrType=%s,vendor=%s,md=%s
func checkBlueDeviceReport(h string, m map[string]string) {
	addr, ok := m["address"]
	if !ok {
		return
	}
	rssi := getNumberFromTWLog(m["rssi"])
	id := h + ":" + addr
	now := time.Now().Unix()
	e := datastore.GetBlueDevice(id)
	if e != nil {
		e.Count++
		e.LastTime = now
		e.RSSI = append(e.RSSI, datastore.RSSIEnt{Value: int(rssi), Time: now})
		if len(e.RSSI) > MAX_DATA_SIZE {
			e.RSSI = e.RSSI[1:]
		}
		return
	}
	datastore.AddBlueDevice(&datastore.BlueDeviceEnt{
		ID:          id,
		Host:        h,
		Address:     addr,
		AddressType: m["addrType"],
		Name:        m["name"],
		Count:       1,
		RSSI: []datastore.RSSIEnt{
			{Value: int(rssi), Time: now},
		},
		Vendor:    m["vendor"],
		ExtData:   m["md"],
		LastTime:  now,
		FirstTime: now,
	})
}

// type=OMRONEnv,address=%s,name=%s,rssi=%d,seq=%d,temp=%.02f,hum=%.02f,lx=%d,press=%.02f,sound=%.02f,eTVOC=%d,eCO2=%d
func checkOMRONEnvReport(h string, m map[string]string) {
	addr, ok := m["address"]
	if !ok {
		return
	}
	rssi := getNumberFromTWLog(m["rssi"])
	id := h + ":" + addr
	now := time.Now().Unix()
	e := datastore.GetEnvMonitor(id)
	if e != nil {
		e.Count++
		e.LastTime = time.Now().Unix()
		e.EnvData = append(e.EnvData, datastore.EnvDataEnt{
			Time:               now,
			RSSI:               int(rssi),
			Temp:               getFloatFromTWLog(m["temp"]),
			Humidity:           getFloatFromTWLog(m["hum"]),
			Illuminance:        getFloatFromTWLog(m["lx"]),
			BarometricPressure: getFloatFromTWLog(m["press"]),
			Sound:              getFloatFromTWLog(m["sound"]),
			ETVOC:              getFloatFromTWLog(m["eTVOC"]),
			ECo2:               getFloatFromTWLog(m["eCO2"]),
		})
		if len(e.EnvData) > MAX_DATA_SIZE {
			e.EnvData = e.EnvData[1:]
		}
		return
	}
	datastore.AddEnvMonitor(&datastore.EnvMonitorEnt{
		ID:      id,
		Host:    h,
		Address: addr,
		Name:    m["name"],
		Count:   1,
		EnvData: []datastore.EnvDataEnt{
			{
				Time:               now,
				RSSI:               int(rssi),
				Temp:               getFloatFromTWLog(m["temp"]),
				Humidity:           getFloatFromTWLog(m["hum"]),
				Illuminance:        getFloatFromTWLog(m["lx"]),
				BarometricPressure: getFloatFromTWLog(m["press"]),
				Sound:              getFloatFromTWLog(m["sound"]),
				ETVOC:              getFloatFromTWLog(m["eTVOC"]),
				ECo2:               getFloatFromTWLog(m["eCO2"]),
			},
		},
		LastTime:  now,
		FirstTime: now,
	})
}
