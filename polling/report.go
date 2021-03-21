package polling

// LOG監視ポーリング処理

import (
	"fmt"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingReport(pe *datastore.PollingEnt) {
	var scoreTotal float64
	var scoreMin float64
	scoreMin = 99999.0
	count := 0
	switch pe.Mode {
	case "device":
		datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
			scoreTotal += d.Score
			if d.Score > 0.0 && d.Score < scoreMin {
				scoreMin = d.Score
			}
			count++
			return true
		})
	case "user":
		datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
			scoreTotal += u.Score
			if u.Score > 0.0 && u.Score < scoreMin {
				scoreMin = u.Score
			}
			count++
			return true
		})
	case "server":
		datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
			scoreTotal += s.Score
			if s.Score > 0.0 && s.Score < scoreMin {
				scoreMin = s.Score
			}
			count++
			return true
		})
	case "flow":
		datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
			scoreTotal += f.Score
			if f.Score > 0.0 && f.Score < scoreMin {
				scoreMin = f.Score
			}
			count++
			return true
		})
	default:
		setPollingError("report", pe, fmt.Errorf("invalid remote mode"))
		return
	}
	var scoreAvg float64
	if count > 0 {
		scoreAvg = scoreTotal / float64(count)
	}
	pe.Result["count"] = float64(count)
	pe.Result["scoreMin"] = scoreMin
	pe.Result["scoreAvg"] = scoreAvg
	script := pe.Script
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm := otto.New()
	vm.Set("count", count)
	vm.Set("scoreMin", scoreMin)
	vm.Set("scoreAvg", scoreAvg)
	value, err := vm.Run(script)
	if err == nil {
		setPollingError("log", pe, fmt.Errorf("invalid script"))
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}
