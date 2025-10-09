package polling

// LOG監視ポーリング処理

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func doPollingArpLog(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "stats":
		doPollingArpLogStats(pe)
	case "check":
		doPollingArpLogCheck(pe)
	default:
		doPollingArpLogCount(pe)
	}
}

func doPollingArpLogCount(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	filter := pe.Filter
	script := pe.Script
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("arplog", pe, fmt.Errorf("invalid filter"))
			return
		}
	}
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	count := 0
	datastore.ForEachLog(st, et, pe.Type, func(l *datastore.LogEnt) bool {
		msg := l.Log
		if regexFilter != nil && !regexFilter.Match([]byte(msg)) {
			return true
		}
		count++
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	delete(pe.Result, "error")
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	vm.Set("count", count)
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("arplog", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingArpLogCheck(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	filter := pe.Filter
	extractor := pe.Extractor
	script := pe.Script
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("arplog", pe, err)
			return
		}
	}
	if extractor == "" {
		setPollingError("arplog", pe, fmt.Errorf("no extractor"))
		return
	}
	grokEnt := datastore.GetGrokEnt(extractor)
	if grokEnt == nil {
		setPollingError("arplog", pe, fmt.Errorf("no extractor"))
		return
	}
	grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		setPollingError("arplog", pe, err)
		return
	}
	if err = grokExtractor.AddPattern(extractor, grokEnt.Pat); err != nil {
		setPollingError("arplog", pe, err)
		return
	}
	grokCap := fmt.Sprintf("%%{%s}", extractor)
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	var lastError error
	failed := false
	datastore.ForEachLog(st, et, pe.Type, func(l *datastore.LogEnt) bool {
		msg := l.Log
		if regexFilter != nil && !regexFilter.Match([]byte(msg)) {
			return true
		}
		values, err := grokExtractor.Parse(grokCap, msg)
		if err != nil {
			return true
		}
		for k, v := range values {
			vm.Set(k, v)
			pe.Result[k] = v
		}
		if script != "" {
			value, err := vm.Run(script)
			if err == nil {
				if ok, _ := value.ToBoolean(); !ok {
					failed = true
					return false
				}
			} else {
				lastError = err
				return false
			}
		}
		return true
	})
	pe.Result["lastTime"] = et
	if lastError != nil {
		setPollingError("arplog", pe, lastError)
		return
	}
	if failed {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "normal")
}

func doPollingArpLogStats(pe *datastore.PollingEnt) {
	script := pe.Script
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	count := 0
	patternMap := make(map[string]int)
	stateMap := make(map[string]int)
	ipMap := make(map[string]int)
	macMap := make(map[string]int)
	datastore.ForEachLog(st, et, pe.Type, func(l *datastore.LogEnt) bool {
		a := strings.Split(l.Log, ",")
		if len(a) < 3 {
			return true
		}
		state := a[0]
		ip := a[1]
		mac := a[2]
		patternMap[ip+mac]++
		stateMap[state]++
		ipMap[ip]++
		macMap[mac]++
		count++
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	pe.Result["pattern"] = float64(len(patternMap))
	pe.Result["states"] = float64(len(stateMap))
	pe.Result["IPs"] = float64(len(ipMap))
	pe.Result["MACs"] = float64(len(macMap))
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("arplog", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}
