// Package notify : 通知処理
package notify

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/twsnmp/twsnmpfc/backend"
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
	lastLog := fmt.Sprintf("%016x", time.Now().Add(time.Hour*time.Duration(-1)).UnixNano())
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
	log.Printf("check notify last=%s len=%d", lastLog, len(list))
	if len(list) > 0 {
		nl := getLevelNum(datastore.NotifyConf.Level)
		if nl == 3 {
			return fmt.Sprintf("%016x", list[0].Time)
		}
		body, repair := getNotifyBody(list, nl)
		if len(body) > 0 {
			err := sendMail(datastore.NotifyConf.Subject+"(障害)", body)
			r := ""
			if err != nil {
				log.Printf("send mail err=%v", err)
				r = fmt.Sprintf("失敗 エラー=%v", err)
			}
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: fmt.Sprintf("通知メール送信 %s", r),
			})
		}
		if len(repair) > 0 {
			err := sendMail(datastore.NotifyConf.Subject+"(復帰)", repair)
			r := ""
			if err != nil {
				log.Printf("send mail err=%v", err)
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

//
func getNotifyBody(list []*datastore.EventLogEnt, nl int) (string, string) {
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
			repair = append(repair, l)
			continue
		}
		np := getLevelNum(l.Level)
		if np > nl {
			continue
		}
		failure = append(failure, l)
	}
	return eventLogListToString(false, failure), eventLogListToString(true, repair)
}

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
