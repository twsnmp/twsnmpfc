package report

import (
	"log"
	"net"
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
		setDevicePenalty(d)
		d.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcDeviceScore()
}

func checkDeviceReport(dr *deviceReportEnt) {
	ip := dr.IP
	mac := dr.MAC
	d := datastore.GetDevice(mac)
	if d != nil {
		if d.IP != ip {
			d.IP = ip
			d.Name, d.NodeID = findNodeInfoFromIP(ip)
			setDevicePenalty(d)
			// IPアドレスが変わるもの
			d.Penalty++
		}
		d.LastTime = dr.Time
		d.UpdateTime = time.Now().UnixNano()
		return
	}
	d = &datastore.DeviceEnt{
		ID:         mac,
		IP:         ip,
		Vendor:     datastore.FindVendor(mac),
		FirstTime:  dr.Time,
		LastTime:   dr.Time,
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
			log.Printf("device use not allowed IP mac=%s ip=%s", d.ID, d.IP)
			d.Penalty++
		}
	}
	ip := net.ParseIP(d.IP)
	if !datastore.IsPrivateIP(ip) {
		d.Penalty++
	}
}
