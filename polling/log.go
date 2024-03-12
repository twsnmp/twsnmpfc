package polling

// LOG監視ポーリング処理

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

func doPollingLog(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	filter := pe.Filter
	params := strings.TrimSpace(pe.Params)
	extractor := pe.Extractor
	script := pe.Script
	mode := pe.Mode
	if params != "" && mode == "count" {
		switch pe.Type {
		case "trap":
			filter = `From.*` + regexp.QuoteMeta(params) + `:[\s\S\n]*` + filter
		case "syslog":
			filter = filter + `[\s\S\n]*hostname\s+` + regexp.QuoteMeta(params) + `\s+`
		}
	}
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
			return
		}
	}
	grokCap := ""
	grokOk := ""
	server := ""
	if extractor != "" {
		if mode == "user" {
			n := datastore.GetNode(pe.NodeID)
			server = n.IP
		} else if script == "" && mode != "device" && mode != "flow" {
			setPollingError("log", pe, fmt.Errorf("no script"))
			return
		}
		grokEnt := datastore.GetGrokEnt(extractor)
		if grokEnt == nil {
			setPollingError("log", pe, fmt.Errorf("no extractor pattern"))
			return
		}
		grokOk = grokEnt.Ok
		grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		if err != nil {
			setPollingError("log", pe, fmt.Errorf("no extractor pattern"))
			return
		}
		if err = grokExtractor.AddPattern(extractor, grokEnt.Pat); err != nil {
			setPollingError("log", pe, fmt.Errorf("no extractor pattern"))
			return
		}
		grokCap = fmt.Sprintf("%%{%s}", extractor)
	}
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	vm := otto.New()
	addJavaScriptFunctions(pe, vm)
	count := 0
	okCount := 0
	failed := false
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
		if grokExtractor != nil {
			values, err := grokExtractor.Parse(grokCap, msg)
			if err != nil {
				return true
			}
			if mode == "device" {
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
			}
			if mode == "user" {
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
			}
			if mode == "flow" {
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
			}
			count++
			for k, v := range values {
				vm.Set(k, v)
				pe.Result[k] = v
			}
			value, err := vm.Run(script)
			if err == nil {
				if ok, _ := value.ToBoolean(); !ok {
					failed = true
					setPollingState(pe, pe.Level)
					return false
				}
			} else {
				failed = true
				setPollingError("log", pe, fmt.Errorf("invalid script"))
				return false
			}
		} else {
			count++
		}
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	if extractor != "" {
		if mode == "user" {
			if count > 0 {
				pe.Result["rate"] = 100.0 * (float64(okCount) / float64(count))
			} else {
				pe.Result["rate"] = 0.0
			}
			pe.Result["ok"] = float64(okCount)
		}
		if !failed {
			setPollingState(pe, "normal")
		}
		return
	}
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm.Set("count", count)
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("log", pe, fmt.Errorf("invalid script err=%v", err))
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingSyslog(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "pri":
		doPollingSyslogPri(pe)
	case "state":
		doPollingSyslogState(pe)
	default:
		doPollingLog(pe)
	}
}

func doPollingSyslogPri(pe *datastore.PollingEnt) bool {
	var err error
	var regexFilter *regexp.Regexp
	filter := pe.Filter
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
			return false
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
	return true
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

func doPollingSyslogState(pe *datastore.PollingEnt) bool {
	var err error
	var ngFilter *regexp.Regexp
	var okFilter *regexp.Regexp
	if pe.Filter == "" || pe.Params == "" {
		setPollingError("log", pe, fmt.Errorf("no polling filter"))
		return false
	}
	if ngFilter, err = regexp.Compile(pe.Filter); err != nil {
		setPollingError("log", pe, fmt.Errorf("invalid ng filter"))
		return false
	}
	if okFilter, err = regexp.Compile(pe.Params); err != nil {
		setPollingError("log", pe, fmt.Errorf("invalid ok filter"))
		return false
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
			// 現状維持
			return true
		}
	} else {
		if ngTime < okTime {
			// OKが後の場合は正常（NGがない場合も含む）
			setPollingState(pe, "normal")
			return true
		}
	}
	//それ以外はすべてNG
	setPollingState(pe, pe.Level)
	return true
}
