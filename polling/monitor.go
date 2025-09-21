package polling

// モニターのポーリング

import (
	"fmt"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingMonitor(pe *datastore.PollingEnt) bool {
	var err error
	if len(datastore.MonitorDataes) < 1 {
		setPollingError("monitor", pe, fmt.Errorf("no monitor data"))
		return false
	}
	st := int64(0)
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	m := datastore.MonitorDataes[len(datastore.MonitorDataes)-1]
	if m.At < st {
		return false
	}
	pe.Result["lastTime"] = m.At
	pe.Result["cpu"] = m.CPU
	pe.Result["mem"] = m.Mem
	pe.Result["disk"] = m.Disk
	pe.Result["load"] = m.Load
	pe.Result["net"] = m.Net
	pe.Result["conn"] = m.Conn
	pe.Result["myCpu"] = m.MyCPU
	pe.Result["myMem"] = m.MyMem
	pe.Result["goRoutine"] = m.NumGoroutine
	pe.Result["heap"] = m.HeapAlloc
	pe.Result["swap"] = m.Swap
	script := pe.Script
	if script == "" {
		setPollingState(pe, "normal")
		return true
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("monitor", pe, err)
		return false
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	return true
}
