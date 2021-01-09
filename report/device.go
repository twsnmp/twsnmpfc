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

func (r *Report) ReportDevice(mac, ip string, t int64) {
	mac = normMACAddr(mac)
	r.deviceReportCh <- &deviceReportEnt{
		Time: t,
		MAC:  mac,
		IP:   ip,
	}
}

func (r *Report) checkDeviceReport(dr *deviceReportEnt) {
	ip := dr.IP
	mac := dr.MAC
	d := r.ds.GetDevice(mac)
	if d != nil {
		if d.IP != ip {
			d.IP = ip
			d.Name = r.findNameFromIP(ip)
			r.setDevicePenalty(d)
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
		Name:       r.findNameFromIP(ip),
		Vendor:     r.ds.FindVendor(mac),
		FirstTime:  dr.Time,
		LastTime:   dr.Time,
		UpdateTime: time.Now().UnixNano(),
	}
	r.setDevicePenalty(d)
	r.ds.AddDevice(d)
}

func (r *Report) setDevicePenalty(d *datastore.DeviceEnt) {
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
