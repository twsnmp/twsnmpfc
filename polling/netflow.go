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
	"github.com/vjeantet/grok"
)

func doPollingNetflow(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "count":
		doPollingNetflowCount(pe)
	case "traffic":
		doPollingNetFlowTraffic(pe)
	default:
		doPollingNetflowCheck(pe)
	}
}

func doPollingNetflowCount(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	filter := pe.Filter
	script := pe.Script
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("netflow", pe, err)
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
		setPollingError("netflow", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingNetflowCheck(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	filter := pe.Filter
	extractor := pe.Extractor
	script := pe.Script
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("netflow", pe, err)
			return
		}
	}
	if extractor == "" {
		setPollingError("netflow", pe, fmt.Errorf("no extractor"))
		return
	}
	grokEnt := datastore.GetGrokEnt(extractor)
	if grokEnt == nil {
		setPollingError("netflow", pe, fmt.Errorf("no extractor"))
		return
	}
	grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		setPollingError("netflow", pe, fmt.Errorf("no extractor"))
		return
	}
	if err = grokExtractor.AddPattern(extractor, grokEnt.Pat); err != nil {
		setPollingError("netflow", pe, err)
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
		setPollingError("netflow", pe, err)
		return
	}
	if failed {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "normal")
}

func doPollingNetFlowTraffic(pe *datastore.PollingEnt) {
	var err error
	var filterSrc *regexp.Regexp
	var filterIP *regexp.Regexp
	var filterDst *regexp.Regexp
	var filterPort int
	var filterProt int
	if pe.Filter != "" {
		fs := strings.Split(pe.Filter, ",")
		for _, fe := range fs {
			f := strings.Split(fe, "=")
			if len(f) != 2 {
				continue
			}
			switch f[0] {
			case "src":
				filterSrc = makeRegexpFilter(f[1])
			case "dst":
				filterDst = makeRegexpFilter(f[1])
			case "ip":
				filterIP = makeRegexpFilter(f[1])
			case "port":
				filterPort, _ = strconv.Atoi(f[1])
			case "prot":
				filterProt = getProt(f[1])
			default:
				setPollingError("netflow", pe, fmt.Errorf("invalid filter format"))
				return
			}
		}
	}
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	var totalBytes float64
	var totalPackets float64
	var bps float64
	var pps float64
	isNetFlow := pe.Type == "netflow"
	datastore.ForEachLog(st, et, pe.Type, func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		var ok bool
		var sa string
		var sp float64
		var da string
		var dp float64
		var bytes float64
		var packets float64
		var ft float64
		var lt float64
		var pi int
		if isNetFlow {
			if sa, ok = sl["srcAddr"].(string); !ok {
				return true
			}
			if sp, ok = sl["srcPort"].(float64); !ok {
				return true
			}
			if da, ok = sl["dstAddr"].(string); !ok {
				return true
			}
			if dp, ok = sl["dstPort"].(float64); !ok {
				return true
			}
			pi = 0
			if v, ok := sl["protocol"]; ok {
				pi = int(v.(float64))
			}
			if packets, ok = sl["packets"].(float64); !ok {
				return true
			}
			if bytes, ok = sl["bytes"].(float64); !ok {
				return true
			}
			if lt, ok = sl["last"].(float64); !ok {
				return true
			}
			if ft, ok = sl["first"].(float64); !ok {
				return true
			}
		} else {
			if ft, ok = sl["flowStartSysUpTime"].(float64); !ok {
				return true
			}
			if lt, ok = sl["flowEndSysUpTime"].(float64); !ok {
				return true
			}
			if sa, ok = sl["sourceIPv4Address"].(string); !ok {
				if sa, ok = sl["sourceIPv6Address"].(string); !ok {
					return true
				}
			}
			if da, ok = sl["destinationIPv4Address"].(string); !ok {
				if da, ok = sl["destinationIPv6Address"].(string); !ok {
					return true
				}
			}
			sp = 0
			dp = 0
			var icmpTypeCode float64
			if icmpTypeCode, ok = sl["icmpTypeCodeIPv6"].(float64); ok {
				sp = float64(int(icmpTypeCode) / 256)
				dp = float64(int(icmpTypeCode) % 256)
				pi = 1
			} else if icmpTypeCode, ok = sl["icmpTypeCodeIPv4"].(float64); ok {
				sp = float64(int(icmpTypeCode) / 256)
				dp = float64(int(icmpTypeCode) % 256)
				pi = 1
			} else if pif, ok := sl["protocolIdentifier"].(float64); ok {
				if sp, ok = sl["sourceTransportPort"].(float64); !ok {
					return true
				}
				if dp, ok = sl["destinationTransportPort"].(float64); !ok {
					return true
				}
				pi = int(pif)
			}
			if packets, ok = sl["packetDeltaCount"].(float64); !ok {
				return true
			}
			if bytes, ok = sl["octetDeltaCount"].(float64); !ok {
				return true
			}
		}
		// Filter
		if filterProt > 0 {
			if filterProt != pi {
				return true
			}
		}
		if filterPort > 0 {
			if filterPort != int(sp) && filterPort != int(dp) {
				return true
			}
		}
		if filterIP != nil {
			if !filterIP.Match([]byte(sa)) && !filterIP.Match([]byte(da)) {
				return true
			}
		}
		if filterSrc != nil {
			if !filterIP.Match([]byte(sa)) {
				return true
			}
		}
		if filterDst != nil {
			if !filterIP.Match([]byte(da)) {
				return true
			}
		}
		dur := (lt - ft) / 100.0
		totalBytes += bytes
		totalPackets += packets
		if dur > 0.0 {
			bps += (bytes / dur)
			pps += (packets / dur)
		}
		return true
	})
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	pe.Result["lastTime"] = et
	pe.Result["bytes"] = totalBytes
	pe.Result["packets"] = totalPackets
	pe.Result["bps"] = bps
	pe.Result["pps"] = pps
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm.Set("bps", bps)
	vm.Set("pps", pps)
	vm.Set("bytes", totalBytes)
	vm.Set("packets", totalPackets)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("netflow", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func makeRegexpFilter(f string) *regexp.Regexp {
	reg, err := regexp.Compile(f)
	if err != nil {
		return nil
	}
	return reg
}
