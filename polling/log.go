package polling

// LOG監視ポーリング処理

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func getLogContent(l string) string {
	sm := make(map[string]interface{})
	if err := json.Unmarshal([]byte(l), &sm); err != nil {
		log.Printf("getLogContent err=%v", err)
		return ""
	}
	if s, ok := sm["content"]; ok {
		return s.(string)
	}
	log.Println("getLogContent no content")
	return ""
}

func (p *Polling) doPollingLog(pe *datastore.PollingEnt) {
	cmds := splitCmd(pe.Polling)
	if len(cmds) != 3 {
		p.setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
		return
	}
	filter := "`" + cmds[0] + "`"
	extractor := cmds[1]
	script := cmds[2]
	if _, err := regexp.Compile(filter); err != nil {
		p.setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
		return
	}
	vm := otto.New()
	lr := make(map[string]string)
	st := ""
	if err := json.Unmarshal([]byte(pe.LastResult), &lr); err != nil {
		log.Printf("doPollingLog err=%v", err)
	} else {
		st = lr["lastTime"]
	}
	if _, err := time.Parse("2006-01-02T15:04", st); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(pe.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := p.ds.GetLogs(&datastore.LogFilterEnt{
		Filter:    filter,
		StartTime: st,
		EndTime:   et,
		LogType:   pe.Type,
	})
	lr["lastTime"] = et
	_ = vm.Set("count", len(logs))
	_ = vm.Set("interval", pe.PollInt)
	lr["count"] = fmt.Sprintf("%d", len(logs))
	pe.LastVal = float64(len(logs))
	if extractor == "" {
		value, err := vm.Run(script)
		if err == nil {
			pe.LastResult = makeLastResult(lr)
			if ok, _ := value.ToBoolean(); ok {
				p.setPollingState(pe, "normal")
			} else {
				p.setPollingState(pe, pe.Level)
			}
			return
		}
		p.setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
		return
	}
	grokEnt := p.ds.GetGrokEnt(extractor)
	if grokEnt == nil {
		p.setPollingError("log", pe, fmt.Errorf("no extractor pattern"))
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
		p.setPollingError("log", pe, fmt.Errorf("no extractor pattern"))
		return
	}
	cap := fmt.Sprintf("%%{%s}", extractor)
	for _, l := range logs {
		cont := getLogContent(l.Log)
		values, err := g.Parse(cap, cont)
		if err != nil {
			continue
		}
		for k, v := range values {
			_ = vm.Set(k, v)
			lr[k] = v
		}
		value, err := vm.Run(script)
		if err == nil {
			if ok, _ := value.ToBoolean(); !ok {
				pe.LastResult = makeLastResult(lr)
				p.setPollingState(pe, pe.Level)
				return
			}
		} else {
			p.setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
			return
		}
	}
	pe.LastResult = makeLastResult(lr)
	p.setPollingState(pe, "normal")
}

var syslogPriFilter = regexp.MustCompile(`"priority":(\d+),`)

func (p *Polling) doPollingSyslogPri(pe *datastore.PollingEnt) bool {
	cmds := splitCmd(pe.Polling)
	if len(cmds) < 1 {
		p.setPollingError("log", pe, fmt.Errorf("invalid syslog pri watch format"))
		return false
	}
	filter := cmds[0]
	_, err := regexp.Compile(filter)
	if err != nil {
		p.setPollingError("log", pe, fmt.Errorf("invalid syslogpri watch format"))
		return false
	}
	endTime := time.Unix((time.Now().Unix()/3600)*3600, 0)
	startTime := endTime.Add(-time.Hour * 1)
	if int64(pe.LastVal) >= startTime.UnixNano() && pe.LastResult != "" {
		// Skip
		return false
	}
	pe.LastVal = float64(startTime.UnixNano())
	st := startTime.Format("2006-01-02T15:04")
	et := endTime.Format("2006-01-02T15:04")
	logs := p.ds.GetLogs(&datastore.LogFilterEnt{
		Filter:    "`" + filter + "`",
		StartTime: st,
		EndTime:   et,
		LogType:   "syslog",
	})
	priMap := make(map[int]int)
	for _, l := range logs {
		pa := syslogPriFilter.FindAllStringSubmatch(string(l.Log), -1)
		if pa == nil || len(pa) < 1 || len(pa[0]) < 2 {
			continue
		}
		pri, err := strconv.ParseInt(pa[0][1], 10, 64)
		if err != nil || pri < 0 || pri > 256 {
			continue
		}
		priMap[int(pri)]++
	}
	lr := make(map[string]string)
	for pri, c := range priMap {
		lr[fmt.Sprintf("pri_%d", pri)] = fmt.Sprintf("%d", c)
	}
	pe.LastResult = makeLastResult(lr)
	p.setPollingState(pe, "normal")
	return true
}

func (p *Polling) doPollingSyslogDevice(pe *datastore.PollingEnt) {
	cmds := splitCmd(pe.Polling)
	if len(cmds) != 2 {
		p.setPollingError("log", pe, fmt.Errorf("invalid syslog device watch format"))
		return
	}
	filter := cmds[0]
	extractor := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		p.setPollingError("log", pe, fmt.Errorf("invalid syslog device watch format"))
		return
	}
	grokEnt := p.ds.GetGrokEnt(extractor)
	if grokEnt == nil {
		p.setPollingError("log", pe, fmt.Errorf("invalid syslog device watch format"))
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
		p.setPollingError("log", pe, fmt.Errorf("invalid syslog device watch format err=%v", err))
		return
	}
	lr := make(map[string]string)
	st := ""
	if err := json.Unmarshal([]byte(pe.LastResult), &lr); err != nil {
		log.Printf("doPollingSyslogDevice err=%v", err)
	} else {
		st = lr["lastTime"]
	}
	if _, err := time.Parse("2006-01-02T15:04", st); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(pe.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := p.ds.GetLogs(&datastore.LogFilterEnt{
		Filter:    "`" + filter + "`",
		StartTime: st,
		EndTime:   et,
		LogType:   "syslog",
	})
	lr["lastTime"] = et
	lr["count"] = fmt.Sprintf("%d", len(logs))
	count := 0
	cap := fmt.Sprintf("%%{%s}", extractor)
	for _, l := range logs {
		cont := getLogContent(l.Log)
		values, err := g.Parse(cap, cont)
		if err != nil {
			log.Printf("err=%v", err)
			continue
		}
		mac, ok := values["mac"]
		if !ok {
			continue
		}
		ip, ok := values["ip"]
		if !ok {
			continue
		}
		count++
		p.report.ReportDevice(mac, ip, l.Time)
	}
	lr["hit"] = fmt.Sprintf("%d", count)
	pe.LastVal = float64(count)
	pe.LastResult = makeLastResult(lr)
	p.setPollingState(pe, "normal")
}

func (p *Polling) doPollingSyslogUser(pe *datastore.PollingEnt) {
	n := p.ds.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found nodeID=%s", pe.NodeID)
		return
	}
	cmds := splitCmd(pe.Polling)
	if len(cmds) != 2 {
		p.setPollingError("log", pe, fmt.Errorf("invalid syslog user watch format"))
		return
	}
	filter := cmds[0]
	extractor := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		p.setPollingError("log", pe, fmt.Errorf("invalid filter for syslog user"))
		return
	}
	grokEnt := p.ds.GetGrokEnt(extractor)
	if grokEnt == nil {
		p.setPollingError("log", pe, fmt.Errorf("invalid extractor for syslog user"))
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
		log.Printf("doPollingSyslogUser err=%v", err)
	}
	lr := make(map[string]string)
	st := ""
	if err := json.Unmarshal([]byte(pe.LastResult), &lr); err != nil {
		log.Printf("doPollingSyslogUser err=%v", err)
	} else {
		st = lr["lastTime"]
	}
	if _, err := time.Parse("2006-01-02T15:04", st); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(pe.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := p.ds.GetLogs(&datastore.LogFilterEnt{
		Filter:    "`" + filter + "`",
		StartTime: st,
		EndTime:   et,
		LogType:   "syslog",
	})
	lr["lastTime"] = et
	lr["count"] = fmt.Sprintf("%d", len(logs))
	okCount := 0
	totalCount := 0
	cap := fmt.Sprintf("%%{%s}", extractor)
	for _, l := range logs {
		cont := getLogContent(l.Log)
		values, err := g.Parse(cap, cont)
		if err != nil {
			log.Printf("err=%v", err)
			continue
		}
		stat, ok := values["stat"]
		if !ok {
			continue
		}
		user, ok := values["user"]
		if !ok {
			continue
		}
		client := values["client"]
		ok = grokEnt.Ok == stat
		totalCount++
		if ok {
			okCount++
		}
		p.report.ReportUser(user, n.IP, client, ok, l.Time)
	}
	if totalCount > 0 {
		pe.LastVal = float64(okCount) / float64(totalCount)
	} else {
		pe.LastVal = 1.0
	}
	lr["total"] = fmt.Sprintf("%d", totalCount)
	lr["ok"] = fmt.Sprintf("%d", okCount)
	lr["rate"] = fmt.Sprintf("%f", pe.LastVal*100.0)
	pe.LastResult = makeLastResult(lr)
	p.setPollingState(pe, "normal")
}

func (p *Polling) doPollingSyslogFlow(pe *datastore.PollingEnt) {
	cmds := splitCmd(pe.Polling)
	if len(cmds) != 2 {
		p.setPollingError("syslogFlow", pe, fmt.Errorf("invalid watch format"))
		return
	}
	filter := cmds[0]
	extractor := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		p.setPollingError("syslogFlow", pe, fmt.Errorf("invalid filter"))
		return
	}
	grokEnt := p.ds.GetGrokEnt(extractor)
	if grokEnt == nil {
		p.setPollingError("syslogFlow", pe, fmt.Errorf("invalid extractor"))
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
		p.setPollingError("syslogFlow", pe, fmt.Errorf("invalid extractor"))
		return
	}
	lr := make(map[string]string)
	st := ""
	if err := json.Unmarshal([]byte(pe.LastResult), &lr); err != nil {
		log.Printf("doPollingSyslogFlow err=%v", err)
	} else {
		st = lr["lastTime"]
	}
	if _, err := time.Parse("2006-01-02T15:04", st); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(pe.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := p.ds.GetLogs(&datastore.LogFilterEnt{
		Filter:    "`" + filter + "`",
		StartTime: st,
		EndTime:   et,
		LogType:   "syslog",
	})
	lr["lastTime"] = et
	lr["count"] = fmt.Sprintf("%d", len(logs))
	count := 0
	cap := fmt.Sprintf("%%{%s}", extractor)
	for _, l := range logs {
		cont := getLogContent(l.Log)
		values, err := g.Parse(cap, cont)
		if err != nil {
			continue
		}
		src, ok := values["src"]
		if !ok {
			continue
		}
		dst, ok := values["dst"]
		if !ok {
			continue
		}
		sport, ok := values["sport"]
		if !ok {
			continue
		}
		dport, ok := values["dport"]
		if !ok {
			continue
		}
		prot, ok := values["prot"]
		if !ok {
			continue
		}
		nBytes := 0
		for _, b := range []string{"bytes", "sent", "rcvd"} {
			bytes, ok := values[b]
			if ok {
				nB, _ := strconv.Atoi(bytes)
				nBytes += nB
			}
		}
		nProt := getProt(prot)
		nSPort, _ := strconv.Atoi(sport)
		nDPort, _ := strconv.Atoi(dport)
		p.report.ReportFlow(src, nSPort, dst, nDPort, nProt, int64(nBytes), l.Time)
		count++
	}
	lr["hit"] = fmt.Sprintf("%d", count)
	pe.LastVal = float64(count)
	pe.LastResult = makeLastResult(lr)
	p.setPollingState(pe, "normal")
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
