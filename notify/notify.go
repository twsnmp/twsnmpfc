// Package notify : 通知処理
package notify

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	lastExecLevel int
)

func Start(ctx context.Context, wg *sync.WaitGroup) error {
	lastExecLevel = -1
	wg.Add(1)
	go notifyBackend(ctx, wg)
	return nil
}

func notifyBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start notify")
	lastSendReport := time.Now().Add(time.Hour * time.Duration(-24))
	lastLog := time.Now().Add(time.Hour * time.Duration(-1)).UnixNano()
	lastLog = checkNotify(lastLog)
	timer := time.NewTicker(time.Second * 60)
	i := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("stop notify")
			timer.Stop()
			return
		case <-timer.C:
			i++
			if i >= datastore.NotifyConf.Interval {
				i = 0
				lastLog = checkNotify(lastLog)
			}
			checkExecCmd()
			if datastore.NotifyConf.Report &&
				lastSendReport.Day() != time.Now().Day() &&
				len(datastore.DBStatsLog) > 1 &&
				len(datastore.MonitorDataes) > 1 {
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

func checkNotify(last int64) int64 {
	exAll := datastore.NotifySchedule[""]
	list := []*datastore.EventLogEnt{}
	lastLogTime := int64(0)
	skip := 0
	datastore.ForEachLastEventLog(last, func(l *datastore.EventLogEnt) bool {
		if lastLogTime < l.Time {
			lastLogTime = l.Time
		}
		if exAll != "" && isExcludeTime(exAll, l.Time) {
			skip++
			return true
		}
		if l.NodeID != "" {
			if sc, ok := datastore.NotifySchedule[l.NodeID]; ok && sc != "" && isExcludeTime(sc, l.Time) {
				skip++
				return true
			}
		}
		list = append(list, l)
		return true
	})
	log.Printf("check notify last=%v next=%v len=%d skip=%d", time.Unix(0, last), time.Unix(0, lastLogTime), len(list), skip)
	if len(list) > 0 {
		sendNotifyMail(list)
	}
	if lastLogTime > 0 {
		return lastLogTime
	}
	return time.Now().UnixNano()
}

var notifySchedulePat = regexp.MustCompile(`(\S+)\s+(\d{2}):(\d{2})-(\d{2}):(\d{2})`)

func isExcludeTime(sc string, t int64) bool {
	tm := time.Unix(0, t)
	wd := tm.Format("Mon")
	md := tm.Format("2")
	for _, s := range strings.Split(sc, ",") {
		a := notifySchedulePat.FindStringSubmatch(s)
		if len(a) == 6 {
			if wd == a[1] || md == a[1] || a[1] == "*" || (a[1] == "Last" && isLastDayOfMonth(tm)) {
				sh, _ := strconv.Atoi(a[2])
				sm, _ := strconv.Atoi(a[3])
				st := sh*60 + sm
				eh, _ := strconv.Atoi(a[4])
				em, _ := strconv.Atoi(a[5])
				et := eh*60 + em
				t := tm.Hour()*60 + tm.Minute()
				if st <= t && t <= et {
					return true
				}
			}
		}
	}
	return false
}

func isLastDayOfMonth(t time.Time) bool {
	lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.Local)
	return t.Day() == lastDay.Day()
}

type notifyData struct {
	failureSubject string
	failureBody    string
	repairSubject  string
	repairBody     string
}

// getNotifyData : 通知メールの本文と件名を作成する
func getNotifyData(list []*datastore.EventLogEnt, nl int) notifyData {
	fNodeMap := make(map[string]int)
	rNodeMap := make(map[string]int)
	failure := []*datastore.EventLogEnt{}
	repair := []*datastore.EventLogEnt{}
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
			rNodeMap[l.NodeName] = np
			repair = append(repair, l)
			continue
		}
		np := getLevelNum(l.Level)
		if np > nl {
			continue
		}
		fNodeMap[l.NodeName] = np
		failure = append(failure, l)
	}
	f, r := "", ""
	fs, rs := "", ""
	if len(failure) > 0 {
		f = eventLogListToString(false, failure)
		fs = datastore.NotifyConf.Subject + "(障害)"
		if datastore.NotifyConf.AddNodeName {
			fs += ":" + getNodes(fNodeMap)
		}
	}
	if len(repair) > 0 {
		r = eventLogListToString(true, repair)
		rs = datastore.NotifyConf.Subject + "(復帰)"
		if datastore.NotifyConf.AddNodeName {
			rs += ":" + getNodes(rNodeMap)
		}
	}
	return notifyData{
		failureSubject: fs,
		failureBody:    f,
		repairSubject:  rs,
		repairBody:     r,
	}
}

func getNodes(m map[string]int) string {
	nodes := []string{}
	l := 0
	for n := range m {
		nodes = append(nodes, n)
		l += len(n)
		if l > 1000 {
			break
		}
	}
	return strings.Join(nodes, ",")
}

// eventLogListToString : イベントログを通知メールの本文に変換する
func eventLogListToString(repair bool, list []*datastore.EventLogEnt) string {
	if !datastore.NotifyConf.HTMLMail {
		r := []string{}
		if repair {
			r = append(r, "【復帰リスト】")
		} else {
			r = append(r, "【障害リスト】")
		}
		for _, l := range list {
			ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
			r = append(r, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
		}
		return strings.Join(r, "\r\n")
	}
	title := datastore.NotifyConf.Subject + "(障害)"
	if repair {
		title = datastore.NotifyConf.Subject + "(復帰)"
	}
	f := template.FuncMap{
		"levelName":     levelName,
		"formatLogTime": formatLogTime,
	}
	t, err := template.New("notify").Funcs(f).Parse(datastore.LoadMailTemplate("notify"))
	if err != nil {
		return fmt.Sprintf("メール作成エラー err=%v", err)
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, map[string]interface{}{
		"Title": title,
		"URL":   datastore.NotifyConf.URL,
		"Logs":  list,
	}); err != nil {
		return fmt.Sprintf("メール作成エラー err=%v", err)
	}
	return buffer.String()
}

func levelName(s string) string {
	switch s {
	case "high":
		return "重度"
	case "low":
		return "軽度"
	case "warn":
		return "注意"
	case "normal", "up":
		return "正常"
	case "repair":
		return "復帰"
	}
	return "不明"
}

func formatLogTime(t int64) string {
	return time.Unix(0, t).Local().Format("2006/01/02 15:04:05")
}
func formatAITime(t int64) string {
	return time.Unix(t, 0).Local().Format("2006/01/02 15:04:05")
}

func formatRSSI(rssi []datastore.RSSIEnt) string {
	min, mean, max := getRSSIStats(rssi)
	return fmt.Sprintf("%.0f/%.1f/%.0f", min, mean, max)
}

func scoreClass(s float64) string {
	if s >= 50 {
		return "info"
	} else if s > 42 {
		return "warn"
	} else if s > 33 {
		return "low"
	}
	return "high"
}

func rssiClass(rssi []datastore.RSSIEnt) string {
	s := 0
	if len(rssi) > 0 {
		s = rssi[len(rssi)-1].Value
	}
	if s >= 0 {
		return "unknown"
	} else if s > -70 {
		return "info"
	} else if s > -80 {
		return "warn"
	}
	return "high"
}

func aiScoreClass(s float64) string {
	if s > 100.0 {
		s = 1.0
	} else {
		s = 100.0 - s
	}
	return scoreClass(s)
}

func formatScore(s float64) string {
	return fmt.Sprintf("%.2f", s)
}

func formatCount(i interface{}) string {
	c := int64(0)
	switch v := i.(type) {
	case int64:
		c = v
	case int:
		c = int64(v)
	case float32:
		c = int64(v)
	case float64:
		c = int64(v)
	}
	return humanize.Comma(c)
}

func SendNotifyChat(l *datastore.EventLogEnt) {
	switch datastore.NotifyConf.ChatType {
	case "discord":
		sendNotifyDiscord(l)
	}
}
