package backend

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func (b *Backend) mapBackend(ctx context.Context) {
	b.clearPollingState()
	b.ds.ForEachNodes(func(n *datastore.NodeEnt) bool {
		b.updateNodeState(n)
		return true
	})
	b.updateLineState()
	go b.checkNewVersion()
	timer := time.NewTicker(time.Second * 10)
	newVersionTimer := time.NewTicker(time.Hour * 24)
	i := 6
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-newVersionTimer.C:
			go b.checkNewVersion()
		case <-timer.C:
			bHit := false
			b.ds.ForEachStateChangedNodes(func(id string) bool {
				if n := b.ds.GetNode(id); n != nil {
					b.updateNodeState(n)
					bHit = true
				}
				b.ds.DeleteNodeStateChanged(id)
				return true
			})
			if bHit {
				b.updateLineState()
			}
			i++
			if i > 5 {
				b.ds.UpdateDBStats()
				b.ds.CheckDBBackup()
				i = 0
			}
		}
	}
}

func (b *Backend) clearPollingState() {
	b.ds.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.State == "repair" {
			p.State = "unknown"
			p.NextTime = 0
		}
		return true
	})
}

func (b *Backend) updateNodeState(n *datastore.NodeEnt) {
	n.State = "unknown"
	b.ds.ForEachPollings(func(p *datastore.PollingEnt) bool {
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
		if s == "repair" {
			n.State = "repair"
		}
		if n.State == "repair" || n.State != "unknown" {
			return true
		}
		n.State = s
		return true
	})
}

func (b *Backend) updateLineState() {
	b.ds.ForEachLines(func(l *datastore.LineEnt) bool {
		if p := b.ds.GetPolling(l.PollingID1); p != nil {
			l.State1 = p.State
		} else {
			l.State1 = "unknown"
		}
		if p := b.ds.GetPolling(l.PollingID2); p != nil {
			l.State2 = p.State
		} else {
			l.State2 = "unknown"
		}
		return true
	})
}

func (b *Backend) checkNewVersion() {
	if !b.ds.NotifyConf.CheckUpdate || b.versionCheckState > 1 {
		return
	}
	url := "https://lhx98.linkclub.jp/twise.co.jp/cgi-bin/twsnmp/twsnmp.cgi?twsver=" + b.versionNum
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("checkNewVersion err=%v", err)
		return
	}
	defer resp.Body.Close()
	ba, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("checkNewVersion err=%v", err)
		return
	}
	if strings.Contains(string(ba), "#TWSNMPVEROK#") {
		if b.versionCheckState == 0 {
			b.ds.AddEventLog(&datastore.EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "TWSNMPのバージョンは最新です。",
			})
			b.versionCheckState = 1
		}
		return
	}
	b.ds.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "warn",
		Event: "TWSNMPの新しいバージョンがあります。",
	})
	b.versionCheckState = 2
}
