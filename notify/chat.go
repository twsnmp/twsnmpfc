// Package notify : 通知処理
package notify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

type discordMsg struct {
	Content string `json:"content"`
}

func SendChat(c *datastore.NotifyConfEnt, message string) error {
	if c.ChatType != "discord" || c.ChatWebhookURL == "" {
		return fmt.Errorf("invalid chat params")
	}
	m := discordMsg{
		Content: message,
	}
	j, err := json.Marshal(m)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		c.ChatWebhookURL,
		strings.NewReader(string(j)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 3)
	if string(r) != "" {
		return fmt.Errorf("%s", r)
	}
	return nil
}

func canSendChat() bool {
	if datastore.NotifyConf.ChatType != "discord" || datastore.NotifyConf.ChatWebhookURL == "" {
		return false
	}
	return true
}

func sendNotifyChat(list []*datastore.EventLogEnt) {
	if !canSendChat() {
		return
	}
	nl := getLevelNum(datastore.NotifyConf.Level)
	if nl == 3 {
		return
	}
	ti := time.Now().Add(time.Duration(-datastore.NotifyConf.Interval) * time.Minute).UnixNano()
	f := 0
	r := 0
	e := 0
	for i := len(list) - 1; i >= 0; i-- {
		l := list[i]
		if ti > l.Time {
			continue
		}
		if datastore.NotifyConf.NotifyRepair {
			if l.Level == "repair" {
				a := strings.Split(l.Event, ":")
				if len(a) < 2 {
					continue
				}
				// 復帰前の状態を確認する
				np := getLevelNum(a[len(a)-1])
				if np > nl {
					continue
				}
				// 復帰を通知する
				if err := SendChat(&datastore.NotifyConf, getChatMessage(l, true)); err != nil {
					log.Printf("send discord error=%v", err)
					e++
					continue
				}
				r++
				continue
			}
		}
		np := getLevelNum(l.Level)
		if np > nl {
			continue
		}
		if err := SendChat(&datastore.NotifyConf, getChatMessage(l, false)); err != nil {
			log.Printf("send discord error=%v", err)
			e++
			continue
		}
		f++
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf("通知チャットメッセージ送信 障害=%d 復帰=%d 送信エラー=%d", f, r, e),
	})
}

func getChatMessage(l *datastore.EventLogEnt, repair bool) string {
	subtitle := "障害"
	if repair {
		subtitle = "復帰"
	}
	return fmt.Sprintf(
		`%s(%s)
		
		発生日時: %s
		状態: %s
		タイプ: %s
		関連ノード: %s
		イベント: %s
		`,
		datastore.NotifyConf.Subject,
		subtitle,
		formatLogTime(l.Time),
		levelName(l.Level),
		l.Type,
		l.NodeName,
		l.Event,
	)
}
