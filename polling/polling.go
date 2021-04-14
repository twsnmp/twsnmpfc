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
	"fmt"
	"log"
	"runtime"
	"sort"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	doPollingCh chan bool
)

func Start(ctx context.Context) error {
	doPollingCh = make(chan bool)
	go pollingBackend(ctx)
	return nil
}

func AutoAddPolling(n *datastore.NodeEnt, pt *datastore.PollingTemplateEnt) {
	switch pt.Type {
	case "snmp":
		autoAddSnmpPolling(n, pt)
	case "vmware":
		autoAddVMwarePolling(n, pt)
	case "tcp", "http", "tls":
		autoAddTCPPolling(n, pt)
	default:
		log.Printf("AutoAddPolling not supported type=%s", pt.Type)
	}
}

func PollNowNode(nodeID string) {
	n := datastore.GetNode(nodeID)
	if n == nil {
		return
	}
	n.State = "unknown"
	datastore.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.NodeID == nodeID && pe.State != "normal" {
			pe.State = "unknown"
			pe.NextTime = 0
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   pe.NodeID,
				NodeName: n.Name,
				Event:    "ポーリング再確認:" + pe.Name,
			})
			datastore.UpdatePolling(pe)
		}
		return true
	})
	datastore.SetNodeStateChanged(n.ID)
	doPollingCh <- true
}

func CheckAllPoll() {
	datastore.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.State != "normal" {
			pe.State = "unknown"
			pe.NextTime = 0
			n := datastore.GetNode(pe.NodeID)
			if n == nil {
				return true
			}
			n.State = "unknown"
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   pe.NodeID,
				NodeName: n.Name,
				Event:    "ポーリング再確認:" + pe.Name,
			})
			datastore.SetNodeStateChanged(n.ID)
			datastore.UpdatePolling(pe)
		}
		return true
	})
	doPollingCh <- true
}

// pollingBackend :  ポーリングのバックグランド処理
func pollingBackend(ctx context.Context) {
	time.Sleep(time.Millisecond * 100)
	timer := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			checkPolling()
		case <-doPollingCh:
			checkPolling()
		}
	}
}

func checkPolling() {
	now := time.Now().UnixNano()
	list := []*datastore.PollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.Level != "off" && p.NextTime < (now+(10*1000*1000*1000)) {
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
		go doPolling(list[i], startTime)
		time.Sleep(time.Millisecond * 2)
	}
}

func doPolling(pe *datastore.PollingEnt, startTime int64) {
	for startTime > time.Now().UnixNano() {
		time.Sleep(time.Millisecond * 100)
	}
	oldState := pe.State
	switch pe.Type {
	case "ping":
		doPollingPing(pe)
	case "snmp":
		doPollingSnmp(pe)
	case "tcp":
		doPollingTCP(pe)
	case "http":
		doPollingHTTP(pe)
	case "tls":
		doPollingTLS(pe)
	case "dns":
		doPollingDNS(pe)
	case "ntp":
		doPollingNTP(pe)
	case "syslog", "trap", "netflow", "ipfix":
		doPollingLog(pe)
	case "cmd":
		doPollingCmd(pe)
	case "ssh":
		doPollingSSH(pe)
	case "vmware":
		doPollingVMWare(pe)
	case "twsnmp":
		doPollingTWSNMP(pe)
	case "report":
		doPollingReport(pe)
	}
	datastore.UpdatePolling(pe)
	if pe.LogMode == datastore.LogModeAlways || pe.LogMode == datastore.LogModeAI || (pe.LogMode == datastore.LogModeOnChange && oldState != pe.State) {
		if err := datastore.AddPollingLog(pe); err != nil {
			log.Printf("addPollingLog err=%v %#v", err, pe)
		}
	}
	if datastore.InfluxdbConf.PollingLog != "" {
		if datastore.InfluxdbConf.PollingLog == "all" || pe.LogMode != datastore.LogModeNone {
			_ = datastore.SendPollingLogToInfluxdb(pe)
		}
	}
}

func setPollingState(pe *datastore.PollingEnt, newState string) {
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
		if n := datastore.GetNode(pe.NodeID); n != nil {
			nodeName = n.Name
		}
		datastore.SetNodeStateChanged(pe.NodeID)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "polling",
			Level:    pe.State,
			NodeID:   pe.NodeID,
			NodeName: nodeName,
			Event:    fmt.Sprintf("ポーリング状態変化:%s(%s):%s", pe.Name, pe.Type, oldState),
		})
	}
}

func setPollingError(s string, pe *datastore.PollingEnt, err error) {
	log.Printf("%s error Polling=%s err=%v", s, pe.Name, err)
	pe.Result["error"] = fmt.Sprintf("%v", err)
	setPollingState(pe, "unknown")
}

func hasSameNamePolling(nodeID, name string) bool {
	r := false
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == nodeID && p.Name == name {
			r = true
			return false
		}
		return true
	})
	return r
}
