package polling

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
	"github.com/vjeantet/grok"
)

func doPollingSyslog(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "stats":
		doPollingSyslogStats(pe)
	case "pri":
		doPollingSyslogPri(pe)
	case "state":
		doPollingSyslogState(pe)
	case "user":
		doPollingSyslogUser(pe)
	case "device":
		doPollingSyslogDevice(pe)
	case "flow":
		doPollingSyslogFlow(pe)
	case "count":
		doPollingSyslogCount(pe)
	default:
		doPollingSyslogCheck(pe)
	}
}

func doPollingSyslogCount(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	filter := pe.Filter
	params := strings.TrimSpace(pe.Params)
	script := pe.Script
	if params != "" {
		filter = filter + `[\s\S\n]*hostname\s+` + regexp.QuoteMeta(params) + `\s+`
	}
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("syslog", pe, err)
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
		count++
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
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
		setPollingError("syslog", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingSyslogUser(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	filter := pe.Filter
	params := strings.TrimSpace(pe.Params)
	extractor := pe.Extractor
	script := pe.Script
	if params != "" {
		filter = filter + `[\s\S\n]*hostname\s+` + regexp.QuoteMeta(params) + `\s+`
	}
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("syslog", pe, fmt.Errorf("invalid filter for user"))
			return
		}
	}
	grokCap := ""
	grokOk := ""
	server := ""
	if extractor == "" {
		setPollingError("syslog", pe, fmt.Errorf("no extractor for user"))
		return
	}
	n := datastore.GetNode(pe.NodeID)
	server = n.IP
	grokEnt := datastore.GetGrokEnt(extractor)
	if grokEnt == nil {
		setPollingError("syslog", pe, fmt.Errorf("no extractor for user"))
		return
	}
	grokOk = grokEnt.Ok
	grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		setPollingError("syslog", pe, err)
		return
	}
	if err = grokExtractor.AddPattern(extractor, grokEnt.Pat); err != nil {
		setPollingError("log", pe, err)
		return
	}
	grokCap = fmt.Sprintf("%%{%s}", extractor)
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
	okCount := 0
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
		stat, ok := values["stat"]
		if !ok {
			return true
		}
		user, ok := values["user"]
		if !ok {
			return true
		}
		client := values["client"]
		ok = grokOk == stat
		count++
		if ok {
			okCount++
		}
		report.ReportUser(user, server, client, ok, l.Time)
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	if count > 0 {
		pe.Result["rate"] = 100.0 * (float64(okCount) / float64(count))
	} else {
		pe.Result["rate"] = 0.0
	}
	pe.Result["ok"] = float64(okCount)
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm.Set("rate", pe.Result["rate"])
	vm.Set("ok", pe.Result["ok"])
	vm.Set("count", count)
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("log", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingSyslogDevice(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	filter := pe.Filter
	params := strings.TrimSpace(pe.Params)
	extractor := pe.Extractor
	script := pe.Script
	if params != "" {
		filter = filter + `[\s\S\n]*hostname\s+` + regexp.QuoteMeta(params) + `\s+`
	}
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("syslog", pe, fmt.Errorf("invalid filter for device"))
			return
		}
	}
	grokCap := ""
	if extractor == "" {
		setPollingError("syslog", pe, fmt.Errorf("no extractor for device"))
		return
	}
	grokEnt := datastore.GetGrokEnt(extractor)
	if grokEnt == nil {
		setPollingError("syslog", pe, fmt.Errorf("no extractor fo device"))
		return
	}
	grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		setPollingError("syslog", pe, err)
		return
	}
	if err = grokExtractor.AddPattern(extractor, grokEnt.Pat); err != nil {
		setPollingError("log", pe, err)
		return
	}
	grokCap = fmt.Sprintf("%%{%s}", extractor)
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
		mac, ok := values["mac"]
		if !ok {
			return true
		}
		ip, ok := values["ip"]
		if !ok {
			return true
		}
		count++
		report.ReportDevice(mac, ip, l.Time)
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
		setPollingError("log", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingSyslogFlow(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	filter := pe.Filter
	params := strings.TrimSpace(pe.Params)
	extractor := pe.Extractor
	script := pe.Script
	if params != "" {
		filter = filter + `[\s\S\n]*hostname\s+` + regexp.QuoteMeta(params) + `\s+`
	}
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("syslog", pe, err)
			return
		}
	}
	grokCap := ""
	grokEnt := datastore.GetGrokEnt(extractor)
	if grokEnt == nil {
		setPollingError("syslog", pe, fmt.Errorf("no extractor for flow"))
		return
	}
	grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		setPollingError("syslog", pe, err)
		return
	}
	if err = grokExtractor.AddPattern(extractor, grokEnt.Pat); err != nil {
		setPollingError("syslog", pe, err)
		return
	}
	grokCap = fmt.Sprintf("%%{%s}", extractor)

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
		src, ok := values["src"]
		if !ok {
			return true
		}
		dst, ok := values["dst"]
		if !ok {
			return true
		}
		sport, ok := values["sport"]
		if !ok {
			return true
		}
		dport, ok := values["dport"]
		if !ok {
			return true
		}
		prot, ok := values["prot"]
		if !ok {
			return true
		}
		nBytes := 0
		for _, b := range []string{"bytes", "sent", "rcvd"} {
			bytes, ok := values[b]
			if ok {
				nB, _ := strconv.Atoi(bytes)
				nBytes += nB
			}
		}
		nPackets := 0
		for _, b := range []string{"spkt", "rpkt"} {
			pkts, ok := values[b]
			if ok {
				nP, _ := strconv.Atoi(pkts)
				nPackets += nP
			}
		}
		nProt := getProt(prot)
		nSPort, _ := strconv.Atoi(sport)
		nDPort, _ := strconv.Atoi(dport)
		report.ReportFlow(src, nSPort, dst, nDPort, nProt, int64(nPackets), int64(nBytes), l.Time)
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
		setPollingError("syslog", pe, fmt.Errorf("invalid script err=%v", err))
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingSyslogCheck(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	filter := pe.Filter
	params := strings.TrimSpace(pe.Params)
	extractor := pe.Extractor
	script := pe.Script
	if params != "" {
		filter = filter + `[\s\S\n]*hostname\s+` + regexp.QuoteMeta(params) + `\s+`
	}
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("syslog", pe, err)
			return
		}
	}
	grokEnt := datastore.GetGrokEnt(extractor)
	if grokEnt == nil {
		setPollingError("syslog", pe, fmt.Errorf("no extractor for check"))
		return
	}
	grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		setPollingError("syslog", pe, err)
		return
	}
	if err = grokExtractor.AddPattern(extractor, grokEnt.Pat); err != nil {
		setPollingError("syslog", pe, err)
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
		setPollingError("syslog", pe, lastError)
		return
	}
	if failed {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "normal")

}

func doPollingSyslogPri(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	filter := pe.Filter
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("syslog", pe, err)
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
	priMap := make(map[float64]int)
	datastore.ForEachLog(st, et, "syslog", func(l *datastore.LogEnt) bool {
		msg := ""
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		for k, v := range sl {
			msg += k + "=" + fmt.Sprintf("%v", v) + "\t"
		}
		if regexFilter != nil && !regexFilter.Match([]byte(msg)) {
			return true
		}
		count++
		if v, ok := sl["priority"]; ok {
			if pri, ok := v.(float64); ok {
				priMap[pri]++
			}
		}
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	for pri, c := range priMap {
		pe.Result[fmt.Sprintf("pri_%d", int(pri))] = float64(c)
	}
	setPollingState(pe, "normal")
}

func getProt(p string) int {
	if strings.Contains(p, "tcp") {
		return 6
	}
	if strings.Contains(p, "udp") {
		return 17
	}
	if strings.Contains(p, "icmp") {
		return 1
	}
	return 0
}

func doPollingSyslogState(pe *datastore.PollingEnt) {
	var err error
	var ngFilter *regexp.Regexp
	var okFilter *regexp.Regexp
	if pe.Filter == "" || pe.Params == "" {
		setPollingError("syslog", pe, fmt.Errorf("no filter fo state"))
		return
	}
	if ngFilter, err = regexp.Compile(pe.Filter); err != nil {
		setPollingError("syslog", pe, err)
		return
	}
	if okFilter, err = regexp.Compile(pe.Params); err != nil {
		setPollingError("syslog", pe, err)
		return
	}
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	var okTime int64
	var ngTime int64
	datastore.ForEachLog(st, et, "syslog", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		msg := ""
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		for k, v := range sl {
			msg += k + "=" + fmt.Sprintf("%v", v) + "\t"
		}
		if okFilter.Match([]byte(msg)) {
			okTime = l.Time
			return true
		}
		if ngFilter.Match([]byte(msg)) {
			ngTime = l.Time
			return true
		}
		return true
	})
	pe.Result["lastTime"] = et
	if okTime == 0 {
		if ngTime == 0 {
			// どちらもない場合
			if pe.State == "unknown" {
				// 正常とする
				setPollingState(pe, "normal")
			}
			return
		}
	} else {
		if ngTime < okTime {
			// OKが後の場合は正常（NGがない場合も含む）
			setPollingState(pe, "normal")
			return
		}
	}
	//それ以外はすべてNG
	setPollingState(pe, pe.Level)
}

func doPollingSyslogStats(pe *datastore.PollingEnt) {
	script := pe.Script
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	count := 0
	normal := 0
	warns := 0
	errors := 0
	patternMap := make(map[string]int)
	errorPatternMap := make(map[string]int)
	datastore.ForEachLog(st, et, pe.Type, func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		sv, ok := sl["severity"].(float64)
		if !ok {
			sv = 5
		}
		host, ok := sl["hostname"].(string)
		if !ok {
			host = ""
		}
		var message string
		tag, ok := sl["tag"].(string)
		if !ok {
			if tag, ok = sl["app_name"].(string); !ok {
				tag = ""
			}
			for i, k := range []string{"proc_id", "msg_id", "message", "structured_data"} {
				if m, ok := sl[k].(string); ok && m != "" {
					if i > 0 {
						message += " "
					}
					message += m
				}
			}
		} else {
			message, ok = sl["content"].(string)
			if !ok {
				message = ""
			}
		}
		msg := host + " " + tag + " " + message
		nl := normalizeSyslog(msg)
		patternMap[nl]++
		switch {
		case sv < 4:
			errorPatternMap[nl]++
			errors++
		case sv == 4.0:
			warns++
		default:
			normal++
		}
		count++
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	pe.Result["error"] = float64(errors)
	pe.Result["warn"] = float64(warns)
	pe.Result["normal"] = float64(normal)
	pe.Result["patterns"] = float64(len(patternMap))
	pe.Result["errorPatterns"] = float64(len(errorPatternMap))
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
		setPollingError("syslog", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

var regNum = regexp.MustCompile(`\b-?\d+(\.\d+)?\b`)
var regUUDI = regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`)
var regEmail = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
var regIP = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
var regMAC = regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}[:-]){5}(?:[0-9a-fA-F]{2})\b`)

func normalizeSyslog(msg string) string {
	normalized := msg
	normalized = regUUDI.ReplaceAllString(normalized, "#UUID#")
	normalized = regEmail.ReplaceAllString(normalized, "#EMAIL#")
	normalized = regIP.ReplaceAllString(normalized, "#IP#")
	normalized = regMAC.ReplaceAllString(normalized, "#MAC#")
	normalized = regNum.ReplaceAllString(normalized, "#NUM#")
	return normalized
}
