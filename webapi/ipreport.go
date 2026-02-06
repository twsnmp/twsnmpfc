package webapi

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getIPReport(c echo.Context) error {
	r := []*datastore.IPReportEnt{}
	datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
		if i.ValidScore {
			r = append(r, i)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

type WebAPIIPv6Ent struct {
	Node   string
	IPv4   string
	IPv6   string
	MAC    string
	Vendor string
}

func getIPv6Report(c echo.Context) error {
	r := []*WebAPIIPv6Ent{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if n.IPv6 != "" {
			ipv6s := strings.Split(n.IPv6, ",")
			mac := n.MAC
			vendor := ""
			if idx := strings.Index(mac, "("); idx > 0 {
				vendor = mac[idx+1:]
				vendor = strings.TrimSuffix(vendor, ")")
				mac = mac[:idx]
			}
			for _, ipv6 := range ipv6s {
				r = append(r, &WebAPIIPv6Ent{
					Node:   n.Name,
					IPv4:   n.IP,
					IPv6:   ipv6,
					MAC:    mac,
					Vendor: vendor,
				})
			}
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteIPReport(c echo.Context) error {
	ip := c.Param("ip")
	if ip == "all" {
		go datastore.ClearReport("ips")
	} else {
		datastore.DeleteReport("ips", []string{ip})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("IPレポートを削除しました(%s)", ip),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetIPReport(c echo.Context) error {
	report.ResetIPReportScore()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "IPレポートの信用スコアを再計算しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type IPAMRangeEnt struct {
	Range  string
	Size   int
	Used   int
	Usage  float64
	UsedIP []int
}

// getIPAM : IPAMレポートを取得
func getIPAM(c echo.Context) error {
	ret := []*IPAMRangeEnt{}
	for _, r := range strings.Split(datastore.MapConf.ArpWatchRange, ",") {
		a := strings.SplitN(r, "-", 2)
		var sIP uint32
		var eIP uint32
		if len(a) == 1 {
			// CIDR
			ip, ipnet, err := net.ParseCIDR(r)
			if err != nil {
				continue
			}
			ipv4 := ip.To4()
			if ipv4 == nil {
				continue
			}
			sIP = ip2int(ipv4)
			for eIP = sIP; ipnet.Contains(int2ip(eIP)); eIP++ {
			}
			eIP--
		} else {
			sIP = ip2int(net.ParseIP(a[0]))
			eIP = ip2int(net.ParseIP(a[1]))
		}
		if sIP >= eIP {
			continue
		}
		e := &IPAMRangeEnt{
			Range:  r,
			UsedIP: make([]int, 100),
		}
		for nIP := sIP; nIP <= eIP; nIP++ {
			ip := int2ip(nIP)
			if !ip.IsGlobalUnicast() || ip.IsMulticast() {
				continue
			}
			sa := ip.String()
			e.Size++
			if r := datastore.GetIPReport(sa); r != nil {
				e.Used++
				e.UsedIP[100*(nIP-sIP)/(eIP-sIP)]++
				continue
			}
		}
		if e.Size > 0 {
			e.Usage = (100.0 * float64(e.Used)) / float64(e.Size)
		}
		ret = append(ret, e)
	}
	return c.JSON(http.StatusOK, ret)
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func int2ip(nIP uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nIP)
	return ip
}
