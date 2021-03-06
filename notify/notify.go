// Package notify : 通知処理
package notify

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	lastExecLevel int
)

func Start(ctx context.Context) error {
	lastExecLevel = -1
	go notifyBackend(ctx)
	return nil
}

func notifyBackend(ctx context.Context) {
	lastSendReport := time.Now().Add(time.Hour * time.Duration(-24))
	lastLog := fmt.Sprintf("%016x", time.Now().Add(time.Hour*time.Duration(-1)).UnixNano())
	lastLog = checkNotify(lastLog)
	timer := time.NewTicker(time.Second * 60)
	i := 0
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			i++
			if i >= datastore.NotifyConf.Interval {
				i = 0
				lastLog = checkNotify(lastLog)
			}
			if datastore.NotifyConf.Report &&
				lastSendReport.Day() != time.Now().Day() &&
				len(datastore.DBStatsLog) > 1 &&
				len(backend.MonitorDataes) > 1 {
				lastSendReport = time.Now()
				sendReport()
			}
		}
	}
}

func getLevelNum(l string) int {
	switch l {
	case "high":
		return 0
	case "low":
		return 1
	case "warn":
		return 2
	}
	return 3
}

func checkNotify(lastLog string) string {
	list := []*datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(lastLog, func(l *datastore.EventLogEnt) bool {
		list = append(list, l)
		return true
	})
	log.Printf("checkNotify lastLog=%s len=%d", lastLog, len(list))
	if len(list) > 0 {
		nl := getLevelNum(datastore.NotifyConf.Level)
		if nl == 3 {
			return fmt.Sprintf("%016x", list[0].Time)
		}
		body := []string{}
		repair := []string{}
		ti := time.Now().Add(time.Duration(-datastore.NotifyConf.Interval) * time.Minute).UnixNano()
		for _, l := range list {
			if ti > l.Time {
				continue
			}
			if datastore.NotifyConf.NotifyRepair && l.Level == "repair" {
				a := strings.Split(l.Event, ":")
				if len(a) < 2 {
					continue
				}
				// 復帰前の状態を確認する
				np := getLevelNum(a[len(a)-1])
				if np > nl {
					continue
				}
				ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
				repair = append(repair, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
				continue
			}
			np := getLevelNum(l.Level)
			if np > nl {
				continue
			}
			ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
			body = append(body, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
		}
		if len(body) > 0 {
			err := sendMail(datastore.NotifyConf.Subject, strings.Join(body, "\r\n"))
			r := ""
			if err != nil {
				log.Printf("sendMail err=%v", err)
				r = fmt.Sprintf("失敗 エラー=%v", err)
			}
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: fmt.Sprintf("通知メール送信 %s", r),
			})
		}
		if len(repair) > 0 {
			err := sendMail(datastore.NotifyConf.Subject+"(復帰)", strings.Join(repair, "\r\n"))
			r := ""
			if err != nil {
				log.Printf("sendMail err=%v", err)
				r = fmt.Sprintf("失敗 エラー=%v", err)
			}
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: fmt.Sprintf("復帰通知メール送信 %s", r),
			})
		}
		lastLog = fmt.Sprintf("%016x", list[0].Time)
	}
	return lastLog
}

func sendMail(subject, body string) error {
	if datastore.NotifyConf.MailServer == "" || datastore.NotifyConf.MailFrom == "" || datastore.NotifyConf.MailTo == "" {
		return nil
	}
	tlsconfig := &tls.Config{
		ServerName:         datastore.NotifyConf.MailServer,
		InsecureSkipVerify: datastore.NotifyConf.InsecureSkipVerify,
	}
	c, err := smtp.Dial(datastore.NotifyConf.MailServer)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.StartTLS(tlsconfig); err != nil {
		log.Printf("sendMail err=%s", err)
	}
	msv := datastore.NotifyConf.MailServer
	a := strings.SplitN(datastore.NotifyConf.MailServer, ":", 2)
	if len(a) == 2 {
		msv = a[0]
	}
	if datastore.NotifyConf.User != "" {
		auth := smtp.PlainAuth("", datastore.NotifyConf.User, datastore.NotifyConf.Password, msv)
		if err = c.Auth(auth); err != nil {
			log.Printf("sendMail err=%s", err)
			return err
		}
	}
	if err = c.Mail(datastore.NotifyConf.MailFrom); err != nil {
		log.Printf("sendMail err=%s", err)
		return err
	}
	for _, rcpt := range strings.Split(datastore.NotifyConf.MailTo, ",") {
		if err = c.Rcpt(rcpt); err != nil {
			log.Printf("sendMail err=%s", err)
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		log.Printf("sendMail err=%s", err)
		return err
	}
	defer w.Close()
	body = convNewline(body, "\r\n")
	message := makeMailMessage(datastore.NotifyConf.MailFrom, datastore.NotifyConf.MailTo, subject, body)
	_, _ = w.Write([]byte(message))
	_ = c.Quit()
	log.Printf("Send Mail to %s", datastore.NotifyConf.MailTo)
	return nil
}

func convNewline(str, nlcode string) string {
	return strings.NewReplacer(
		"\r\n", nlcode,
		"\r", nlcode,
		"\n", nlcode,
	).Replace(str)
}

func SendTestMail(testConf *datastore.NotifyConfEnt) error {
	tlsconfig := &tls.Config{
		ServerName:         testConf.MailServer,
		InsecureSkipVerify: testConf.InsecureSkipVerify,
	}
	c, err := smtp.Dial(testConf.MailServer)
	if err != nil {
		log.Printf("SendTestMail err=%s", err)
		return err
	}
	defer c.Close()
	if err = c.StartTLS(tlsconfig); err != nil {
		log.Printf("SendTestMail err=%s", err)
	}
	msv := testConf.MailServer
	a := strings.SplitN(testConf.MailServer, ":", 2)
	if len(a) == 2 {
		msv = a[0]
	}
	if testConf.User != "" {
		auth := smtp.PlainAuth("", testConf.User, testConf.Password, msv)
		if err = c.Auth(auth); err != nil {
			log.Printf("SendTestMail err=%s", err)
			return err
		}
	}
	if err = c.Mail(testConf.MailFrom); err != nil {
		log.Printf("SendTestMail err=%s", err)
		return err
	}
	for _, rcpt := range strings.Split(testConf.MailTo, ",") {
		if err = c.Rcpt(rcpt); err != nil {
			log.Printf("SendTestMail err=%s", err)
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		log.Printf("SendTestMail err=%s", err)
		return err
	}
	defer w.Close()
	body := "Test Mail.\r\n試験メール.\r\n"
	message := makeMailMessage(testConf.MailFrom, testConf.MailTo, testConf.Subject, body)
	_, _ = w.Write([]byte(message))
	_ = c.Quit()
	return nil
}

func makeMailMessage(from, to, subject, body string) string {
	var header bytes.Buffer
	header.WriteString("From: " + from + "\r\n")
	header.WriteString("To: " + to + "\r\n")
	header.WriteString(encodeSubject(subject))
	header.WriteString("MIME-Version: 1.0\r\n")
	header.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	header.WriteString("Content-Transfer-Encoding: base64\r\n")

	var message bytes.Buffer = header
	message.WriteString("\r\n")
	message.WriteString(add76crlf(base64.StdEncoding.EncodeToString([]byte(body))))

	return message.String()
}

// 76バイト毎にCRLFを挿入する
func add76crlf(msg string) string {
	var buffer bytes.Buffer
	for k, c := range strings.Split(msg, "") {
		buffer.WriteString(c)
		if k%76 == 75 {
			buffer.WriteString("\r\n")
		}
	}
	return buffer.String()
}

// UTF8文字列を指定文字数で分割
func utf8Split(utf8string string, length int) []string {
	resultString := []string{}
	var buffer bytes.Buffer
	for k, c := range strings.Split(utf8string, "") {
		buffer.WriteString(c)
		if k%length == length-1 {
			resultString = append(resultString, buffer.String())
			buffer.Reset()
		}
	}
	if buffer.Len() > 0 {
		resultString = append(resultString, buffer.String())
	}
	return resultString
}

// サブジェクトをMIMEエンコードする
func encodeSubject(subject string) string {
	var buffer bytes.Buffer
	buffer.WriteString("Subject:")
	for _, line := range utf8Split(subject, 13) {
		buffer.WriteString(" =?utf-8?B?")
		buffer.WriteString(base64.StdEncoding.EncodeToString([]byte(line)))
		buffer.WriteString("?=\r\n")
	}
	return buffer.String()
}

func sendReport() {
	body := []string{}
	logs := []string{}
	body = append(body, "【現在のマップ情報】")
	body = append(body, getMapInfo()...)
	body = append(body, "")
	body = append(body, "【データベース情報】")
	body = append(body, getDBInfo()...)
	body = append(body, "")
	body = append(body, "【システムリソース情報】(Min/Mean/Max)")
	body = append(body, getResInfo()...)
	body = append(body, "")
	list := []*datastore.EventLogEnt{}
	datastore.ForEachLastEventLog("", func(l *datastore.EventLogEnt) bool {
		list = append(list, l)
		return true
	})
	high := 0
	low := 0
	warn := 0
	normal := 0
	other := 0
	if len(list) > 0 {
		ti := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
		for _, l := range list {
			if ti > l.Time {
				continue
			}
			switch l.Level {
			case "high":
				high++
			case "low":
				low++
			case "warn":
				warn++
			case "normal", "repair":
				normal++
			default:
				other++
			}
			ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
			logs = append(logs, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
		}
	}
	n, b := getDeviceReport()
	body = append(body, "【信用スコアが下位10%のデバイス】")
	body = append(body, b...)
	body = append(body, "")
	body = append(body, "【48時間以内に新しく発見したデバイス】")
	body = append(body, n...)
	body = append(body, "")
	n, b = getUserReport()
	body = append(body, "【信用スコアが下位10%のユーザーID】")
	body = append(body, b...)
	body = append(body, "")
	body = append(body, "【48時間以内に新しく発見したユーザーID】")
	body = append(body, n...)
	body = append(body, "")
	n, b = getIPReport()
	body = append(body, "【信用スコアが下位10%のIPアドレス】")
	body = append(body, b...)
	body = append(body, "")
	body = append(body, "【24時間以内に新しく発見したIPアドレス】")
	body = append(body, n...)
	body = append(body, "")
	body = append(body, "【AI分析情報】")
	body = append(body, getAIInfo()...)
	body = append(body, "")
	body = append(body, "【24時間以内の状態別ログ件数】")
	body = append(body, fmt.Sprintf("重度=%d,軽度=%d,注意=%d,正常=%d,その他=%d", high, low, warn, normal, other))
	body = append(body, "")
	body = append(body, "【最新24時間のログ】")
	body = append(body, logs...)
	if err := sendMail(fmt.Sprintf("TWSNMP定期レポート %s", time.Now().Format(time.RFC3339)), strings.Join(body, "\r\n")); err != nil {
		log.Printf("sendMail err=%v", err)
	} else {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: "定期レポートメール送信",
		})
	}
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
		fmt.Sprintf("MAP状態=%s", state),
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
		if i.ValidScore && i.Score < 37.5 {
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
