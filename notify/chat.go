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

type discordEnt struct {
	Title       string `json:"title"`
	Color       string `json:"color"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type discordMsg struct {
	Embeds []discordEnt `json:"embeds"`
}

func SendChat(c *datastore.NotifyConfEnt, title, level, message string) error {
	if c.ChatType != "discord" || c.ChatWebhookURL == "" {
		return fmt.Errorf("invalid chat params")
	}
	color := "10070709"
	switch level {
	case "high":
		color = "15548997"
	case "low":
		color = "15418782"
	case "warn":
		color = "16705372"
	case "normal":
		color = "5763719"
	case "repair", "info":
		color = "5793266"
	}
	m := discordMsg{
		Embeds: []discordEnt{{
			Title:       title,
			Color:       color,
			Description: message,
			URL:         c.URL,
		}},
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

// SendNotifyChat : 通知のチャットメッセージを送信する
func SendNotifyChat(l *datastore.EventLogEnt) {
	if !canSendChat() {
		return
	}
	nl := getLevelNum(datastore.NotifyConf.Level)
	if nl == 3 {
		return
	}
	if l.Level == "repair" {
		if !datastore.NotifyConf.NotifyRepair {
			return
		}
		a := strings.Split(l.Event, ":")
		if len(a) < 2 {
			return
		}
		// 復帰前の状態を確認する
		np := getLevelNum(a[len(a)-1])
		if np > nl {
			return
		}
		// 復帰を通知する
		title, message := getChatMessage(l, true)
		if err := SendChat(&datastore.NotifyConf, title, "repair", message); err != nil {
			log.Printf("send discord error=%v", err)
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "system",
				Level:    "warn",
				NodeID:   l.NodeID,
				NodeName: l.NodeName,
				Event:    "復帰通知のチャットメッセージを送信できません",
			})
			return
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "system",
			Level:    "info",
			NodeID:   l.NodeID,
			NodeName: l.NodeName,
			Event:    "復帰通知のチャットメッセージを送信しました",
		})
		return
	}
	np := getLevelNum(l.Level)
	if np > nl {
		return
	}
	title, message := getChatMessage(l, false)
	if err := SendChat(&datastore.NotifyConf, title, l.Level, message); err != nil {
		log.Printf("send discord error=%v", err)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "system",
			Level:    "warn",
			NodeID:   l.NodeID,
			NodeName: l.NodeName,
			Event:    "障害通知のチャットメッセージを送信できません",
		})
		return
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "system",
		Level:    "info",
		NodeID:   l.NodeID,
		NodeName: l.NodeName,
		Event:    "障害通知のチャットメッセージを送信しました",
	})
}

func getChatMessage(l *datastore.EventLogEnt, repair bool) (string, string) {
	subtitle := "障害"
	if repair {
		subtitle = "復帰"
	}
	return fmt.Sprintf("%s(%s)", datastore.NotifyConf.Subject, subtitle),
		fmt.Sprintf(
			`発生日時: %s
		状態: %s
		タイプ: %s
		関連ノード: %s
		イベント: %s
		`,
			formatLogTime(l.Time),
			levelName(l.Level),
			l.Type,
			l.NodeName,
			l.Event,
		)
}
