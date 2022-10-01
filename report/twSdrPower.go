package report

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func ReportTWSdrPower(l map[string]interface{}) {
	twSdrPowerCh <- l
}

func checkTWSdrPowerReport(l map[string]interface{}) {
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
	case "Power":
		checkSdrPowerReport(h, m)
	case "Stats":
		checkStats(h, "twSdrPower", m)
	case "Monitor":
		checkMonitor(h, "twSdrPower", m)
	default:
		log.Printf("twSdrPower unknown type=%v m=%v", t, m)
	}
}

var sdrPowerLog = []*datastore.SdrPowerEnt{}

// type=Power,id=632a2449,freq=24000000,dbm=-16.443
func checkSdrPowerReport(h string, m map[string]string) {
	_, ok := m["freq"]
	if !ok {
		return
	}
	ts := getNumberFromTWLogHex(m["id"])
	freq := getNumberFromTWLog(m["freq"])
	dbm := getFloatFromTWLog(m["dbm"])
	sdrPowerLog = append(sdrPowerLog, &datastore.SdrPowerEnt{
		Host: h,
		Time: ts,
		Freq: freq,
		Dbm:  dbm,
	})
}

func saveSdrPowerReport() {
	if len(sdrPowerLog) > 0 {
		st := time.Now()
		datastore.AddSdrPower(sdrPowerLog)
		log.Printf("saveSdrPowerReport len=%d dur=%v", len(sdrPowerLog), time.Since(st))
		sdrPowerLog = []*datastore.SdrPowerEnt{}
	}
}

func checkOldEnvSdrPower(delOld int64) {
	ids := []string{}
	datastore.ForEachSdrPower(0, "", func(e *datastore.SdrPowerEnt) bool {
		if e.Time < delOld {
			ids = append(ids, fmt.Sprintf("%016x:%s:%016x", e.Time, e.Host, e.Freq))
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("sdrPower", ids)
	}
}
