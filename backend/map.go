package backend

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func mapBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start map")
	clearPollingState()
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		updateNodeState(n)
		return true
	})
	hasLineInfo := updateLineState()
	go checkNewVersion()
	timer := time.NewTicker(time.Second * 10)
	newVersionTimer := time.NewTicker(time.Hour * 24)
	i := 6
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			newVersionTimer.Stop()
			log.Println("stop map")
			return
		case <-newVersionTimer.C:
			datastore.SaveMapData()
			go checkNewVersion()
		case <-timer.C:
			change := 0
			datastore.ForEachStateChangedNodes(func(id string) bool {
				if n := datastore.GetNode(id); n != nil {
					updateNodeState(n)
					change++
				}
				datastore.DeleteNodeStateChanged(id)
				return true
			})
			if change > 0 {
				hasLineInfo = updateLineState()
			}
			i++
			if i > 5 {
				if hasLineInfo {
					hasLineInfo = updateLineState()
				}
				datastore.UpdateDBStats()
				datastore.CheckDBBackup()
				i = 0
			}
		}
	}
}

func clearPollingState() {
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.State == "repair" {
			p.State = "unknown"
			p.NextTime = 0
		}
		return true
	})
}

func updateNodeState(n *datastore.NodeEnt) {
	n.State = "unknown"
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID != n.ID {
			return true
		}
		s := p.State
		if s == "high" {
			n.State = "high"
			return false
		}
		if s == "low" {
			n.State = "low"
			return true
		}
		if n.State == "low" {
			return true
		}
		if s == "warn" {
			n.State = "warn"
			return true
		}
		if n.State == "warn" {
			return true
		}
		if s == "repair" {
			// 復帰時に自動確認
			if !n.AutoAck {
				n.State = "repair"
			} else {
				p.State = "normal"
				s = "normal"
			}
		}
		if n.State == "repair" || n.State != "unknown" {
			return true
		}
		n.State = s
		return true
	})
}

func updateLineState() bool {
	hasLineInfo := false
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if p := datastore.GetPolling(l.PollingID1); p != nil {
			l.State1 = p.State
		} else {
			l.State1 = "unknown"
		}
		if p := datastore.GetPolling(l.PollingID2); p != nil {
			l.State2 = p.State
		} else {
			l.State2 = "unknown"
		}
		if l.PollingID != "" {
			if p := datastore.GetPolling(l.PollingID); p != nil {
				l.State = p.State
				if v, ok := p.Result["bps"]; ok {
					if vf, ok := v.(float64); ok {
						l.Width = int(vf / (1024 * 1024 * 10))
						if l.Width > 5 {
							l.Width = 5
						}
						hasLineInfo = true
						l.Info = humanize.Bytes(uint64(vf)) + "PS"
					}
				} else {
					if v, ok := p.Result["pps"]; ok {
						if vf, ok := v.(float64); ok {
							hasLineInfo = true
							l.Info = humanize.Commaf(vf) + "PPS"
						}
					}
				}
				if v, ok := p.Result["obps"]; ok {
					if vf, ok := v.(float64); ok {
						hasLineInfo = true
						l.Info += "/" + humanize.Bytes(uint64(vf)) + "PS"
					}
				}
			}
		}
		return true
	})
	return hasLineInfo
}

func checkNewVersion() {
	if !datastore.NotifyConf.CheckUpdate || versionCheckState > 1 {
		return
	}
	url := "https://lhx98.linkclub.jp/twise.co.jp/cgi-bin/twsnmpfc.cgi?ver=" + versionNum
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("check new version err=%v", err)
		return
	}
	defer resp.Body.Close()
	ba, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("check new version err=%v", err)
		return
	}
	if CmpVersion(versionNum, strings.TrimSpace(string(ba))) >= 0 {
		if versionCheckState == 0 {
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "TWSNMPのバージョンは最新です。",
			})
			versionCheckState = 1
		}
		return
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "warn",
		Event: "TWSNMPの新しいバージョンがあります。",
	})
	versionCheckState = 2
}

func CmpVersion(mv, sv string) int {
	mv = strings.ReplaceAll(mv, "(", " ")
	mv = strings.ReplaceAll(mv, "v", "")
	mv = strings.ReplaceAll(mv, "x", "0")
	sv = strings.ReplaceAll(sv, "v", "")
	mva := strings.Split(mv, ".")
	sva := strings.Split(sv, ".")
	for i := 0; i < len(mva) && i < len(sva) && i < 3; i++ {
		sn, err := strconv.ParseInt(sva[i], 10, 64)
		if err != nil {
			log.Println(err)
			return 0
		}
		mn, err := strconv.ParseInt(mva[i], 10, 64)
		if err != nil {
			log.Println(err)
			return 0
		}
		if sn > mn {
			return -1
		}
		if sn < mn {
			return 1
		}
	}
	return 0
}
