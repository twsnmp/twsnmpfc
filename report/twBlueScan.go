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
	case "SwitchBotEnv":
		checkSwitchBotEnvReport(h, m)
	case "Stats":
		checkStats(h, "twBlueScan", m)
	case "Monitor":
		checkMonitor(h, "twBlueScan", m)
	default:
		log.Printf("twbluescan unknown type=%v m=%v", t, m)
	}
}

// type=Device,address=%s,name=%s,rssi=%d,addrType=%s,vendor=%s,info=%s,ft=%s,lt=%s
func checkBlueDeviceReport(h string, m map[string]string) {
	addr, ok := m["address"]
	if !ok {
		return
	}
	lt := getTimeFromTWLog(m["ft"])
	rssi := getNumberFromTWLog(m["rssi"])
	id := makeID(h + ":" + addr)
	e := datastore.GetBlueDevice(id)
	if e != nil {
		e.Count++
		e.LastTime = lt
		if i, ok := m["info"]; ok && i != "" && i != e.Info {
			e.Info = i
		}
		if v, ok := m["vendor"]; ok && v != "" && v != e.Vendor {
			e.Vendor = v
		}
		e.RSSI = append(e.RSSI, datastore.RSSIEnt{Value: int(rssi), Time: lt})
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
			{Value: int(rssi), Time: lt},
		},
		Vendor:    m["vendor"],
		Info:      m["info"],
		LastTime:  lt,
		FirstTime: getTimeFromTWLog(m["ft"]),
	})
}

// type=OMRONEnv,address=%s,name=%s,rssi=%d,seq=%d,temp=%.02f,hum=%.02f,lx=%d,press=%.02f,sound=%.02f,eTVOC=%d,eCO2=%d
func checkOMRONEnvReport(h string, m map[string]string) {
	addr, ok := m["address"]
	if !ok {
		return
	}
	rssi := getNumberFromTWLog(m["rssi"])
	id := makeID(h + ":" + addr)
	now := time.Now().UnixNano()
	e := datastore.GetEnvMonitor(id)
	if e != nil {
		e.Count++
		e.LastTime = time.Now().UnixNano()
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

// type=SwitchBotEnv,address=%s,name=%s,rssi=%d,temp=%.02f,hum=%.02f,bat=%d
func checkSwitchBotEnvReport(h string, m map[string]string) {
	addr, ok := m["address"]
	if !ok {
		return
	}
	rssi := getNumberFromTWLog(m["rssi"])
	id := makeID(h + ":" + addr)
	now := time.Now().UnixNano()
	e := datastore.GetEnvMonitor(id)
	if e != nil {
		e.Count++
		e.LastTime = time.Now().UnixNano()
		e.EnvData = append(e.EnvData, datastore.EnvDataEnt{
			Time:     now,
			RSSI:     int(rssi),
			Temp:     getFloatFromTWLog(m["temp"]),
			Humidity: getFloatFromTWLog(m["hum"]),
			Battery:  int(getNumberFromTWLog(m["bat"])),
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
				Time:     now,
				RSSI:     int(rssi),
				Temp:     getFloatFromTWLog(m["temp"]),
				Humidity: getFloatFromTWLog(m["hum"]),
				Battery:  int(getNumberFromTWLog(m["bat"])),
			},
		},
		LastTime:  now,
		FirstTime: now,
	})
}

func checkOldBlueDevice(safeOld, delOld int64) {
	ids := []string{}
	datastore.ForEachBludeDevice(func(e *datastore.BlueDeviceEnt) bool {
		if e.LastTime < safeOld {
			if e.LastTime < delOld || (strings.HasPrefix(e.AddressType, "LE Random(") && e.Name == "" && e.Count < 120) {
				ids = append(ids, e.ID)
			}
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("blueDevice", ids)
		log.Printf("delete blueDevice=%d", len(ids))
	}
}

func checkOldEnvMonitor(delOld int64) {
	ids := []string{}
	datastore.ForEachEnvMonitor(func(e *datastore.EnvMonitorEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("envMonitor", ids)
		log.Printf("delete envMonitor=%d", len(ids))
	}
}
