package report

import (
	"fmt"
	"strconv"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/notify"
)

func UpdateSensor(h, t string, r int) {
	id := fmt.Sprintf("%s:%s:", h, t)
	now := time.Now().UnixNano()
	e := datastore.GetSensor(id)
	if e != nil {
		e.Total += int64(r)
		e.Send++
		e.LastTime = now
		return
	}
	datastore.AddSensor(&datastore.SensorEnt{
		ID:        id,
		Host:      h,
		Type:      t,
		State:     "normal",
		Total:     int64(r),
		Send:      1,
		LastTime:  now,
		FirstTime: now,
	})
}

func checkStats(h, t string, m map[string]string) {
	now := time.Now().UnixNano()
	param := m["param"]
	count := getNumberFromTWLog(m["count"])
	send := getNumberFromTWLog(m["send"])
	if send < 1 {
		send = 1
	}
	total := getNumberFromTWLog(m["total"])
	ps := getFloatFromTWLog(m["ps"])
	id := makeID(h + ":" + t + ":" + param)
	e := datastore.GetSensor(id)
	if e != nil {
		e.Total += count
		e.Send += send
		e.Stats = append(e.Stats, datastore.SensorStatsEnt{
			Time:  now,
			Total: total,
			Count: count,
			Send:  send,
			PS:    ps,
		})
		for len(e.Stats) > 1440*2 {
			e.Stats = e.Stats[1:]
		}
		e.LastTime = time.Now().UnixNano()
		return
	}
	datastore.AddSensor(&datastore.SensorEnt{
		ID:    id,
		Host:  h,
		Type:  t,
		Param: param,
		Total: count,
		Send:  send,
		Stats: []datastore.SensorStatsEnt{
			{
				Time:  now,
				Total: total,
				Count: count,
				Send:  send,
				PS:    ps,
			},
		},
		LastTime:  now,
		FirstTime: now,
	})
}

/*
 type=Monitor,
 cpu=32.927,load=2.650,mem=5.544,recv=240157990,sent=25535931,rxSpeed=32.021,txSpeed=3.405,process=1
*/

func checkMonitor(h, t string, m map[string]string) {
	now := time.Now().UnixNano()
	param := m["param"]
	cpu := getFloatFromTWLog(m["cpu"])
	load := getFloatFromTWLog(m["load"])
	mem := getFloatFromTWLog(m["mem"])
	txSpeed := getFloatFromTWLog(m["txSpeed"])
	rxSpeed := getFloatFromTWLog(m["rxSpeed"])
	sent := getNumberFromTWLog(m["sent"])
	recv := getNumberFromTWLog(m["recv"])
	proc := getNumberFromTWLog(m["process"])
	id := makeID(h + ":" + t + ":" + param)
	e := datastore.GetSensor(id)
	if e != nil {
		e.Monitors = append(e.Monitors, datastore.SensorMonitorEnt{
			Time:    now,
			CPU:     cpu,
			Mem:     mem,
			Load:    load,
			TxSpeed: txSpeed,
			RxSpeed: rxSpeed,
			Recv:    recv,
			Sent:    sent,
			Process: proc,
		})
		for len(e.Monitors) > 1440*2 {
			e.Monitors = e.Monitors[1:]
		}
		e.LastTime = time.Now().UnixNano()
		return
	}
}

func setSensorState() {
	to := datastore.ReportConf.SensorTimeout
	warn := time.Now().Add(-time.Duration(to) * time.Hour).UnixNano()
	low := time.Now().Add(-time.Duration(to*6) * time.Hour).UnixNano()
	high := time.Now().Add(-time.Duration(to*24) * time.Hour).UnixNano()
	now := time.Now().UnixNano()
	datastore.ForEachSensors(func(s *datastore.SensorEnt) bool {
		oldState := s.State
		if s.LastTime < high {
			s.State = "high"
		} else if s.LastTime < low {
			s.State = "low"
		} else if s.LastTime < warn {
			s.State = "warn"
		} else {
			s.State = "normal"
		}
		if s.Ignore {
			s.State = "unknown"
		}
		if oldState != s.State {
			l := &datastore.EventLogEnt{
				Type:     "sensor",
				Level:    s.State,
				NodeID:   "",
				NodeName: s.Host,
				Event:    "センサー状態変化:" + s.Host + "," + s.Type + "," + s.Param,
			}
			datastore.AddEventLog(l)
			notify.SendNotifyChat(l)
		}
		if s.Type == "netflow" || s.Type == "ipfix" || s.Type == "syslog" {
			count := s.Total
			send := s.Send
			if len(s.Stats) > 0 {
				count -= s.Stats[len(s.Stats)-1].Total
				send -= s.Stats[len(s.Stats)-1].LastSend
			}
			s.Stats = append(s.Stats, datastore.SensorStatsEnt{
				Time:     now,
				Total:    s.Total,
				Count:    count,
				Send:     send,
				LastSend: s.Send,
				PS:       float64(count) / (5.0 * 60.0),
			})
			for len(s.Stats) > 1440*2 {
				s.Stats = s.Stats[1:]
			}
		}
		return true
	})
}

func getNumberFromTWLog(s string) int64 {
	if c, err := strconv.ParseInt(s, 10, 64); err == nil {
		return c
	}
	return 0
}

func getFloatFromTWLog(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}

func getTimeFromTWLog(s string) int64 {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UnixNano()
	}
	return time.Now().UnixNano()
}
