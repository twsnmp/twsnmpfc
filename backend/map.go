package backend

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/twsnmp/twsnmpfc/datastore"
)

var SaveMapInterval = 60

func mapBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start map backend")
	clearPollingState()
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		updateNodeState(n)
		return true
	})
	updateLineState()
	go checkNewVersion()
	timer := time.NewTicker(time.Second * 10)
	newVersionTimer := time.NewTicker(time.Hour * 24)
	i := 6
	save := 0
	checkOR := false
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			newVersionTimer.Stop()
			log.Println("stop map backend")
			return
		case <-newVersionTimer.C:
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
				updateLineState()
				checkOR = true
			}
			i++
			if i > 5 {
				if !checkOR {
					updateLineState()
				}
				datastore.UpdateDBStats()
				datastore.CheckDBBackup()
				if checkOR {
					checkOperationRate()
					checkOR = false
				}
				i = 0
				save++
				if save > SaveMapInterval {
					datastore.SaveMapData()
					save = 0
				}
			}
		}
	}
}

func checkOperationRate() {
	if datastore.MapConf.DisableOperLog {
		log.Println("disable oprate log")
		return
	}
	total := 0
	down := 0
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		total++
		switch n.State {
		case "normal":
		case "repair":
		case "unknown":
		default:
			down++
		}
		return true
	})

	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "oprate",
		Level: "info",
		Event: fmt.Sprintf("ノード数=%d,障害ノード=%d,稼働率=%.2f%%", total, down, 100.0*float64(total-down)/float64(total)),
	})
}

// clearPollingState : 復帰状態のポーリング状態を不明にして、再ポーリングする
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
		if p.NodeID != n.ID || p.Level == "off" {
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
				return true
			} else {
				p.State = "normal"
				s = "normal"
			}
		}
		if n.State != "unknown" {
			return true
		}
		if s == "info" {
			s = "normal"
		}
		n.State = s
		return true
	})
}

func updateLineState() {
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		l.State1 = "unknown"
		if strings.HasPrefix(l.NodeID1, "NET:") {
			n := datastore.GetNetwork(l.NodeID1)
			hit := false
			if n != nil {
				for _, p := range n.Ports {
					if p.ID == l.PollingID1 {
						l.State1 = p.State
						hit = true
					}
				}
			}
			if !hit {
				log.Printf("delete invalid line node=%+v line=%+v", n, l)
				datastore.DeleteLine(l.ID)
				return true
			}
		} else {
			if p := datastore.GetPolling(l.PollingID1); p != nil {
				l.State1 = p.State
			}
		}
		l.State2 = l.State1
		if strings.HasPrefix(l.NodeID2, "NET:") {
			n := datastore.GetNetwork(l.NodeID2)
			hit := false
			if n != nil {
				for _, p := range n.Ports {
					if p.ID == l.PollingID2 {
						l.State2 = p.State
						hit = true
					}
				}
			}
			if !hit {
				log.Printf("delete invalid line node=%+v line=%+v", n, l)
				datastore.DeleteLine(l.ID)
				return true
			}
		} else {
			if p := datastore.GetPolling(l.PollingID2); p != nil {
				l.State2 = p.State
				if l.PollingID1 == "" {
					l.State1 = l.State2
				}
			}
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
						l.Info = humanize.Bytes(uint64(vf)) + "PS"
					}
				} else {
					if v, ok := p.Result["pps"]; ok {
						if vf, ok := v.(float64); ok {
							l.Info = humanize.Commaf(vf) + "PPS"
						}
					}
				}
				if v, ok := p.Result["obps"]; ok {
					if vf, ok := v.(float64); ok {
						l.Info += "/" + humanize.Bytes(uint64(vf)) + "PS"
					}
				}
			}
		}
		return true
	})
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
	ba, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("check new version err=%v", err)
		return
	}
	level := "info"
	v := strings.TrimSpace(string(ba))
	if CmpVersion(versionNum, v) >= 0 {
		if versionCheckState == 0 {
			versionCheckState = 1
		}
	} else {
		versionCheckState = 2
		level = "warn"
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: level,
		Event: fmt.Sprintf("利用中のTWSNMP FCのバージョンは%s、最新バージョンは%s", versionNum, v),
	})
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
