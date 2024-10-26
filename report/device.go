package report

import (
	"net"
	"sort"
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

func checkOldDevices() {
	ids := []string{}
	list := []*datastore.DeviceEnt{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if d.LastTime < delOld {
			ids = append(ids, d.ID)
		} else {
			list = append(list, d)
		}
		return true
	})
	if datastore.ReportConf.Limit < len(list) {
		sort.Slice(list, func(i, j int) bool {
			p1 := list[i].Score - float64(list[i].LastTime-delOld)/(3600*24*1000*1000*1000)
			p2 := list[j].Score - float64(list[j].LastTime-delOld)/(3600*24*1000*1000*1000)
			return p1 > p2
		})
		for i := 0; i < len(list)-datastore.ReportConf.Limit; i++ {
			ids = append(ids, list[i].ID)
		}
	}
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
