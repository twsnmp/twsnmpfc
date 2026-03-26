package report

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

const MaxDataSize = 12 * 24 * 7

func ReportTWBuleScan(l map[string]interface{}) {
	twBlueScanCh <- l
}

func ReportEnvMonitor(id string, m map[string]interface{}) {
	var temp float64
	var hum float64
	var co2 float64
	var rssi float64
	var bat float64
	var name string
	if v, ok := m["temp"].(float64); ok {
		temp = v
	}
	if v, ok := m["hum"].(float64); ok {
		hum = v
	}
	if v, ok := m["co2"].(float64); ok {
		co2 = v
	}
	if v, ok := m["rssi"].(float64); ok {
		rssi = v
	}
	if v, ok := m["bat"].(float64); ok {
		bat = v
	}
	if v, ok := m["name"].(string); ok {
		name = v
	}
	now := time.Now().UnixNano()
	e := datastore.GetEnvMonitor(id)
	if e != nil {
		e.Count++
		e.LastTime = now
		e.EnvData = append(e.EnvData, datastore.EnvDataEnt{
			Time:     now,
			RSSI:     int(rssi),
			Temp:     temp,
			Humidity: hum,
			Battery:  int(bat),
			ECo2:     co2,
		})
		if len(e.EnvData) > MaxDataSize {
			e.EnvData = e.EnvData[1:]
		}
		return
	}
	datastore.AddEnvMonitor(&datastore.EnvMonitorEnt{
		ID:      id,
		Host:    "localhost",
		Address: "",
		Name:    name,
		Count:   1,
		EnvData: []datastore.EnvDataEnt{
			{
				Time:     now,
				RSSI:     int(rssi),
				Temp:     temp,
				Humidity: hum,
				Battery:  int(bat),
				ECo2:     co2,
			},
		},
		LastTime:  now,
		FirstTime: now,
	})
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
	case "InkbirdEnv":
		checkInkbirdEnvReport(h, m)
	case "SwitchBotPlugMini":
		checkSwitchBotPlugMiniReport(h, m)
	case "SwitchBotMotionSensor":
		checkSwitchBotMotionSensorReport(h, m)
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
	info := m["info"]
	if u, ok := m["uuid"]; ok {
		info += " UUID:" + u
	}
	e := datastore.GetBlueDevice(id)
	if e != nil {
		e.Count++
		e.LastTime = lt
		if info != "" && info != e.Info {
			e.Info = info
		}
		if v, ok := m["vendor"]; ok && v != "" && v != e.Vendor {
			e.Vendor = v
		}
		e.RSSI = append(e.RSSI, datastore.RSSIEnt{Value: int(rssi), Time: lt})
		if len(e.RSSI) > MaxDataSize {
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
		Info:      info,
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
		if len(e.EnvData) > MaxDataSize {
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

// type=SwitchBotEnv,address=%s,name=%s,rssi=%d,temp=%.02f,hum=%.02f,bat=%d,co2=%d
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
		e.LastTime = now
		e.EnvData = append(e.EnvData, datastore.EnvDataEnt{
			Time:     now,
			RSSI:     int(rssi),
			Temp:     getFloatFromTWLog(m["temp"]),
			Humidity: getFloatFromTWLog(m["hum"]),
			Battery:  int(getNumberFromTWLog(m["bat"])),
			ECo2:     float64(getNumberFromTWLog(m["co2"])),
		})
		if len(e.EnvData) > MaxDataSize {
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
				ECo2:     float64(getNumberFromTWLog(m["co2"])),
			},
		},
		LastTime:  now,
		FirstTime: now,
	})
}

func checkInkbirdEnvReport(h string, m map[string]string) {
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
		e.LastTime = now
		e.EnvData = append(e.EnvData, datastore.EnvDataEnt{
			Time:     now,
			RSSI:     int(rssi),
			Temp:     getFloatFromTWLog(m["temp"]),
			Humidity: getFloatFromTWLog(m["hum"]),
			Battery:  int(getNumberFromTWLog(m["bat"])),
			ECo2:     float64(getNumberFromTWLog(m["co2"])),
		})
		if len(e.EnvData) > MaxDataSize {
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
				ECo2:     float64(getNumberFromTWLog(m["co2"])),
			},
		},
		LastTime:  now,
		FirstTime: now,
	})
}

// type=SwitchBotPlugMini,address=%s,name=%s,rssi=%d,sw=%v,over=%v,load=%d
func checkSwitchBotPlugMiniReport(h string, m map[string]string) {
	addr, ok := m["address"]
	if !ok {
		return
	}
	if _, ok := m["load"]; !ok {
		return
	}
	rssi := getNumberFromTWLog(m["rssi"])
	load := getFloatFromTWLog(m["load"]) / 10.0
	sw := m["sw"] == "true"
	over := m["over"] == "true"
	id := makeID(h + ":" + addr)
	now := time.Now().UnixNano()
	e := datastore.GetPowerMonitor(id)
	if e != nil {
		e.Count++
		e.LastTime = time.Now().UnixNano()
		e.Data = append(e.Data, datastore.PowerMonitorDataEnt{
			Time:   now,
			Load:   load,
			Switch: sw,
			Over:   over,
			RSSI:   int(rssi),
		})
		if len(e.Data) > MaxDataSize {
			e.Data = e.Data[1:]
		}
		return
	}
	datastore.AddPowerMonitor(&datastore.PowerMonitorEnt{
		ID:      id,
		Host:    h,
		Address: addr,
		Name:    m["name"],
		Count:   1,
		Data: []datastore.PowerMonitorDataEnt{
			{
				Time:   now,
				Switch: sw,
				Over:   over,
				Load:   load,
				RSSI:   int(rssi),
			},
		},
		LastTime:  now,
		FirstTime: now,
	})
}

// 2024/02/06 06:13:49.663 info:local5 twBlueScan type=SwitchBotMotionSensor,address=d7:bb:ea:e7:cf:58,name=,rssi=-64,moving=false,event=report,lastMoveDiff=513,lastMove=2024-02-06T06:05:16+09:00,battery=228,light=false
func checkSwitchBotMotionSensorReport(h string, m map[string]string) {
	addr, ok := m["address"]
	if !ok {
		return
	}
	if _, ok := m["moving"]; !ok {
		return
	}
	event := m["event"]
	rssi := getNumberFromTWLog(m["rssi"])
	battery := getNumberFromTWLog(m["battery"])
	lastMoveDiff := getNumberFromTWLog(m["lastMoveDiff"])
	moving := m["moving"] == "true"
	light := m["light"] == "true"
	lastMove := getTimeFromTWLog(m["lastMove"])
	id := makeID(h + ":" + addr)
	now := time.Now().UnixNano()
	e := datastore.GetMotionSensor(id)
	if e != nil {
		e.Count++
		e.LastTime = time.Now().UnixNano()
		e.Data = append(e.Data, datastore.MotionSensorDataEnt{
			Time:         now,
			Event:        event,
			Moving:       moving,
			Light:        light,
			Battery:      battery,
			LastMove:     lastMove,
			LastMoveDiff: lastMoveDiff,
			RSSI:         int(rssi),
		})
		if len(e.Data) > MaxDataSize*2 {
			e.Data = e.Data[1:]
		}
		return
	}
	datastore.AddMotionSensor(&datastore.MotionSensorEnt{
		ID:      id,
		Host:    h,
		Address: addr,
		Name:    m["name"],
		Count:   1,
		Data: []datastore.MotionSensorDataEnt{
			{
				Time:         now,
				Event:        event,
				Moving:       moving,
				Light:        light,
				Battery:      battery,
				LastMove:     lastMove,
				LastMoveDiff: lastMoveDiff,
				RSSI:         int(rssi),
			},
		},
		LastTime:  now,
		FirstTime: now,
	})
}

func checkOldBlueDevice() {
	ids := []string{}
	list := []*datastore.BlueDeviceEnt{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	delOldRandam := time.Now().AddDate(0, 0, -1).UnixNano()
	datastore.ForEachBlueDevice(func(e *datastore.BlueDeviceEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		} else if e.LastTime < delOldRandam && strings.Contains(e.AddressType, " Random") {
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
		datastore.DeleteReport("blueDevice", ids)
	}
}

func checkOldEnvMonitor() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachEnvMonitor(func(e *datastore.EnvMonitorEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("envMonitor", ids)
	}
}

func checkOldMotionSensor() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachMotionSensor(func(e *datastore.MotionSensorEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("motionSensor", ids)
	}
}
