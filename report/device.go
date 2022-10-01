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

func checkOldDevices(safeOld, delOld int64) {
	ids := []string{}
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if d.LastTime < safeOld {
			if d.LastTime < delOld || (d.Score > 50.0 && d.LastTime == d.FirstTime) {
				ids = append(ids, d.ID)
			}
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("devices", ids)
	}
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
