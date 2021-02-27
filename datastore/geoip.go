package datastore

import (
	"fmt"
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

func openGeoIP(path string) {
	var err error
	geoip, err = geoip2.Open(path)
	if err != nil {
		log.Printf("OpenGeoIP err=%v", err)
	}
}

func GetLoc(sip string) string {
	if l, ok := geoipMap[sip]; ok {
		return l
	}
	loc := ""
	ip := net.ParseIP(sip)
	if IsPrivateIP(ip) {
		loc = "LOCAL,0,0,"
	} else {
		if geoip == nil {
			return loc
		}
		record, err := geoip.City(ip)
		if err == nil {
			loc = fmt.Sprintf("%s,%f,%f,%s", record.Country.IsoCode, record.Location.Latitude, record.Location.Longitude, record.City.Names["en"])
		} else {
			log.Printf("getLoc err=%v", err)
			loc = "LOCAL,0,0,"
		}
	}
	geoipMap[sip] = loc
	return loc
}

var privateIPBlocks []*net.IPNet

func IsPrivateIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() {
		return true
	}
	if len(privateIPBlocks) == 0 {
		for _, cidr := range []string{
			"10.0.0.0/8",     // RFC1918
			"172.16.0.0/12",  // RFC1918
			"192.168.0.0/16", // RFC1918
		} {
			_, block, err := net.ParseCIDR(cidr)
			if err == nil {
				privateIPBlocks = append(privateIPBlocks, block)
			}
		}
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func IsGlobalUnicast(ips string) bool {
	ip := net.ParseIP(ips)
	return ip.IsGlobalUnicast()
}
