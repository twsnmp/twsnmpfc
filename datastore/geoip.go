package datastore

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/oschwald/geoip2-golang"
)

// UpdateGeoIP : GeoIP DBを更新する
func UpdateGeoIP(path string) error {
	DeleteGeoIP()
	dst := filepath.Join(dspath, "geoip.mmdb")
	if err := os.Rename(path, dst); err != nil {
		return err
	}
	return openGeoIP(dst)
}

// DeleteGeoIP : GeoIP DBを削除する
func DeleteGeoIP() {
	closeGeoIP()
	dst := filepath.Join(dspath, "geoip.mmdb")
	if _, err := os.Stat(dst); err == nil {
		os.Remove(dst)
	}
}

func openGeoIP(path string) error {
	var err error
	geoip, err = geoip2.Open(path)
	if err != nil {
		log.Printf("open geoip err=%v", err)
	} else {
		md := geoip.Metadata()
		MapConf.GeoIPInfo = fmt.Sprintf("%d.%d", md.BinaryFormatMajorVersion, md.BinaryFormatMinorVersion)
	}
	return err
}

func closeGeoIP() {
	if geoip != nil {
		geoip.Close()
	}
	geoip = nil
	MapConf.GeoIPInfo = ""
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
	if !ip.IsGlobalUnicast() {
		return false
	}
	if ip.To4() == nil {
		return true
	}
	last := make(net.IP, len(ip.To4()))
	mask := ip.DefaultMask()
	binary.BigEndian.PutUint32(last, binary.BigEndian.Uint32(ip.To4())|^binary.BigEndian.Uint32(net.IP(mask).To4()))
	return !ip.Equal(last)
}
