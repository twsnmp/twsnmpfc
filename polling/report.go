package polling

// LOG監視ポーリング処理

import (
	"fmt"
	"math"

	"github.com/montanaflynn/stats"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingReport(pe *datastore.PollingEnt) {
	scores := []float64{}
	switch pe.Mode {
	case "device":
		datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
			if d.ValidScore {
				scores = append(scores, d.Score)
			}
			return true
		})
	case "user":
		datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
			if u.ValidScore {
				scores = append(scores, u.Score)
			}
			return true
		})
	case "server":
		datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
			if s.ValidScore {
				scores = append(scores, s.Score)
			}
			return true
		})
	case "flow":
		datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
			if f.ValidScore {
				scores = append(scores, f.Score)
			}
			return true
		})
	default:
		setPollingError("report", pe, fmt.Errorf("invalid report mode"))
		return
	}
	if len(scores) < 1 {
		setPollingError("report", pe, fmt.Errorf("no report data"))
		return
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	pe.Result = make(map[string]interface{})
	pe.Result["count"] = float64(len(scores))
	pe.Result["min"], _ = stats.Min(scores)
	pe.Result["max"], _ = stats.Max(scores)
	pe.Result["mean"], _ = stats.Mean(scores)
	pe.Result["mode"], _ = stats.Mode(scores)
	pe.Result["median"], _ = stats.Median(scores)
	pe.Result["stddev"], _ = stats.StandardDeviation(scores)
	for k, v := range pe.Result {
		if v == math.NaN() {
			pe.Result[k] = 0.0
		}
	}
	script := pe.Script
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm.Set("count", pe.Result["count"])
	vm.Set("max", pe.Result["max"])
	vm.Set("min", pe.Result["min"])
	vm.Set("mean", pe.Result["mean"])
	vm.Set("mode", pe.Result["mode"])
	vm.Set("median", pe.Result["median"])
	vm.Set("stddev", pe.Result["stddev"])
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("report", pe, fmt.Errorf("invalid script"))
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}
