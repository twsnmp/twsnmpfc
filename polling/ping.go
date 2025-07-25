package polling

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
)

func doPollingPing(pe *datastore.PollingEnt) {
	if pe.Mode == "line" {
		doPollingCheckLineCond(pe)
		return
	}
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("ping", pe, fmt.Errorf("node not found"))
		return
	}
	size := 64
	ttl := 0
	if pe.Params != "" {
		if strings.Contains(pe.Params, "=") {
			a := strings.Split(pe.Params, ",")
			for _, s := range a {
				b := strings.SplitN(s, "=", 2)
				if len(b) == 2 {
					if i, err := strconv.Atoi(b[1]); err == nil {
						if b[0] == "ttl" && i > 0 && i < 256 {
							ttl = i
						} else if b[0] == "size" && i >= 0 && i < 3000 {
							size = i
						}
					}
				}
			}
		} else {
			if i, err := strconv.Atoi(pe.Params); err == nil {
				size = i
			}
		}
	}
	r := ping.DoPing(n.IP, pe.Timeout, pe.Retry, size, ttl)
	if r.Stat == ping.PingOK {
		pe.Result["rtt"] = float64(r.Time)
		pe.Result["ttl"] = float64(r.RecvTTL)
		delete(pe.Result, "error")
		setPollingState(pe, "normal")
	} else {
		pe.Result["rtt"] = 0.0
		pe.Result["ttl"] = 0.0
		pe.Result["error"] = fmt.Sprintf("%v", r.Error)
		setPollingState(pe, pe.Level)
	}
}

func doPollingCheckLineCond(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("ping", pe, fmt.Errorf("node not found"))
		return
	}
	lastError := ""
	speed := []float64{}
	rtt := []float64{}
	fail := 0
	ttl := 0
	for i := 0; i < 20; i++ {
		r64 := ping.DoPing(n.IP, pe.Timeout, pe.Retry, 64, 0)
		if r64.Stat != ping.PingOK {
			lastError = fmt.Sprintf("%v", r64.Error)
			fail += 1
			continue
		}
		r1364 := ping.DoPing(n.IP, pe.Timeout, pe.Retry, 1364, 0)
		if r1364.Stat != ping.PingOK {
			lastError = fmt.Sprintf("%v", r1364.Error)
			fail += 1
			continue
		}
		if r64.Time == r1364.Time {
			fail += 1
			continue
		}
		ttl = r64.RecvTTL
		a := float64(64.0-1364.0) / float64(r64.Time-r1364.Time)
		b := float64(r64.Time) - a*float64(64.0)
		s := a * (8.0 * 1000.0) //Mbps
		if s > 0.0 && s < 1000.0 && b > 0.0 {
			rtt = append(rtt, b)
			speed = append(speed, s)
			if len(speed) >= 5 {
				break
			}
		} else {
			fail += 1
		}
	}
	if len(speed) < 3 {
		pe.Result["error"] = lastError
		pe.Result["rtt"] = 0.0
		pe.Result["ttl"] = ttl
		pe.Result["rtt_cv"] = 0.0
		pe.Result["speed"] = 0.0
		pe.Result["speed_cv"] = 0.0
		pe.Result["fail"] = float64(fail)
		setPollingState(pe, pe.Level)
		return
	}
	// 5回の測定から平均値と変動係数を計算
	rm, rcv := calcMeanCV(rtt)
	pe.Result["rtt"] = rm
	pe.Result["rtt_cv"] = rcv
	pe.Result["ttl"] = ttl
	sm, scv := calcMeanCV(speed)
	pe.Result["speed"] = sm
	pe.Result["speed_cv"] = scv
	pe.Result["fail"] = float64(fail)
	delete(pe.Result, "error")
	setPollingState(pe, "normal")
}

func calcMeanCV(a []float64) (float64, float64) {
	if len(a) < 1 {
		return 0.0, 0.0
	}
	n := float64(len(a))
	m := float64(0.0)
	for _, d := range a {
		m += d
	}
	m /= n
	if m == 0.0 {
		return 0.0, 0.0
	}
	v := float64(0.0)
	for _, d := range a {
		v += (d - m) * (d - m)
	}
	v /= n
	sigma := math.Sqrt(v)
	return m, sigma / m
}
