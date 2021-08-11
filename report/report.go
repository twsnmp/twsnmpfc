// Package report : ポーリング処理
package report

import (
	"context"
	"fmt"
	"log"
	"net"
	"regexp"
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
	checkCertCh    chan bool
	twpcapReportCh chan map[string]interface{}
	twWinLogCh     chan map[string]interface{}
	allowDNS       map[string]bool
	allowDHCP      map[string]bool
	allowMail      map[string]bool
	allowLDAP      map[string]bool
	allowLocalIP   *regexp.Regexp
	dennyCountries map[string]int64
	dennyServices  map[string]int64
	japanOnly      bool
)

func Start(ctx context.Context) error {
	datastore.LaodReportConf()
	UpdateReportConf()
	datastore.LoadReport()
	deviceReportCh = make(chan *deviceReportEnt, 100)
	userReportCh = make(chan *userReportEnt, 100)
	flowReportCh = make(chan *flowReportEnt, 500)
	twpcapReportCh = make(chan map[string]interface{}, 100)
	twWinLogCh = make(chan map[string]interface{}, 100)
	checkCertCh = make(chan bool, 5)
	go reportBackend(ctx)
	return nil
}

func UpdateReportConf() {
	japanOnly = datastore.ReportConf.JapanOnly
	if japanOnly || len(datastore.ReportConf.DenyCountries) < 1 {
		dennyCountries = nil
	} else {
		dennyCountries = make(map[string]int64)
		for _, c := range datastore.ReportConf.DenyCountries {
			dennyCountries[c] = 0
		}
	}
	if len(datastore.ReportConf.DenyServices) < 1 {
		dennyServices = nil
	} else {
		dennyServices = make(map[string]int64)
		for _, s := range datastore.ReportConf.DenyServices {
			dennyServices[s] = 0
		}
	}
	ips := strings.Split(datastore.ReportConf.AllowDNS, ",")
	if len(ips) < 1 || ips[0] == "" {
		allowDNS = nil
	} else {
		allowDNS = make(map[string]bool)
		for _, ip := range ips {
			allowDNS[ip] = true
		}
	}
	ips = strings.Split(datastore.ReportConf.AllowDHCP, ",")
	if len(ips) < 1 || ips[0] == "" {
		allowDHCP = nil
	} else {
		allowDHCP = make(map[string]bool)
		for _, ip := range ips {
			allowDHCP[ip] = true
		}
	}
	ips = strings.Split(datastore.ReportConf.AllowMail, ",")
	if len(ips) < 1 || ips[0] == "" {
		allowMail = nil
	} else {
		allowMail = make(map[string]bool)
		for _, ip := range ips {
			allowMail[ip] = true
		}
	}
	ips = strings.Split(datastore.ReportConf.AllowLDAP, ",")
	if len(ips) < 1 || ips[0] == "" {
		allowLDAP = nil
	} else {
		allowLDAP = make(map[string]bool)
		for _, ip := range ips {
			allowLDAP[ip] = true
		}
	}
	allowLocalIP = nil
	p := strings.TrimSpace(datastore.ReportConf.AllowLocalIP)
	if p != "" {
		if strings.HasSuffix(p, "*") {
			p = strings.ReplaceAll(p, ".", "\\.")
			p = "^" + strings.ReplaceAll(p, "*", ".*")
		}
		if reg, err := regexp.Compile(p); err == nil {
			allowLocalIP = reg
		} else {
			log.Printf("UpdateReportConf err=%v", err)
		}
	}
	if !datastore.ReportConf.IncludeNoMACIP {
		datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
			if i.MAC == "" {
				datastore.DeleteReport("ips", i.IP)
			}
			return true
		})
	}
}

func reportBackend(ctx context.Context) {
	timer := time.NewTicker(time.Minute * 5)
	checkOldReport()
	calcScore()
	checkCerts()
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
				checkCertCh <- true
				log.Printf("start calc report score")
				checkOldReport()
				setSensorState()
				calcScore()
				datastore.SaveReport(last)
				last = time.Now().UnixNano()
				log.Printf("end calc report score")
				clearIpToNameCache()
				log.Printf("twpcap report stats=%d etherType=%d ipToMac=%d dns=%d dhcp=%d ntp=%d tls=%d radius=%d otehr=%d",
					statsCount, etherTypeCount, ipToMacCount, dnsCount, dhcpCount, ntpCount, tlsCount, radiusCount, otherCount)
			}
		case dr := <-deviceReportCh:
			checkDeviceReport(dr)
		case ur := <-userReportCh:
			checkUserReport(ur)
		case fr := <-flowReportCh:
			checkFlowReport(fr)
		case twpcap := <-twpcapReportCh:
			checkTWPCAPReport(twpcap)
		case twwinlog := <-twWinLogCh:
			checkTWWinLogReport(twwinlog)
		case <-checkCertCh:
			checkCerts()
		}
	}
}

func checkOldReport() {
	safeOld := time.Now().Add(time.Hour * time.Duration(-1*datastore.ReportConf.RetentionTimeForSafe)).UnixNano()
	delOld := time.Now().AddDate(0, 0, -datastore.MapConf.LogDays).UnixNano()
	log.Println("start check old report")
	checkOldServers(safeOld, delOld)
	checkOldFlows(safeOld, delOld)
	checkOldDevices(safeOld, delOld)
	checkOldIPReport(safeOld, delOld)
	checkOldUsers(delOld)
	checkOldEtherType(delOld)
	checkOldDNSQ(delOld)
	checkOldRadiusFlow(safeOld, delOld)
	checkOldTLSFlow(safeOld, delOld)
	log.Println("end check old report")
}

func checkOldServers(safeOld, delOld int64) {
	count := 0
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		if s.LastTime < safeOld {
			if s.LastTime < delOld || s.Score > 50.0 || s.Count < 10 {
				datastore.DeleteReport("servers", s.ID)
				count++
			}
		}
		return true
	})
	if count > 0 {
		log.Printf("report delete severs=%d", count)
	}
}

func checkOldFlows(safeOld, delOld int64) {
	count := 0
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		if f.LastTime < safeOld {
			if f.LastTime < delOld || f.Score > 50.0 || f.Count < 10 {
				datastore.DeleteReport("flows", f.ID)
				count++
			}
		}
		return true
	})
	if count > 0 {
		log.Printf("report delete flows=%d", count)
	}
}

func checkOldDevices(safeOld, delOld int64) {
	count := 0
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if d.LastTime < safeOld {
			if d.LastTime < delOld || (d.Score > 50.0 && d.LastTime == d.FirstTime) {
				datastore.DeleteReport("devices", d.ID)
				count++
			}
		}
		return true
	})
	if count > 0 {
		log.Printf("report delete devices=%d", count)
	}
}

func checkOldIPReport(safeOld, delOld int64) {
	count := 0
	datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
		if i.LastTime < safeOld {
			if i.LastTime < delOld || (i.Score > 50.0 && i.LastTime == i.FirstTime) {
				datastore.DeleteReport("ips", i.IP)
				count++
			}
		}
		return true
	})
	if count > 0 {
		log.Printf("report delete ip=%d", count)
	}
}

func checkOldUsers(delOld int64) {
	count := 0
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		if u.LastTime < delOld {
			datastore.DeleteReport("users", u.ID)
			count++
		}
		return true
	})
	if count > 0 {
		log.Printf("delete users=%d", count)
	}
}

func checkOldEtherType(delOld int64) {
	count := 0
	datastore.ForEachEtherType(func(u *datastore.EtherTypeEnt) bool {
		if u.LastTime < delOld {
			datastore.DeleteReport("ether", u.ID)
			count++
		}
		return true
	})
	if count > 0 {
		log.Printf("delete etherType=%d", count)
	}
}

func checkOldDNSQ(delOld int64) {
	count := 0
	datastore.ForEachDNSQ(func(u *datastore.DNSQEnt) bool {
		if u.LastTime < delOld {
			datastore.DeleteReport("dns", u.ID)
			count++
		}
		return true
	})
	if count > 0 {
		log.Printf("delete DNSQ=%d", count)
	}
}

func checkOldRadiusFlow(safeOld, delOld int64) {
	count := 0
	datastore.ForEachRADIUSFlows(func(i *datastore.RADIUSFlowEnt) bool {
		if i.LastTime < safeOld {
			if i.LastTime < delOld || (i.Score > 50.0 && i.LastTime == i.FirstTime) {
				datastore.DeleteReport("radius", i.ID)
				count++
			}
		}
		return true
	})
	if count > 0 {
		log.Printf("report delete radiusFlow=%d", count)
	}
}

func checkOldTLSFlow(safeOld, delOld int64) {
	count := 0
	datastore.ForEachTLSFlows(func(i *datastore.TLSFlowEnt) bool {
		if i.LastTime < safeOld {
			if i.LastTime < delOld || (i.Score > 50.0 && i.LastTime == i.FirstTime) {
				datastore.DeleteReport("tls", i.ID)
				count++
			}
		}
		return true
	})
	if count > 0 {
		log.Printf("report delete tlsFlow=%d", count)
	}
}

func calcScore() {
	calcDeviceScore()
	calcServerScore()
	calcFlowScore()
	calcUserScore()
	calcIPReportScore()
	calcTLSFlowScore()
	calcRADIUSFlowScore()
	calcCertScore()
}

func calcDeviceScore() {
	var xs []float64
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if ip := datastore.GetIPReport(d.IP); ip != nil && ip.Penalty > 0 {
			d.Penalty++
		}
		if d.Penalty > 100 {
			d.Penalty = 100
		}
		xs = append(xs, float64(100-d.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if sd != 0 {
			d.Score = ((10 * (float64(100-d.Penalty) - m) / sd) + 50)
		} else {
			d.Score = 50.0
		}
		d.ValidScore = true
		return true
	})
}

func calcIPReportScore() {
	var xs []float64
	datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
		if i.Penalty > 100 {
			i.Penalty = 100
		}
		xs = append(xs, float64(100-i.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
		if sd != 0 {
			i.Score = ((10 * (float64(100-i.Penalty) - m) / sd) + 50)
		} else {
			i.Score = 50.0
		}
		i.ValidScore = true
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
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		if sd != 0 {
			f.Score = ((10 * (float64(100-f.Penalty) - m) / sd) + 50)
		} else {
			f.Score = 50.0
		}
		f.ValidScore = true
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
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		if sd != 0 {
			u.Score = ((10 * (float64(100-u.Penalty) - m) / sd) + 50)
		} else {
			u.Score = 50.0
		}
		u.ValidScore = true
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
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		if sd != 0 {
			s.Score = ((10 * (float64(100-s.Penalty) - m) / sd) + 50)
		} else {
			s.Score = 50.0
		}
		s.ValidScore = true
		return true
	})
}

func calcRADIUSFlowScore() {
	var xs []float64
	datastore.ForEachRADIUSFlows(func(e *datastore.RADIUSFlowEnt) bool {
		if e.Penalty > 100 {
			e.Penalty = 100
		}
		xs = append(xs, float64(100-e.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachRADIUSFlows(func(e *datastore.RADIUSFlowEnt) bool {
		if sd != 0 {
			e.Score = ((10 * (float64(100-e.Penalty) - m) / sd) + 50)
		} else {
			e.Score = 50.0
		}
		e.ValidScore = true
		return true
	})
}

func calcTLSFlowScore() {
	var xs []float64
	datastore.ForEachTLSFlows(func(e *datastore.TLSFlowEnt) bool {
		if e.Penalty > 100 {
			e.Penalty = 100
		}
		xs = append(xs, float64(100-e.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachTLSFlows(func(e *datastore.TLSFlowEnt) bool {
		if sd != 0 {
			e.Score = ((10 * (float64(100-e.Penalty) - m) / sd) + 50)
		} else {
			e.Score = 50.0
		}
		e.ValidScore = true
		return true
	})
}

func calcCertScore() {
	var xs []float64
	datastore.ForEachCerts(func(e *datastore.CertEnt) bool {
		if e.Penalty > 100 {
			e.Penalty = 100
		}
		xs = append(xs, float64(100-e.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachCerts(func(e *datastore.CertEnt) bool {
		if sd != 0 {
			e.Score = ((10 * (float64(100-e.Penalty) - m) / sd) + 50)
		} else {
			e.Score = 50.0
		}
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
	switch report {
	case "devices":
		datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
			d.Penalty = 0
			setDevicePenalty(d)
			d.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcDeviceScore()
	case "users":
		datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
			u.Penalty = 0
			u.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcUserScore()
	case "servers":
		datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
			if s.Loc == "" {
				s.Loc = datastore.GetLoc(s.Server)
			}
			setServerPenalty(s)
			s.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcServerScore()
	case "flows":
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
	case "radius":
		datastore.ForEachRADIUSFlows(func(e *datastore.RADIUSFlowEnt) bool {
			setRADIUSFlowPenalty(e)
			e.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcRADIUSFlowScore()
	case "tls":
		datastore.ForEachTLSFlows(func(f *datastore.TLSFlowEnt) bool {
			if f.ServerLoc == "" {
				f.ServerLoc = datastore.GetLoc(f.Server)
			}
			if f.ClientLoc == "" {
				f.ClientLoc = datastore.GetLoc(f.Client)
			}
			setTLSFlowPenalty(f)
			f.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcTLSFlowScore()
	case "cert":
		datastore.ForEachCerts(func(c *datastore.CertEnt) bool {
			setCertPenalty(c)
			c.UpdateTime = time.Now().UnixNano()
			return true
		})
		calcCertScore()
	}
}

// utils
func normMACAddr(m string) string {
	if hw, err := net.ParseMAC(m); err == nil {
		m = strings.ToUpper(hw.String())
		return m
	}
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

type ipToNameCacheEnt struct {
	Name      string
	NodeID    string
	TimeLimit int64
}

var ipToNameCache sync.Map
var hitCache = 0

func findNodeInfoFromIP(ip string) (string, string) {
	if v, ok := ipToNameCache.Load(ip); ok {
		if c, ok := v.(*ipToNameCacheEnt); ok {
			hitCache++
			return c.Name, c.NodeID
		}
	}
	n := datastore.FindNodeFromIP(ip)
	if n != nil {
		addIpToNameChahe(ip, n.Name, n.ID)
		return n.Name, n.ID
	}
	if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
		addIpToNameChahe(ip, names[0], "")
		return names[0], ""
	}
	addIpToNameChahe(ip, ip, "")
	return ip, ""
}

func addIpToNameChahe(ip, name, nodeID string) {
	ipToNameCache.Store(ip, &ipToNameCacheEnt{
		Name:      name,
		NodeID:    nodeID,
		TimeLimit: time.Now().Add(time.Hour * 24).Unix(),
	})
}

func clearIpToNameCache() {
	del := 0
	sz := 0
	now := time.Now().Unix()
	ipToNameCache.Range(func(k, v interface{}) bool {
		if c, ok := v.(*ipToNameCacheEnt); !ok || c.TimeLimit < now {
			del++
			ipToNameCache.Delete(k)
		} else {
			sz++
		}
		return true
	})
	log.Printf("ipToName cache size=%d,hit=%d,del=%d", sz, hitCache, del)
}

func isSafeCountry(loc string) bool {
	if loc == "" {
		return true
	}
	if !japanOnly && dennyCountries == nil {
		return true
	}
	a := strings.Split(loc, ",")
	if len(a) < 1 {
		return true
	}
	c := a[0]
	if japanOnly {
		return c == "JP"
	}
	if _, ok := dennyCountries[c]; ok {
		dennyCountries[c]++
		return false
	}
	return true
}

func isSafeService(s, ip string) bool {
	if s == "" {
		return true
	}
	if dennyServices != nil {
		if _, ok := dennyServices[s]; ok {
			dennyServices[s]++
			return false
		}
	}
	if allowDNS != nil {
		if strings.HasPrefix(s, "domain/") {
			return allowDNS[ip]
		}
	}
	if allowDHCP != nil {
		if strings.HasPrefix(s, "bootps/") {
			return allowDHCP[ip]
		}
	}
	if allowMail != nil {
		if strings.HasPrefix(s, "smtp") || strings.HasPrefix(s, "pop3") || strings.HasPrefix(s, "imap") {
			return allowMail[ip]
		}
	}
	if allowLDAP != nil {
		if strings.HasPrefix(s, "ldap") || strings.HasPrefix(s, "kerberos") {
			return allowLDAP[ip]
		}
	}
	return true
}

type AddrInfoEnt struct {
	Level string
	Title string
	Value string
}

func GetAddressInfo(addr string) *[]AddrInfoEnt {
	if _, err := net.ParseMAC(addr); err == nil {
		return getMACInfo(addr)
	}
	return getIPInfo(addr)
}

type ipInfoCache struct {
	Time   int64
	IPInfo *[]AddrInfoEnt
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

func getIPInfo(ip string) *[]AddrInfoEnt {
	if c, ok := ipInfoCacheMap[ip]; ok {
		if c.Time > time.Now().Unix()-60*60*24 {
			return c.IPInfo
		}
	}
	ret := []AddrInfoEnt{}
	if n := datastore.FindNodeFromIP(ip); n != nil {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "管理対象ノード", Value: n.Name})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "管理対象ノード", Value: "いいえ"})
	}
	if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
		for _, n := range names {
			ret = append(ret, AddrInfoEnt{Level: "info", Title: "DNSホスト名", Value: n})
		}
	} else {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: "DNSホスト名", Value: "不明"})
	}
	loc := datastore.GetLoc(ip)
	if !isSafeCountry(loc) {
		ret = append(ret, AddrInfoEnt{Level: "high", Title: "位置", Value: loc})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "位置", Value: loc})
	}
	if strings.Contains(loc, "LOCAL,") {
		ipInfoCacheMap[ip] = &ipInfoCache{
			Time:   time.Now().Unix(),
			IPInfo: &ret,
		}
		return &ret
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		client := &rdap.Client{}
		ri, err := client.QueryIP(ip)
		if err != nil {
			log.Printf("RDAP QueryIP error=%v", err)
			ret = append(ret, AddrInfoEnt{Level: "warn", Title: "RDAP:error", Value: fmt.Sprintf("%v", err)})
			return
		}
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:IP Version", Value: ri.IPVersion}) //IPバージョン
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Type", Value: ri.Type})            // 種類
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Handole", Value: ri.Handle})       //範囲
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Name", Value: ri.Name})            // 所有者
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Country", Value: ri.Country})      // 国
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Whois Server", Value: ri.Port43})  // Whoisの情報源
	}()
	rblMap := &sync.Map{}
	for i, source := range blacklists {
		wg.Add(1)
		go func(i int, source string) {
			defer wg.Done()
			rbl := godnsbl.Lookup(source, ip)
			if len(rbl.Results) > 0 && rbl.Results[0].Listed {
				rblMap.Store(source, rbl.Results[0].Text)
			} else {
				rblMap.Store(source, "")
			}
		}(i, source)
	}
	wg.Wait()
	rblMap.Range(func(key, value interface{}) bool {
		dnsbl := key.(string)
		result := value.(string)
		if result == "" {
			ret = append(ret, AddrInfoEnt{Level: "info", Title: "DNSBL:" + dnsbl, Value: "掲載なし"})
		} else {
			ret = append(ret, AddrInfoEnt{Level: "high", Title: "DNSBL:" + dnsbl, Value: result})
		}
		return true
	})
	ipInfoCacheMap[ip] = &ipInfoCache{
		Time:   time.Now().Unix(),
		IPInfo: &ret,
	}
	return &ret
}

func getMACInfo(addr string) *[]AddrInfoEnt {
	mac := normMACAddr(addr)
	ret := []AddrInfoEnt{}
	if n := datastore.FindNodeFromMAC(mac); n != nil {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "ノード名", Value: n.Name})
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "IPアドレス", Value: n.IP})
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "説明", Value: n.Descr})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "管理対象ノード", Value: "いいえ"})
	}
	ret = append(ret, AddrInfoEnt{Level: "info", Title: "ベンダー", Value: datastore.FindVendor(mac)})
	ip := findIPFromArp(mac)
	if ip == "" {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: "ARP監視", Value: "なし"})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "ARP監視", Value: ip})
	}
	d := datastore.GetDevice(mac)
	if d == nil {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: "デバイスレポート", Value: "なし"})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "デバイスレポート:IP",
			Value: d.IP})
		lvl := "info"
		if d.Score < 30.0 {
			lvl = "high"
		} else if d.Score < 50.0 {
			lvl = "warn"
		}
		ret = append(ret, AddrInfoEnt{Level: lvl, Title: "デバイスレポート:スコア",
			Value: fmt.Sprintf("%.02f", d.Score)})
		t := time.Unix(0, d.FirstTime)
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "デバイスレポート:初回",
			Value: t.Format("2006/01/02 15:04")})
		t = time.Unix(0, d.LastTime)
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "デバイスレポート:最終",
			Value: t.Format("2006/01/02 15:04")})
	}
	return &ret
}

func findIPFromArp(mac string) string {
	ip := ""
	datastore.ForEachArp(func(a *datastore.ArpEnt) bool {
		if a.MAC == mac {
			ip = a.IP
			return false
		}
		return true
	})
	return ip
}
