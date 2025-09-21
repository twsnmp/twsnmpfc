package polling

// LOG監視ポーリング処理

import (
	"fmt"
	"regexp"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func doPollingArpLog(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "count":
		doPollingArpLogCount(pe)
	default:
		doPollingArpLogCheck(pe)
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
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
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
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
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
