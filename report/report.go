// Package report : ポーリング処理
package report

import (
	"context"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/mrichman/godnsbl"
	"github.com/openrdap/rdap"
	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	deviceReportCh chan *deviceReportEnt
	userReportCh   chan *userReportEnt
	flowReportCh   chan *flowReportEnt
	badIPs         map[string]int64
)

func StartReport(ctx context.Context) error {
	deviceReportCh = make(chan *deviceReportEnt, 100)
	userReportCh = make(chan *userReportEnt, 100)
	flowReportCh = make(chan *flowReportEnt, 500)
	badIPs = make(map[string]int64)
	go reportBackend(ctx)
	return nil
}

func reportBackend(ctx context.Context) {
	timer := time.NewTicker(time.Minute * 5)
	checkOldReport()
	calcScore()
	last := int64(0)
	for {
		select {
		case <-ctx.Done():
			{
				timer.Stop()
				datastore.SaveReport(0)
				log.Printf("Stop reportBackend")
				return
			}
		case <-timer.C:
			{
				checkOldReport()
				calcScore()
				datastore.SaveReport(last)
				last = time.Now().UnixNano()
			}
		case dr := <-deviceReportCh:
			checkDeviceReport(dr)
		case ur := <-userReportCh:
			checkUserReport(ur)
		case fr := <-flowReportCh:
			checkFlowReport(fr)
		}
	}
}

func checkOldReport() {
	oh := -24
	svs := datastore.LenServers()
	if svs > 10000 {
		oh = -12 / (svs / 10000)
		if oh > -3 {
			oh = -3
		}
	}
	old := time.Now().Add(time.Hour * time.Duration(oh)).UnixNano()
	tooOld := time.Now().AddDate(0, 0, -datastore.MapConf.LogDays).UnixNano()
	checkOldServers(old, tooOld)
	checkOldFlows(old, tooOld)
	checkOldDevices(old)
	checkOldUsers(old)
}

func checkOldServers(old, tooOld int64) {
	count := 0
	ids := []string{}
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		if s.LastTime < old {
			if s.LastTime < tooOld || s.LastTime-s.FirstTime < 3600*1000*1000*1000 {
				ids = append(ids, s.ID)
			} else {
				for k, n := range s.Services {
					if n < 10 {
						delete(s.Services, k)
					}
				}
				if len(s.Services) < 1 {
					ids = append(ids, s.ID)
				}
			}
		}
		return true
	})
	for _, id := range ids {
		datastore.DeleteReport("servers", id)
		count++
	}
	if count > 0 {
		log.Printf("DeleteSevers=%d", count)
	}
}

func checkOldFlows(old, tooOld int64) {
	count := 0
	ids := []string{}
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		if f.LastTime < old {
			if f.LastTime < tooOld || f.LastTime-f.FirstTime < 3600*1000*1000*1000 {
				ids = append(ids, f.ID)
			} else {
				for k, n := range f.Services {
					if n < 10 {
						delete(f.Services, k)
					}
				}
				if len(f.Services) < 1 {
					ids = append(ids, f.ID)
				}
			}
		}
		return true
	})
	for _, id := range ids {
		datastore.DeleteReport("flows", id)
		count++
	}
	if count > 0 {
		log.Printf("DeleteFlows=%d", count)
	}
}

func checkOldDevices(tooOld int64) {
	count := 0
	ids := []string{}
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if d.LastTime < tooOld {
			ids = append(ids, d.ID)
		}
		return true
	})
	for _, id := range ids {
		datastore.DeleteReport("devices", id)
		count++
	}
	if count > 0 {
		log.Printf("DeleteDevices=%d", count)
	}
}

func checkOldUsers(tooOld int64) {
	count := 0
	ids := []string{}
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		if u.LastTime < tooOld {
			ids = append(ids, u.ID)
		}
		return true
	})
	for _, id := range ids {
		datastore.DeleteReport("users", id)
		count++
	}
	if count > 0 {
		log.Printf("DeleteUsers=%d", count)
	}
}

func calcScore() {
	calcDeviceScore()
	calcServerScore()
	calcFlowScore()
	calcUserScore()
	badIPs = make(map[string]int64)
}

func calcDeviceScore() {
	var xs []float64
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if n, ok := badIPs[d.IP]; ok {
			d.Penalty += n
		}
		if d.Penalty > 100 {
			d.Penalty = 100
		}
		xs = append(xs, float64(100-d.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	if sd == 0 {
		return
	}
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		d.Score = ((10 * (float64(100-d.Penalty) - m) / sd) + 50)
		return true
	})
}

func calcFlowScore() {
	var xs []float64
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		if f.Penalty > 100 {
			f.Penalty = 100
		}
		xs = append(xs, float64(100-f.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	if sd == 0 {
		return
	}
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		f.Score = ((10 * (float64(100-f.Penalty) - m) / sd) + 50)
		return true
	})
}

func calcUserScore() {
	var xs []float64
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		if u.Penalty > 100 {
			u.Penalty = 100
		}
		xs = append(xs, float64(100-u.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	if sd == 0 {
		return
	}
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		u.Score = ((10 * (float64(100-u.Penalty) - m) / sd) + 50)
		return true
	})
}

func calcServerScore() {
	var xs []float64
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		if s.Penalty > 100 {
			s.Penalty = 100
		}
		xs = append(xs, float64(100-s.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	if sd == 0 {
		return
	}
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		s.Score = ((10 * (float64(100-s.Penalty) - m) / sd) + 50)
		return true
	})
}

func getMeanSD(xs *[]float64) (float64, float64) {
	m, err := stats.Mean(*xs)
	if err != nil {
		return 0, 0
	}
	sd, err := stats.StandardDeviation(*xs)
	if err != nil {
		return 0, 0
	}
	return m, sd
}

func resetPenalty(report string) {
	if report == "devices" {
		datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
			d.Penalty = 0
			setDevicePenalty(d)
			d.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcDeviceScore()
	} else if report == "users" {
		datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
			u.Penalty = 0
			u.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcUserScore()
	} else if report == "servers" {
		datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
			if s.Loc == "" {
				s.Loc = datastore.GetLoc(s.Server)
			}
			setServerPenalty(s)
			s.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcServerScore()
	} else if report == "flows" {
		datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
			if f.ServerLoc == "" {
				f.ServerLoc = datastore.GetLoc(f.Server)
			}
			if f.ClientLoc == "" {
				f.ClientLoc = datastore.GetLoc(f.Client)
			}
			setFlowPenalty(f)
			f.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcFlowScore()
	}
}

// utils
func normMACAddr(m string) string {
	m = strings.Replace(m, "-", ":", -1)
	a := strings.Split(m, ":")
	r := ""
	for _, e := range a {
		if r != "" {
			r += ":"
		}
		if len(e) == 1 {
			r += "0"
		}
		r += e
	}
	return strings.ToUpper(r)
}

func findNameFromIP(ip string) string {
	if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
		return names[0]
	}
	n := datastore.FindNodeFromIP(ip)
	if n != nil {
		return n.Name
	}
	return ip
}

type ipInfoCache struct {
	Time   int64
	IPInfo *[][]string
}

var ipInfoCacheMap = make(map[string]*ipInfoCache)

var blacklists = []string{
	"b.barracudacentral.org",
	"bl.spamcop.net",
	"blacklist.woody.ch",
	"bogons.cymru.com",
	"cbl.abuseat.org",
	"combined.abuse.ch",
	"db.wpbl.info",
	"dnsbl-1.uceprotect.net",
	"dnsbl-2.uceprotect.net",
	"dnsbl-3.uceprotect.net",
	"dnsbl.dronebl.org",
	"dnsbl.inps.de",
	"dnsbl.sorbs.net",
	"drone.abuse.ch",
	"duinv.aupads.org",
	"dul.dnsbl.sorbs.net",
	"dyna.spamrats.com",
	"dynip.rothen.com",
	"http.dnsbl.sorbs.net",
	"ips.backscatterer.org",
	"ix.dnsbl.manitu.net",
	"korea.services.net",
	"misc.dnsbl.sorbs.net",
	"noptr.spamrats.com",
	"orvedb.aupads.org",
	"pbl.spamhaus.org",
	"proxy.bl.gweep.ca",
	"psbl.surriel.com",
	"relays.bl.gweep.ca",
	"relays.nether.net",
	"sbl.spamhaus.org",
	"smtp.dnsbl.sorbs.net",
	"socks.dnsbl.sorbs.net",
	"spam.abuse.ch",
	"spam.dnsbl.sorbs.net",
	"spam.spamrats.com",
	"spamrbl.imp.ch",
	"ubl.unsubscore.com",
	"virus.rbl.jp",
	"web.dnsbl.sorbs.net",
	"wormrbl.imp.ch",
	"xbl.spamhaus.org",
	"zen.spamhaus.org",
	"zombie.dnsbl.sorbs.net",

	"z.mailspike.net",
	"spamsources.fabel.dk",
	"spambot.bls.digibase.ca",
	"spam.dnsbl.anonmails.de",
	"singular.ttk.pte.hu",
	"all.s5h.net",
	"ubl.lashback.com",
	"dnsbl.spfbl.net",
}

func getIPInfo(ip string) *[][]string {
	if c, ok := ipInfoCacheMap[ip]; ok {
		if c.Time > time.Now().Unix()-60*60*24*7 {
			return c.IPInfo
		}
	}
	ret := [][]string{}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		client := &rdap.Client{}
		ri, err := client.QueryIP(ip)
		if err != nil {
			log.Printf("RDAP QueryIP error=%v", err)
			return
		}
		ret = append(ret, []string{"RDAP:IP Version", ri.IPVersion}) //IPバージョン
		ret = append(ret, []string{"RDAP:Type", ri.Type})            // 種類
		ret = append(ret, []string{"RDAP:Handole", ri.Handle})       //範囲
		ret = append(ret, []string{"RDAP:Name", ri.Name})            // 所有者
		ret = append(ret, []string{"RDAP:Country", ri.Country})      // 国
		ret = append(ret, []string{"RDAP:Whois Server", ri.Port43})  // Whoisの情報源
	}()
	rblMap := &sync.Map{}
	for i, source := range blacklists {
		wg.Add(1)
		go func(i int, source string) {
			defer wg.Done()
			rbl := godnsbl.Lookup(source, ip)
			if len(rbl.Results) > 0 && rbl.Results[0].Listed {
				rblMap.Store(source, `<i class="fas fa-exclamation-circle state state_high"></i>Listed :`+rbl.Results[0].Text)
			} else {
				rblMap.Store(source, `<i class="fas fa-check-circle state state_repair"></i>Not Listed`)
			}
		}(i, source)
	}
	wg.Wait()
	rblMap.Range(func(key, value interface{}) bool {
		ret = append(ret, []string{"DNSBL:" + key.(string), value.(string)})
		return true
	})
	ipInfoCacheMap[ip] = &ipInfoCache{
		Time:   time.Now().Unix(),
		IPInfo: &ret,
	}
	return &ret
}
