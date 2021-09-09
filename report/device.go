package report

import (
	"net"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

type deviceReportEnt struct {
	Time int64
	MAC  string
	IP   string
}

func ReportDevice(mac, ip string, t int64) {
	mac = normMACAddr(mac)
	deviceReportCh <- &deviceReportEnt{
		Time: t,
		MAC:  mac,
		IP:   ip,
	}
}

func ResetDevicesScore() {
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		d.Name, d.NodeID = findNodeInfoFromIP(d.IP)
		setDevicePenalty(d)
		d.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcDeviceScore()
}

func checkDeviceReport(dr *deviceReportEnt) {
	ip := dr.IP
	mac := dr.MAC
	checkIPReport(ip, mac, dr.Time)
	if strings.Contains(ip, ":") {
		// skip IPv6
		return
	}
	updateDeviceReport(mac, ip, dr.Time)
}

func updateDeviceReport(mac, ip string, t int64) {
	d := datastore.GetDevice(mac)
	if d != nil {
		if t < d.LastTime {
			return
		}
		if d.IP != ip {
			d.IP = ip
			d.Name, d.NodeID = findNodeInfoFromIP(ip)
			setDevicePenalty(d)
			// IPアドレスが変わるもの
			d.Penalty++
		}
		d.LastTime = t
		d.UpdateTime = time.Now().UnixNano()
		return
	}
	d = &datastore.DeviceEnt{
		ID:         mac,
		IP:         ip,
		Vendor:     datastore.FindVendor(mac),
		FirstTime:  t,
		LastTime:   t,
		UpdateTime: time.Now().UnixNano(),
	}
	d.Name, d.NodeID = findNodeInfoFromIP(ip)
	setDevicePenalty(d)
	datastore.AddDevice(d)
}

func setDevicePenalty(d *datastore.DeviceEnt) {
	d.Penalty = 0
	// ベンダー禁止のもの
	if d.Vendor == "Unknown" {
		d.Penalty++
	}
	// ホスト名が不明なもの
	if d.IP == d.Name {
		d.Penalty++
	}
	// 使用してよいローカルIP
	if allowLocalIP != nil {
		if !allowLocalIP.MatchString(d.IP) {
			d.Penalty++
		}
	}
	ip := net.ParseIP(d.IP)
	if !datastore.IsPrivateIP(ip) {
		d.Penalty++
	}
}
