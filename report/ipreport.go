package report

import (
	"net"
	"strings"
	"time"

	"github.com/mdlayher/netx/eui64"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func ResetIPReportScore() {
	datastore.ForEachIPReport(func(ip *datastore.IPReportEnt) bool {
		ip.Name, ip.NodeID = findNodeInfoFromIP(ip.IP)
		if ip.Name == ip.IP {
			if n := datastore.FindNodeFromMAC(ip.MAC); n != nil {
				ip.Name = n.Name
			}
		}
		setIPReportPenalty(ip)
		ip.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcIPReportScore()
}

func checkIPReport(ip, mac string, t int64) {
	if mac == "" {
		mac = getMACFromIPv6Addr(ip)
	}
	if mac == "" && !datastore.ReportConf.IncludeNoMACIP {
		// MACアドレスが不明なものは記録しないモード
		return
	}
	i := datastore.GetIPReport(ip)
	if i != nil {
		if t < i.LastTime {
			return
		}
		if mac != "" && i.MAC != mac {
			datastore.CheckNodeAddress(ip, mac, i.MAC)
			i.MAC = mac
			i.Change++
			i.Name, i.NodeID = findNodeInfoFromIP(ip)
			if i.Name == i.IP {
				if n := datastore.FindNodeFromMAC(mac); n != nil {
					i.Name = n.Name
				}
			}
			setIPReportPenalty(i)
		}
		i.Count++
		i.LastTime = t
		i.UpdateTime = time.Now().UnixNano()
		return
	}
	i = &datastore.IPReportEnt{
		IP:         ip,
		MAC:        mac,
		Count:      1,
		Change:     0,
		Loc:        datastore.GetLoc(ip),
		Vendor:     datastore.FindVendor(mac),
		FirstTime:  t,
		LastTime:   t,
		UpdateTime: time.Now().UnixNano(),
	}
	i.Name, i.NodeID = findNodeInfoFromIP(ip)
	if i.Name == i.IP {
		if n := datastore.FindNodeFromMAC(mac); n != nil {
			i.Name = n.Name
		}
	}
	setIPReportPenalty(i)
	datastore.AddIPReport(i)
	datastore.CheckNodeAddress(ip, mac, "")
}

func setIPReportPenalty(i *datastore.IPReportEnt) {
	i.Penalty = 0
	// ベンダー禁止のもの
	if i.Vendor == "Unknown" {
		i.Penalty++
	}
	// ホスト名が不明なもの
	if i.IP == i.Name {
		i.Penalty++
	}
	// 禁止の国
	if !isSafeCountry(i.Loc) {
		i.Penalty++
	}
	if i.MAC == "" {
		return
	}
	// 頻繁にMACアドレスが変わる場合
	if float64(i.Change)/float64(i.Count) > 0.1 {
		i.Penalty++
	}
	if allowLocalIP != nil && !strings.HasPrefix(i.IP, "fe80:") {
		if !allowLocalIP.MatchString(i.IP) {
			i.Penalty++
		}
	}
}

func getMACFromIPv6Addr(s string) string {
	if !strings.Contains(s, "ff:fe") {
		return ""
	}
	ip := net.ParseIP(s)
	// Retrieve IPv6 prefix and MAC address from IPv6 address
	_, mac, err := eui64.ParseIP(ip)
	if err == nil {
		return normMACAddr(mac.String())
	}
	return ""
}

func checkOldIPReport() {
	ids := []string{}
	ht := time.Now().AddDate(0, 0, -2).UnixNano()
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
		if i.LastTime < delOld || (datastore.ReportConf.AICleanup && i.LastTime < ht && aiCleanup(i.Count, i.FirstTime, i.LastTime)) {
			ids = append(ids, i.IP)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("ips", ids)
	}
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
