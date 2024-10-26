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
	"github.com/twsnmp/rdap"
	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	deviceReportCh = make(chan *deviceReportEnt, 100)
	userReportCh   = make(chan *userReportEnt, 100)
	flowReportCh   = make(chan *flowReportEnt, 1000)
	checkCertCh    chan bool
	twpcapReportCh = make(chan map[string]interface{}, 100)
	twWinLogCh     = make(chan map[string]interface{}, 100)
	twBlueScanCh   = make(chan map[string]interface{}, 100)
	twWifiScanCh   = make(chan map[string]interface{}, 100)
	twSdrPowerCh   = make(chan map[string]interface{}, 100)
	allowDNS       map[string]bool
	allowDHCP      map[string]bool
	allowMail      map[string]bool
	allowLDAP      map[string]bool
	allowLocalIP   *regexp.Regexp
	dennyCountries map[string]int64
	dennyServices  map[string]int64
	japanOnly      bool
)

func Start(ctx context.Context, wg *sync.WaitGroup) error {
	datastore.LoadReportConf()
	UpdateReportConf()
	deviceReportCh = make(chan *deviceReportEnt, 100)
	userReportCh = make(chan *userReportEnt, 100)
	flowReportCh = make(chan *flowReportEnt, 500)
	twpcapReportCh = make(chan map[string]interface{}, 100)
	twWinLogCh = make(chan map[string]interface{}, 100)
	checkCertCh = make(chan bool, 5)
	twBlueScanCh = make(chan map[string]interface{}, 100)
	twWifiScanCh = make(chan map[string]interface{}, 100)
	twSdrPowerCh = make(chan map[string]interface{}, 100)
	wg.Add(1)
	go reportBackend(ctx, wg)
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
		}
	}
	if !datastore.ReportConf.IncludeNoMACIP {
		ids := []string{}
		datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
			if i.MAC == "" {
				ids = append(ids, i.IP)
			}
			return true
		})
		if len(ids) > 0 {
			datastore.DeleteReport("ips", ids)
		}
	}
}

func reportBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	timer5Min := time.NewTicker(time.Minute * 5)
	timer1Min := time.NewTicker(time.Minute * 1)
	log.Println("start report")
	datastore.LoadReport()
	log.Println("load report done")
	setSensorState()
	calcScore()
	checkCerts()
	last := time.Now().UnixNano()
	skip := 0
	for {
		select {
		case <-ctx.Done():
			{
				timer5Min.Stop()
				timer1Min.Stop()
				datastore.SaveReport(last)
				log.Printf("stop report")
				return
			}
		case <-timer5Min.C:
			{
				checkCertCh <- true
				st := time.Now()
				setSensorState()
				calcScore()
				skip--
				if skip < 0 {
					skip = 12
					go checkOldReport()
					datastore.SaveReport(last)
					last = time.Now().UnixNano()
				}
				clearIpToNameCache()
				log.Printf("report timer process dur=%v", time.Since(st))
			}
		case <-timer1Min.C:
			saveSdrPowerReport()
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
		case l := <-twBlueScanCh:
			checkTWBlueScanReport(l)
		case l := <-twWifiScanCh:
			checkTWWifiScanReport(l)
		case l := <-twSdrPowerCh:
			checkTWSdrPowerReport(l)
		case <-checkCertCh:
			checkCerts()
		}
	}
}

var oldCheck = false

func checkOldReport() {
	if oldCheck {
		return
	}
	oldCheck = true
	st := time.Now()
	checkOldDevices()
	checkOldServers()
	checkOldFlows()
	checkOldIPReport()
	checkOldUsers()
	checkOldEtherType()
	checkOldDNSQ()
	checkOldRadiusFlow()
	checkOldTLSFlow()
	checkOldWinEventID()
	checkOldWinLogon()
	checkOldWinAccount()
	checkOldWinKerberos()
	checkOldWinPrivilege()
	checkOldWinProcess()
	checkOldWinTask()
	checkOldBlueDevice()
	checkOldEnvMonitor()
	checkOldWifiAP()
	checkOldSdrPower()
	checkOldMotionSensor()
	log.Printf("check old report dur=%v", time.Since(st))
	oldCheck = false
}

func calcScore() {
	st := time.Now()
	calcDeviceScore()
	calcServerScore()
	calcFlowScore()
	calcUserScore()
	calcIPReportScore()
	calcTLSFlowScore()
	calcRADIUSFlowScore()
	calcCertScore()
	calcWinLogonScore()
	calcWinKerberosScore()
	log.Printf("calcreport score dur=%v", time.Since(st))
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
			if c.NodeID == "" || datastore.GetNode(c.NodeID) != nil {
				hitCache++
				return c.Name, c.NodeID
			}
		}
	}
	n := datastore.FindNodeFromIP(ip)
	if n != nil {
		addIpToNameChahe(ip, n.Name, n.ID)
		return n.Name, n.ID
	}
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*50)
	defer cancel()
	if names, err := r.LookupAddr(ctx, ip); err == nil && len(names) > 0 {
		addIpToNameChahe(ip, names[0], "")
		return names[0], ""
	}
	addIpToNameChahe(ip, ip, "")
	return ip, ""
}

func FindoHostFromIP(ip string) string {
	n, _ := findNodeInfoFromIP(ip)
	return n
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
		if c, ok := v.(*ipToNameCacheEnt); !ok || c.TimeLimit < now ||
			(c.NodeID != "" && datastore.GetNode(c.NodeID) == nil) {
			del++
			ipToNameCache.Delete(k)
		} else {
			sz++
		}
		return true
	})
	log.Printf("clear ip to name cache size=%d hit=%d del=%d", sz, hitCache, del)
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

func GetAddressInfo(addr, dnsbl, noCache string) *[]AddrInfoEnt {
	if addr == "" {
		return &[]AddrInfoEnt{}
	}
	if _, err := net.ParseMAC(addr); err == nil {
		return getMACInfo(addr)
	}
	return getIPInfo(addr, dnsbl, noCache)
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

func getIPInfo(ip, dnsbl, noCache string) *[]AddrInfoEnt {
	if noCache != "true" {
		if c, ok := ipInfoCacheMap[ip]; ok {
			if c.Time > time.Now().Unix()-60*60*24 {
				return c.IPInfo
			}
		}
	}
	ret := []AddrInfoEnt{}
	ret = append(ret, AddrInfoEnt{Level: "info", Title: "IPアドレス", Value: ip})
	if n := datastore.FindNodeFromIP(ip); n != nil {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "管理対象ノード", Value: n.Name})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "管理対象ノード", Value: "いいえ"})
	}
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*50)
	defer cancel()
	if names, err := r.LookupAddr(ctx, ip); err == nil && len(names) > 0 {
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
			log.Printf("rdap query err=%v", err)
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
	if dnsbl == "true" {
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
	ret = append(ret, AddrInfoEnt{Level: "info", Title: "MACアドレス", Value: addr})
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
