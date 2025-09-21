package polling

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func doPollingSnmpTrap(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "count":
		doPollingSnmpTrapCount(pe)
	default:
		doPollingSnmpTrapCheck(pe)
	}
}

func doPollingSnmpTrapCount(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	filter := pe.Filter
	params := strings.TrimSpace(pe.Params)
	script := pe.Script
	if params != "" {
		filter = `From.*` + regexp.QuoteMeta(params) + `:[\s\S\n]*` + filter
	}
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("trap", pe, err)
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
	keys := []string{}
	datastore.ForEachLog(st, et, pe.Type, func(l *datastore.LogEnt) bool {
		msg := ""
		if pe.Type == "arplog" {
			msg = l.Log
		} else {
			var sl = make(map[string]interface{})
			if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
				return true
			}
			if len(keys) < 1 {
				for k := range sl {
					keys = append(keys, k)
				}
				sort.Strings(keys)
			}
			for _, k := range keys {
				if v, ok := sl[k]; ok {
					msg += k + "\t" + fmt.Sprintf("%v", v) + "\n"
				}
			}
		}
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
		setPollingError("trap", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingSnmpTrapCheck(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	filter := pe.Filter
	params := strings.TrimSpace(pe.Params)
	extractor := pe.Extractor
	script := pe.Script
	if params != "" {
		filter = `From.*` + regexp.QuoteMeta(params) + `:[\s\S\n]*` + filter
	}
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("snmptrap", pe, err)
			return
		}
	}
	if extractor == "" {
		setPollingError("snmptrap", pe, fmt.Errorf("no extractor"))
		return
	}
	grokEnt := datastore.GetGrokEnt(extractor)
	if grokEnt == nil {
		setPollingError("snmptrap", pe, fmt.Errorf("no extractor"))
		return
	}
	grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		setPollingError("snmptrap", pe, err)
		return
	}
	if err = grokExtractor.AddPattern(extractor, grokEnt.Pat); err != nil {
		setPollingError("snmptrap", pe, err)
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
	failed := false
	var lastError error
	keys := []string{}
	datastore.ForEachLog(st, et, pe.Type, func(l *datastore.LogEnt) bool {
		msg := ""
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		if len(keys) < 1 {
			for k := range sl {
				keys = append(keys, k)
			}
			sort.Strings(keys)
		}
		for _, k := range keys {
			if v, ok := sl[k]; ok {
				msg += k + "\t" + fmt.Sprintf("%v", v) + "\n"
			}
		}
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
		setPollingError("snmptrap", pe, lastError)
	}
	if failed {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "normal")
}
