package report

import (
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

func checkDeviceReport(dr *deviceReportEnt) {
	ip := dr.IP
	mac := dr.MAC
	d := datastore.GetDevice(mac)
	if d != nil {
		if d.IP != ip {
			d.IP = ip
			d.Name = findNameFromIP(ip)
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
		Name:       findNameFromIP(ip),
		Vendor:     datastore.FindVendor(mac),
		FirstTime:  dr.Time,
		LastTime:   dr.Time,
		UpdateTime: time.Now().UnixNano(),
	}
	setDevicePenalty(d)
	datastore.AddDevice(d)
}

func setDevicePenalty(d *datastore.DeviceEnt) {
	// ベンダー禁止のもの
	if d.Vendor == "Unknown" {
		d.Penalty++
	}
	// ホスト名が不明なもの
	if d.IP == d.Name {
		d.Penalty++
	}
	ip := net.ParseIP(d.IP)
	if !datastore.IsPrivateIP(ip) {
		d.Penalty++
	}
}
