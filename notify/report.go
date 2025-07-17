// Package notify : 通知処理
package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func sendReport() {
	if datastore.NotifyConf.HTMLMail {
		sendReportHTML()
	} else {
		sendReportPlain()
	}
}

func sendReportPlain() {
	body := []string{}
	body = append(body, "【現在のマップ情報】")
	body = append(body, getMapInfo(false)...)
	body = append(body, "")
	body = append(body, "【データストア情報】")
	body = append(body, getDBInfo(false)...)
	body = append(body, "")
	body = append(body, "【システムリソース情報】(Min/Mean/Max)")
	body = append(body, getResInfo(false)...)
	body = append(body, "")
	logSum, logs, _ := getLastEventLog()
	body = append(body, "【最新24時間のログ集計】")
	body = append(body, logSum...)
	body = append(body, "")
	body = append(body, "【センサー情報】")
	body = append(body, getSensorInfo()...)
	body = append(body, "")
	body = append(body, "【AI分析情報】")
	body = append(body, getAIInfo()...)
	body = append(body, "")
	nd, bd := getDeviceReport()
	nu, bu := getUserReport()
	nip, bip := getIPReport()
	if datastore.NotifyConf.NotifyNewInfo {
		body = append(body, "【48時間以内に新しく発見したデバイス】")
		body = append(body, nd...)
		body = append(body, "")
		body = append(body, "【48時間以内に新しく発見したユーザーID】")
		body = append(body, nu...)
		body = append(body, "")
		body = append(body, "【24時間以内に新しく発見したIPアドレス】")
		body = append(body, nip...)
		body = append(body, "")
		body = append(body, "【24時間以内に新しく発見したWifi AP】")
		body = append(body, getWifiAPReport()...)
		body = append(body, "")
		body = append(body, "【24時間以内に新しく発見したBluetooth デバイス】")
		body = append(body, getBlueDevcieReport()...)
		body = append(body, "")
		body = append(body, "【最新24時間の障害ログ】")
		body = append(body, logs...)
	}
	if datastore.NotifyConf.NotifyLowScore {
		body = append(body, "")
		body = append(body, "【信用スコアが下位10%のデバイス】")
		body = append(body, bd...)
		body = append(body, "")
		body = append(body, "【信用スコアが下位10%のユーザーID】")
		body = append(body, bu...)
		body = append(body, "")
		body = append(body, "【信用スコアが下位1%のIPアドレス】")
		body = append(body, bip...)
	}
	subject := fmt.Sprintf("%s(定期レポート) at %s", datastore.NotifyConf.Subject, time.Now().Format(time.RFC3339))
	if err := sendMail(subject, strings.Join(body, "\r\n")); err != nil {
		log.Printf("send report mail err=%v", err)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "high",
			Event: "定期レポートメールの送信に失敗しました",
		})
	} else {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: "定期レポートメール送信",
		})
	}
}

func getLastEventLog() ([]string, []string, []*datastore.EventLogEnt) {
	sum := []string{}
	slogs := []string{}
	logs := []*datastore.EventLogEnt{}
	high := 0
	low := 0
	warn := 0
	normal := 0
	other := 0
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
		if l.Time < st {
			return false
		}
		switch l.Level {
		case "high":
			high++
		case "low":
			low++
		case "warn":
			warn++
			return true
		case "normal", "repair":
			normal++
			return true
		default:
			other++
			return true
		}
		ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
		slogs = append(slogs, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
		logs = append(logs, l)
		return true
	})
	sum = append(sum,
		fmt.Sprintf("重度=%d,軽度=%d,注意=%d,正常=%d,その他=%d", high, low, warn, normal, other))
	return sum, slogs, logs
}

func getMapInfo(htmlMode bool) []string {
	high := 0
	low := 0
	warn := 0
	normal := 0
	repair := 0
	unknown := 0
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		switch n.State {
		case "high":
			high++
		case "low":
			low++
		case "warn":
			warn++
		case "normal":
			normal++
		case "repair":
			repair++
		default:
			unknown++
		}
		return true
	})
	state := "不明"
	class := "none"
	if high > 0 {
		state = "重度"
		class = "high"
	} else if low > 0 {
		state = "軽度"
		class = "low"
	} else if warn > 0 {
		class = "warn"
		state = "注意"
	} else if normal+repair > 0 {
		class = "normal"
		state = "正常"
	}
	if htmlMode {
		return []string{
			datastore.MapConf.MapName,
			state,
			fmt.Sprintf("重度=%d,軽度=%d,注意=%d,復帰=%d,正常=%d,不明=%d", high, low, warn, repair, normal, unknown),
			class,
		}
	}
	return []string{
		fmt.Sprintf("マップ名=%s", datastore.MapConf.MapName),
		fmt.Sprintf("マップ状態=%s", state),
		fmt.Sprintf("重度=%d,軽度=%d,注意=%d,復帰=%d,正常=%d,不明=%d", high, low, warn, repair, normal, unknown),
	}
}

func getDeviceReport() ([]string, []string) {
	st := time.Now().Add(time.Duration(-48) * time.Hour).UnixNano()
	retNew := []string{}
	retBad := []string{}
	retNew = append(retNew, "Name,Score,IP,MAC,Vendor,Time")
	retBad = append(retBad, "Name,Score,IP,MAC,Vendor,Time")
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if d.FirstTime >= st {
			t := time.Unix(0, d.FirstTime)
			retNew = append(retNew, fmt.Sprintf("%s,%.2f,%s,%s,%s,%s", d.Name, d.Score, d.IP, d.ID, d.Vendor, t.Format(time.RFC3339)))
		}
		if d.ValidScore && d.Score < 37.5 {
			t := time.Unix(0, d.FirstTime)
			retBad = append(retBad, fmt.Sprintf("%s,%.2f,%s,%s,%s,%s", d.Name, d.Score, d.IP, d.ID, d.Vendor, t.Format(time.RFC3339)))
		}
		return true
	})
	return retNew, retBad
}

func getDeviceList() ([]*datastore.DeviceEnt, []*datastore.DeviceEnt) {
	st := time.Now().Add(time.Duration(-48) * time.Hour).UnixNano()
	retNew := []*datastore.DeviceEnt{}
	retBad := []*datastore.DeviceEnt{}
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if d.FirstTime >= st {
			retNew = append(retNew, d)
		}
		if d.ValidScore && d.Score < 37.5 {
			retBad = append(retBad, d)
		}
		return true
	})
	return retNew, retBad
}

func getUserReport() ([]string, []string) {
	st := time.Now().Add(time.Duration(-48) * time.Hour).UnixNano()
	retNew := []string{}
	retBad := []string{}
	retNew = append(retNew, "User,Server,Score,Server IP,Clients,Time")
	retBad = append(retBad, "User,Server,Score,Server IP,Clients,Time")
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		if u.FirstTime >= st {
			cls := ""
			for k := range u.ClientMap {
				if cls != "" {
					cls += ";"
				}
				cls += k
			}
			t := time.Unix(0, u.FirstTime)
			retNew = append(retNew, fmt.Sprintf("%s,%s,%.2f,%s,%s,%s", u.UserID, u.ServerName, u.Score, u.Server, cls, t.Format(time.RFC3339)))
		}
		if u.ValidScore && u.Score < 37.5 {
			cls := ""
			for k := range u.ClientMap {
				if cls != "" {
					cls += ";"
				}
				cls += k
			}
			t := time.Unix(0, u.FirstTime)
			retBad = append(retBad, fmt.Sprintf("%s,%s,%.2f,%s,%s,%s", u.UserID, u.ServerName, u.Score, u.Server, cls, t.Format(time.RFC3339)))
		}
		return true
	})
	return retNew, retBad
}

func getUserList() ([]*datastore.UserEnt, []*datastore.UserEnt) {
	st := time.Now().Add(time.Duration(-48) * time.Hour).UnixNano()
	retNew := []*datastore.UserEnt{}
	retBad := []*datastore.UserEnt{}
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		if u.FirstTime >= st {
			retNew = append(retNew, u)
		}
		if u.ValidScore && u.Score < 37.5 {
			retBad = append(retBad, u)
		}
		return true
	})
	return retNew, retBad
}

func getIPReport() ([]string, []string) {
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	retNew := []string{}
	retBad := []string{}
	retNew = append(retNew, "IP,Name,Score,MAC,Loc,Time")
	datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
		if i.FirstTime >= st {
			t := time.Unix(0, i.FirstTime)
			retNew = append(retNew, fmt.Sprintf("%s,%s,%.2f,%s,%s,%s", i.IP, i.Name, i.Score, i.MAC, i.Loc, t.Format(time.RFC3339)))
		}
		if i.ValidScore && i.Score < 26.5 {
			t := time.Unix(0, i.FirstTime)
			retBad = append(retBad, fmt.Sprintf("%s,%s,%.2f,%s,%s,%s", i.IP, i.Name, i.Score, i.MAC, i.Loc, t.Format(time.RFC3339)))
		}
		return true
	})
	return retNew, retBad
}

func getIPList() ([]*datastore.IPReportEnt, []*datastore.IPReportEnt) {
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	retNew := []*datastore.IPReportEnt{}
	retBad := []*datastore.IPReportEnt{}
	datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
		if i.FirstTime >= st {
			retNew = append(retNew, i)
		}
		if i.ValidScore && i.Score < 26.5 {
			retBad = append(retBad, i)
		}
		return true
	})
	return retNew, retBad
}

func getDBInfo(htmlMode bool) []string {
	size := humanize.Bytes(uint64(datastore.DBStats.Size))
	if len(datastore.DBStatsLog) < 1 {
		return []string{
			size,
			"",
			"",
		}
	}
	dt := datastore.DBStats.Time - datastore.DBStatsLog[0].Time
	ds := datastore.DBStats.Size - datastore.DBStatsLog[0].Size
	speed := "不明"
	dt /= (1000 * 1000 * 1000)
	if dt > 0 {
		s := ds / dt
		s *= 3600 * 24
		speed = humanize.Bytes(uint64(s))
	}
	delta := humanize.Bytes(uint64(ds))
	if htmlMode {
		return []string{
			size,
			fmt.Sprintf("%s (from %s)", delta, humanize.Time(time.Unix(0, datastore.DBStatsLog[0].Time))),
			fmt.Sprintf("%s/日", speed),
		}
	}
	return []string{
		fmt.Sprintf("現在のサイズ=%s", size),
		fmt.Sprintf("増加サイズ=%s (from %s)", delta, humanize.Time(time.Unix(0, datastore.DBStatsLog[0].Time))),
		fmt.Sprintf("増加速度=%s/日", speed),
	}
}

var myMemClass = "none"
var diskClass = "none"
var loadClass = "none"

func getResInfo(htmlMode bool) []string {
	if len(datastore.MonitorDataes) < 1 {
		return []string{}
	}
	cpu := []float64{}
	mem := []float64{}
	myCPU := []float64{}
	myMem := []float64{}
	gr := []float64{}
	disk := []float64{}
	load := []float64{}
	for _, m := range datastore.MonitorDataes {
		cpu = append(cpu, m.CPU)
		mem = append(mem, m.Mem)
		myCPU = append(myCPU, m.MyCPU)
		myMem = append(myMem, m.MyMem)
		disk = append(disk, m.Disk)
		load = append(load, m.Load)
		gr = append(gr, float64(m.NumGoroutine))
	}
	cpuMin, _ := stats.Min(cpu)
	cpuMean, _ := stats.Mean(cpu)
	cpuMax, _ := stats.Max(cpu)
	memMin, _ := stats.Min(mem)
	memMean, _ := stats.Mean(mem)
	memMax, _ := stats.Max(mem)
	myCPUMin, _ := stats.Min(myCPU)
	myCPUMean, _ := stats.Mean(myCPU)
	myCPUMax, _ := stats.Max(myCPU)
	myMemMin, _ := stats.Min(myMem)
	myMemMean, _ := stats.Mean(myMem)
	myMemMax, _ := stats.Max(myMem)
	diskMin, _ := stats.Min(disk)
	diskMean, _ := stats.Mean(disk)
	diskMax, _ := stats.Max(disk)
	loadMin, _ := stats.Min(load)
	loadMean, _ := stats.Mean(load)
	loadMax, _ := stats.Max(load)
	grMin, _ := stats.Min(gr)
	grMean, _ := stats.Mean(gr)
	grMax, _ := stats.Max(gr)
	if htmlMode {
		if myMemMean > 90.0 && memMean > 90.0 {
			myMemClass = "high"
		} else if myMemMean > 80.0 && memMean > 80.0 {
			myMemClass = "low"
		} else if myMemMean > 60.0 && memMean > 60 {
			myMemClass = "warn"
		} else {
			myMemClass = "none"
		}
		if diskMean > 95.0 {
			diskClass = "high"
		} else if diskMean > 90.0 {
			diskClass = "low"
		} else if diskMean > 80.0 {
			diskClass = "warn"
		} else {
			diskClass = "none"
		}
		if loadMean > float64(runtime.NumCPU()) {
			loadClass = "high"
		} else {
			loadClass = "none"
		}
		return []string{
			fmt.Sprintf("最小:%s%% 平均:%s%% 最大:%s%%",
				humanize.FormatFloat("###.##", cpuMin),
				humanize.FormatFloat("###.##", cpuMean),
				humanize.FormatFloat("###.##", cpuMax),
			),
			fmt.Sprintf("最小:%s%% 平均:%s%% 最大:%s%%",
				humanize.FormatFloat("###.##", memMin),
				humanize.FormatFloat("###.##", memMean),
				humanize.FormatFloat("###.##", memMax),
			),
			fmt.Sprintf("最小:%s%% 平均:%s%% 最大:%s%%",
				humanize.FormatFloat("###.##", myCPUMin),
				humanize.FormatFloat("###.##", myCPUMean),
				humanize.FormatFloat("###.##", myCPUMax),
			),
			fmt.Sprintf("最小:%s%% 平均:%s%% 最大:%s%%",
				humanize.FormatFloat("###.##", myMemMin),
				humanize.FormatFloat("###.##", myMemMean),
				humanize.FormatFloat("###.##", myMemMax),
			),
			fmt.Sprintf("最小:%s%% 平均:%s%% 最大:%s%%",
				humanize.FormatFloat("###.##", diskMin),
				humanize.FormatFloat("###.##", diskMean),
				humanize.FormatFloat("###.##", diskMax),
			),
			fmt.Sprintf("最小:%s 平均:%s 最大:%s",
				humanize.FormatFloat("###.##", loadMin),
				humanize.FormatFloat("###.##", loadMean),
				humanize.FormatFloat("###.##", loadMax),
			),
			fmt.Sprintf("最小:%s 平均:%s 最大:%s",
				humanize.FormatFloat("###.##", grMin),
				humanize.FormatFloat("###.##", grMean),
				humanize.FormatFloat("###.##", grMax),
			),
		}
	}
	return []string{
		fmt.Sprintf("CPU=%s/%s/%s %%",
			humanize.FormatFloat("###.##", cpuMin),
			humanize.FormatFloat("###.##", cpuMean),
			humanize.FormatFloat("###.##", cpuMax),
		),
		fmt.Sprintf("Mem=%s/%s/%s %%",
			humanize.FormatFloat("###.##", memMin),
			humanize.FormatFloat("###.##", memMean),
			humanize.FormatFloat("###.##", memMax),
		),
		fmt.Sprintf("My CPU=%s/%s/%s %%",
			humanize.FormatFloat("###.##", myCPUMin),
			humanize.FormatFloat("###.##", myCPUMean),
			humanize.FormatFloat("###.##", myCPUMax),
		),
		fmt.Sprintf("My Mem=%s/%s/%s %%",
			humanize.FormatFloat("###.##", myMemMin),
			humanize.FormatFloat("###.##", myMemMean),
			humanize.FormatFloat("###.##", myMemMax),
		),
		fmt.Sprintf("Disk=%s/%s/%s %%",
			humanize.FormatFloat("###.##", diskMin),
			humanize.FormatFloat("###.##", diskMean),
			humanize.FormatFloat("###.##", diskMax),
		),
		fmt.Sprintf("Load=%s/%s/%s",
			humanize.FormatFloat("###.##", loadMin),
			humanize.FormatFloat("###.##", loadMean),
			humanize.FormatFloat("###.##", loadMax),
		),
		fmt.Sprintf("NumGoroutine=%s/%s/%s",
			humanize.FormatFloat("###.##", grMin),
			humanize.FormatFloat("###.##", grMean),
			humanize.FormatFloat("###.##", grMax),
		),
	}
}

func getAIInfo() []string {
	ret := []string{"Score,Node,Polling,Count"}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.LogMode != datastore.LogModeAI {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		air, err := datastore.GetAIReesult(p.ID)
		if err != nil || len(air.ScoreData) < 1 {
			return true
		}
		ret = append(ret,
			fmt.Sprintf("%.2f,%s,%s,%d",
				air.ScoreData[len(air.ScoreData)-1][1],
				n.Name,
				p.Name,
				len(air.ScoreData),
			))
		return true
	})
	return ret
}

type aiResultEnt struct {
	LastScore   float64
	NodeName    string
	PollingName string
	Count       int
	LastTime    int64
}

func getAIList() []aiResultEnt {
	ret := []aiResultEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.LogMode != datastore.LogModeAI {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		air, err := datastore.GetAIReesult(p.ID)
		if err != nil || len(air.ScoreData) < 1 {
			return true
		}
		ret = append(ret, aiResultEnt{
			LastScore:   air.ScoreData[len(air.ScoreData)-1][1],
			NodeName:    n.Name,
			PollingName: p.Name,
			Count:       len(air.ScoreData),
			LastTime:    air.LastTime,
		})
		return true
	})
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].LastScore > ret[j].LastScore
	})
	return ret
}

func getSensorInfo() []string {
	ret := []string{}
	ret = append(ret, "State,Host,Type,Params,total,Last Time")
	datastore.ForEachSensors(func(s *datastore.SensorEnt) bool {
		if s.Ignore {
			return true
		}
		t := time.Unix(0, s.LastTime)
		ret = append(ret, fmt.Sprintf("%s,%s,%s,%s,%d,%s",
			s.State, s.Host, s.Type, s.Param, s.Total, t.Format(time.RFC3339)))
		return true
	})
	return ret
}

func getSensorList() []*datastore.SensorEnt {
	ret := []*datastore.SensorEnt{}
	datastore.ForEachSensors(func(s *datastore.SensorEnt) bool {
		if s.Ignore {
			return true
		}
		ret = append(ret, s)
		return true
	})
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].State < ret[j].State
	})
	return ret
}

func getRSSIStats(rssi []datastore.RSSIEnt) (min, mean, max float64) {
	v := []float64{}
	for _, e := range rssi {
		v = append(v, float64(e.Value))
	}
	min, _ = stats.Min(v)
	mean, _ = stats.Mean(v)
	max, _ = stats.Max(v)
	return
}

func getWifiAPReport() []string {
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	m := make(map[string]bool)
	ret := []string{}
	ret = append(ret, "RSSI(min/avg/max),BSSID,SSID,Host,Channel,Count,First Time")
	datastore.ForEachWifiAP(func(a *datastore.WifiAPEnt) bool {
		if a.FirstTime < st {
			return true
		}
		if _, ok := m[a.BSSID]; ok {
			return true
		}
		m[a.BSSID] = true
		min, mean, max := getRSSIStats(a.RSSI)
		t := time.Unix(0, a.FirstTime)
		ret = append(ret, fmt.Sprintf("%.0f/%.1f/%.0f,%s,%s,%s,%s,%d,%s",
			min, mean, max, a.BSSID, a.SSID, a.Host, a.Channel, a.Count, t.Format(time.RFC3339)))
		return true
	})
	return ret
}

func getWifiAPList() []*datastore.WifiAPEnt {
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	m := make(map[string]bool)
	ret := []*datastore.WifiAPEnt{}
	datastore.ForEachWifiAP(func(a *datastore.WifiAPEnt) bool {
		if a.FirstTime < st {
			return true
		}
		if _, ok := m[a.BSSID]; ok {
			return true
		}
		m[a.BSSID] = true
		ret = append(ret, a)
		return true
	})
	return ret
}

func getBlueDevcieReport() []string {
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	m := make(map[string]bool)
	ret := []string{}
	ret = append(ret, "RSSI(min/avg/max),Address,Name,Host,Address Type,Vendor,Count,Time")
	datastore.ForEachBlueDevice(func(b *datastore.BlueDeviceEnt) bool {
		if b.FirstTime < st {
			return true
		}
		if _, ok := m[b.Address]; ok {
			return true
		}
		m[b.Address] = true
		min, mean, max := getRSSIStats(b.RSSI)
		t := time.Unix(0, b.FirstTime)
		ret = append(ret, fmt.Sprintf("%.0f/%.1f/%.0f,%s,%s,%s,%s,%s,%d,%s",
			min, mean, max, b.Address, b.Name, b.Host, b.AddressType, b.Vendor, b.Count, t.Format(time.RFC3339)))
		return true
	})
	return ret
}

func getBlueDevcieList() []*datastore.BlueDeviceEnt {
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	ret := []*datastore.BlueDeviceEnt{}
	m := make(map[string]bool)
	datastore.ForEachBlueDevice(func(b *datastore.BlueDeviceEnt) bool {
		if b.FirstTime < st {
			return true
		}
		if _, ok := m[b.Address]; ok {
			return true
		}
		m[b.Address] = true
		ret = append(ret, b)
		return true
	})
	return ret
}

type reportInfoEnt struct {
	Name  string
	Class string
	Value string
}

// HTML版レポートの送信
func sendReportHTML() {
	info := []reportInfoEnt{}
	a := getMapInfo(true)
	if len(a) > 3 {
		info = append(info, reportInfoEnt{
			Name:  "マップ名",
			Value: a[0],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "マップの状態",
			Value: a[1],
			Class: a[3],
		})
		info = append(info, reportInfoEnt{
			Name:  "状態別のノード数",
			Value: a[2],
			Class: "none",
		})
	}
	a = getDBInfo(true)
	if len(a) > 2 {
		info = append(info, reportInfoEnt{
			Name:  "データストアサイズ",
			Value: a[0],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "データストア増加量",
			Value: a[1],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "データストア増加速度",
			Value: a[2],
			Class: "none",
		})
	}
	a = getResInfo(true)
	if len(a) > 3 {
		info = append(info, reportInfoEnt{
			Name:  "CPU使用率",
			Value: a[0],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "メモリ使用率",
			Value: a[1],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "My CPU使用率",
			Value: a[2],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "My メモリー使用率",
			Value: a[3],
			Class: myMemClass,
		})
		info = append(info, reportInfoEnt{
			Name:  "ディスク使用率",
			Value: a[4],
			Class: diskClass,
		})
		info = append(info, reportInfoEnt{
			Name:  "システム負荷",
			Value: a[5],
			Class: loadClass,
		})
	}
	logSum, _, logs := getLastEventLog()
	if len(logSum) > 0 {
		info = append(info, reportInfoEnt{
			Name:  "状態別のログ数",
			Value: logSum[0],
			Class: "none",
		})
	}
	nd, bd := getDeviceList()
	nu, bu := getUserList()
	nip, bip := getIPList()
	title := fmt.Sprintf("%s(定期レポート) at %s", datastore.NotifyConf.Subject, time.Now().Format("2006/01/02 15:04:05"))
	f := template.FuncMap{
		"levelName":     levelName,
		"formatLogTime": formatLogTime,
		"formatScore":   formatScore,
		"formatRSSI":    formatRSSI,
		"scoreClass":    scoreClass,
		"aiScoreClass":  aiScoreClass,
		"formatAITime":  formatAITime,
		"rssiClass":     rssiClass,
		"formatCount":   formatCount,
	}
	t, err := template.New("report").Funcs(f).Parse(datastore.LoadMailTemplate("report"))
	if err != nil {
		log.Printf("send report mail err=%v", err)
		return
	}
	body := new(bytes.Buffer)
	if err = t.Execute(body, map[string]interface{}{
		"Title":          title,
		"URL":            datastore.NotifyConf.URL,
		"Info":           info,
		"Logs":           logs,
		"Sensors":        getSensorList(),
		"NewDevices":     nd,
		"BadDevices":     bd,
		"NewUsers":       nu,
		"BadUsers":       bu,
		"NewIPs":         nip,
		"BadIPs":         bip,
		"NewWifiAPs":     getWifiAPList(),
		"NewBlueDevcies": getBlueDevcieList(),
		"AIList":         getAIList(),
		"NotifyLowScore": datastore.NotifyConf.NotifyLowScore,
		"NotifyNewInfo":  datastore.NotifyConf.NotifyNewInfo,
	}); err != nil {
		log.Printf("send report mail err=%v", err)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "low",
			Event: fmt.Sprintf("定期レポートメール送信失敗 err=%v", err),
		})
		return
	}
	if err := sendMail(title, body.String()); err != nil {
		log.Printf("send report mail err=%v", err)
	} else {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: "定期レポートメール送信",
		})
	}
}
