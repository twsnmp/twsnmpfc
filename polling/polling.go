// Package polling : ポーリング処理
package polling

/*
polling.go :ポーリング処理を行う
ポーリングの種類は
(1)能動的なポーリング
 ping
 snmp - sysUptime,ifOperStatus,
 http
 https
 tls
 dns
（２）受動的なポーリング
 syslog
 snmp trap
 netflow
 ipfix

*/

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
	"github.com/twsnmp/twsnmpfc/report"
)

type Polling struct {
	ds          *datastore.DataStore
	ping        *ping.Ping
	report      *report.Report
	doPollingCh chan bool
}

func NewPolling(ctx context.Context, ds *datastore.DataStore, report *report.Report, ping *ping.Ping) *Polling {
	p := &Polling{
		ds:     ds,
		ping:   ping,
		report: report,
	}
	go p.pollingBackend(ctx)
	return p
}

func (p *Polling) pollNowNode(nodeID string) {
	n := p.ds.GetNode(nodeID)
	if n == nil {
		return
	}
	p.ds.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.NodeID == nodeID && pe.State != "normal" {
			pe.State = "unknown"
			pe.NextTime = 0
			p.ds.AddEventLog(&datastore.EventLogEnt{
				Type:     "user",
				Level:    pe.State,
				NodeID:   pe.NodeID,
				NodeName: n.Name,
				Event:    "ポーリング再確認:" + pe.Name,
			})
			p.ds.UpdatePolling(pe)
		}
		return true
	})
	p.ds.SetNodeStateChanged(n.ID)
	p.doPollingCh <- true
}

func (p *Polling) CheckAllPoll() {
	p.ds.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.State != "normal" {
			pe.State = "unknown"
			pe.NextTime = 0
			n := p.ds.GetNode(pe.NodeID)
			if n == nil {
				return true
			}
			p.ds.AddEventLog(&datastore.EventLogEnt{
				Type:     "user",
				Level:    pe.State,
				NodeID:   pe.NodeID,
				NodeName: n.Name,
				Event:    "ポーリング再確認:" + pe.Name,
			})
			p.ds.SetNodeStateChanged(n.ID)
			p.ds.UpdatePolling(pe)
		}
		return true
	})
	p.doPollingCh <- true
}

// pollingBackend :  ポーリングのバックグランド処理
func (p *Polling) pollingBackend(ctx context.Context) {
	time.Sleep(time.Millisecond * 100)
	timer := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			p.checkPolling()
		case <-p.doPollingCh:
			p.checkPolling()
		}
	}
}

func (p *Polling) checkPolling() {
	now := time.Now().UnixNano()
	list := []*datastore.PollingEnt{}
	p.ds.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NextTime < (now + (10 * 1000 * 1000 * 1000)) {
			list = append(list, p)
		}
		return true
	})
	if len(list) < 1 {
		return
	}
	log.Printf("doPolling=%d NumGoroutine=%d", len(list), runtime.NumGoroutine())
	sort.Slice(list, func(i, j int) bool {
		return list[i].NextTime < list[j].NextTime
	})
	for i := 0; i < len(list); i++ {
		startTime := list[i].NextTime
		if startTime < now {
			startTime = now
		}
		list[i].NextTime = startTime + (int64(list[i].PollInt) * 1000 * 1000 * 1000)
		go p.doPolling(list[i], startTime)
		time.Sleep(time.Millisecond * 2)
	}
}

func (p *Polling) doPolling(pe *datastore.PollingEnt, startTime int64) {
	for startTime > time.Now().UnixNano() {
		time.Sleep(time.Millisecond * 100)
	}
	oldState := pe.State
	switch pe.Type {
	case "ping":
		p.doPollingPing(pe)
	case "snmp":
		p.doPollingSnmp(pe)
	case "tcp":
		p.doPollingTCP(pe)
	case "http", "https":
		p.doPollingHTTP(pe)
	case "tls":
		p.doPollingTLS(pe)
	case "dns":
		p.doPollingDNS(pe)
	case "ntp":
		p.doPollingNTP(pe)
	case "syslog", "trap", "netflow", "ipfix":
		p.doPollingLog(pe)
	case "syslogpri":
		if !p.doPollingSyslogPri(pe) {
			return
		}
	case "syslogdevice":
		p.doPollingSyslogDevice(pe)
	case "sysloguser":
		p.doPollingSyslogUser(pe)
	case "syslogflow":
		p.doPollingSyslogFlow(pe)
	case "cmd":
		p.doPollingCmd(pe)
	case "ssh":
		p.doPollingSSH(pe)
	case "vmware":
		p.doPollingVMWare(pe)
	case "twsnmp":
		p.doPollingTWSNMP(pe)
	}
	p.ds.UpdatePolling(pe)
	if pe.LogMode == datastore.LogModeAlways || pe.LogMode == datastore.LogModeAI || (pe.LogMode == datastore.LogModeOnChange && oldState != pe.State) {
		if err := p.ds.AddPollingLog(pe); err != nil {
			log.Printf("addPollingLog err=%v %#v", err, pe)
		}
	}
	if p.ds.InfluxdbConf.PollingLog != "" {
		if p.ds.InfluxdbConf.PollingLog == "all" || pe.LogMode != datastore.LogModeNone {
			_ = p.ds.SendPollingLogToInfluxdb(pe)
		}
	}
}

func (p *Polling) setPollingState(pe *datastore.PollingEnt, newState string) {
	sendEvent := false
	oldState := pe.State
	if newState == "normal" {
		if pe.State != "normal" && pe.State != "repair" {
			if pe.State == "unknown" {
				pe.State = "normal"
			} else {
				pe.State = "repair"
			}
			sendEvent = true
		}
	} else if newState == "unknown" {
		if pe.State != "unknown" {
			pe.State = "unknown"
			sendEvent = true
		}
	} else {
		if pe.State != pe.Level {
			pe.State = pe.Level
			sendEvent = true
		}
	}
	if sendEvent {
		nodeName := "unknown"
		if n := p.ds.GetNode(pe.NodeID); n != nil {
			nodeName = n.Name
		}
		p.ds.SetNodeStateChanged(pe.NodeID)
		p.ds.AddEventLog(&datastore.EventLogEnt{
			Type:     "polling",
			Level:    pe.State,
			NodeID:   pe.NodeID,
			NodeName: nodeName,
			Event:    fmt.Sprintf("ポーリング状態変化:%s(%s):%s:%f:%s", pe.Name, pe.Type, oldState, pe.LastVal, pe.LastResult),
		})
	}
}

func (p *Polling) setPollingError(s string, pe *datastore.PollingEnt, err error) {
	log.Printf("%s error Polling=%s err=%v", s, pe.Polling, err)
	lr := make(map[string]string)
	lr["error"] = fmt.Sprintf("%v", err)
	pe.LastResult = makeLastResult(lr)
	p.setPollingState(pe, "unknown")
}

// Util Functions

func makeLastResult(lr map[string]string) string {
	if js, err := json.Marshal(lr); err == nil {
		return string(js)
	}
	return ""
}

func splitCmd(p string) []string {
	ret := []string{}
	bInQ := false
	tmp := ""
	for _, c := range p {
		if c == '|' {
			if !bInQ {
				ret = append(ret, strings.TrimSpace(tmp))
				tmp = ""
			}
			continue
		}
		if c == '`' {
			bInQ = !bInQ
		} else {
			tmp += string(c)
		}
	}
	ret = append(ret, strings.TrimSpace(tmp))
	return ret
}
