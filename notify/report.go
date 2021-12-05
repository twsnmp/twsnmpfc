// Package notify : 通知処理
package notify

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func sendReport() {
	body := []string{}
	body = append(body, "【現在のマップ情報】")
	body = append(body, getMapInfo()...)
	body = append(body, "")
	body = append(body, "【データストア情報】")
	body = append(body, getDBInfo()...)
	body = append(body, "")
	body = append(body, "【システムリソース情報】(Min/Mean/Max)")
	body = append(body, getResInfo()...)
	body = append(body, "")
	logSum, logs := getLastEventLog()
	body = append(body, "【最新24時間のログ集計】")
	body = append(body, logSum...)
	body = append(body, "")
	body = append(body, "【センサー情報】")
	body = append(body, getSensorInfo()...)
	body = append(body, "")
	nd, bd := getDeviceReport()
	body = append(body, "【48時間以内に新しく発見したデバイス】")
	body = append(body, nd...)
	body = append(body, "")
	nu, bu := getUserReport()
	body = append(body, "【48時間以内に新しく発見したユーザーID】")
	body = append(body, nu...)
	body = append(body, "")
	nip, bip := getIPReport()
	body = append(body, "【24時間以内に新しく発見したIPアドレス】")
	body = append(body, nip...)
	body = append(body, "")
	body = append(body, "【24時間以内に新しく発見したWifi AP】")
	body = append(body, getWifiAPReport()...)
	body = append(body, "")
	body = append(body, "【24時間以内に新しく発見したBluetooth デバイス】")
	body = append(body, getBlueDevcieReport()...)
	body = append(body, "")
	body = append(body, "【AI分析情報】")
	body = append(body, getAIInfo()...)
	body = append(body, "")
	body = append(body, "【最新24時間の障害ログ】")
	body = append(body, logs...)

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
	} else {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: "定期レポートメール送信",
		})
	}
}

func getLastEventLog() ([]string, []string) {
	sum := []string{}
	logs := []string{}
	high := 0
	low := 0
	warn := 0
	normal := 0
	other := 0
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	datastore.ForEachLastEventLog("", func(l *datastore.EventLogEnt) bool {
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
		logs = append(logs, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
		return true
	})
	sum = append(sum,
		fmt.Sprintf("重度=%d,軽度=%d,注意=%d,正常=%d,その他=%d", high, low, warn, normal, other))
	return sum, logs
}

func getMapInfo() []string {
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
	if high > 0 {
		state = "重度"
	} else if low > 0 {
		state = "軽度"
	} else if warn > 0 {
		state = "注意"
	} else if normal+repair > 0 {
		state = "正常"
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

func getDBInfo() []string {
	size := humanize.Bytes(uint64(datastore.DBStats.Size))
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
	return []string{
		fmt.Sprintf("現在のサイズ=%s", size),
		fmt.Sprintf("増加サイズ=%s (from %s)", delta, humanize.Time(time.Unix(0, datastore.DBStatsLog[0].Time))),
		fmt.Sprintf("増加速度=%s/日", speed),
	}
}

func getResInfo() []string {
	if len(backend.MonitorDataes) < 1 {
		return []string{}
	}
	cpu := []float64{}
	mem := []float64{}
	disk := []float64{}
	load := []float64{}
	for _, m := range backend.MonitorDataes {
		cpu = append(cpu, m.CPU)
		mem = append(mem, m.Mem)
		disk = append(disk, m.Disk)
		load = append(load, m.Load)
	}
	cpuMin, _ := stats.Min(cpu)
	cpuMean, _ := stats.Mean(cpu)
	cpuMax, _ := stats.Max(cpu)
	memMin, _ := stats.Min(mem)
	memMean, _ := stats.Mean(mem)
	memMax, _ := stats.Max(mem)
	diskMin, _ := stats.Min(disk)
	diskMean, _ := stats.Mean(disk)
	diskMax, _ := stats.Max(disk)
	loadMin, _ := stats.Min(load)
	loadMean, _ := stats.Mean(load)
	loadMax, _ := stats.Max(load)
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

func getSensorInfo() []string {
	ret := []string{}
	ret = append(ret, "State,Host,Type,Params,total,Last Time")
	datastore.ForEachSensors(func(s *datastore.SensorEnt) bool {
		t := time.Unix(0, s.LastTime)
		ret = append(ret, fmt.Sprintf("%s,%s,%s,%s,%d,%s",
			s.State, s.Host, s.Type, s.Param, s.Total, t.Format(time.RFC3339)))
		return true
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
	ret := []string{}
	ret = append(ret, "RSSI(min/avg/max),BSSID,SSID,Host,Channel,Count,First Time")
	datastore.ForEachWifiAP(func(a *datastore.WifiAPEnt) bool {
		if a.FirstTime < st {
			return true
		}
		min, mean, max := getRSSIStats(a.RSSI)
		t := time.Unix(0, a.FirstTime)
		ret = append(ret, fmt.Sprintf("%.0f/%.1f/%.0f,%s,%s,%s,%s,%d,%s",
			min, mean, max, a.BSSID, a.SSID, a.Host, a.Channel, a.Count, t.Format(time.RFC3339)))
		return true
	})
	return ret
}

func getBlueDevcieReport() []string {
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	ret := []string{}
	ret = append(ret, "RSSI(min/avg/max),Address,Name,Host,Address Type,Vendor,Count,Time")
	datastore.ForEachBludeDevice(func(b *datastore.BlueDeviceEnt) bool {
		if b.FirstTime < st {
			return true
		}
		min, mean, max := getRSSIStats(b.RSSI)
		t := time.Unix(0, b.FirstTime)
		ret = append(ret, fmt.Sprintf("%.0f/%.1f/%.0f,%s,%s,%s,%s,%s,%d,%s",
			min, mean, max, b.Address, b.Name, b.Host, b.AddressType, b.Vendor, b.Count, t.Format(time.RFC3339)))
		return true
	})
	return ret
}
