package notify

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var lastSendLine = int64(0)
var failedToken = ""

func SendLine(c *datastore.NotifyConfEnt, message string, stickerPackageId, stickerId int) error {

	if !strings.Contains(c.ChatType, "line") {
		return fmt.Errorf("line disabled")
	}

	token := c.LineToken
	if token == "" || failedToken == token {
		return fmt.Errorf("invalid line token")
	}

	if message == "" {
		return fmt.Errorf("no line message")
	}

	if time.Now().Unix()-lastSendLine < 3 {
		time.Sleep(time.Second * 3)
	}

	data := url.Values{"message": {message}}
	if stickerPackageId > 0 && stickerId > 0 {
		data.Add("stickerPackageId", fmt.Sprintf("%d", stickerPackageId))
		data.Add("stickerId", fmt.Sprintf("%d", stickerId))
	}
	r, _ := http.NewRequest("POST", "https://notify-api.line.me/api/notify", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	if resp.StatusCode == 401 {
		log.Printf("line token is expired %s", token)
		failedToken = token
	}
	return fmt.Errorf("send line code=%d", resp.StatusCode)
}

// sendNotifyLine : Lineへ通知メッセージを送信する
func sendNotifyLine(l *datastore.EventLogEnt) {
	if datastore.NotifyConf.LineToken == "" {
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
		if err := SendLine(&datastore.NotifyConf, title+"\n"+message, 0, 0); err != nil {
			log.Printf("send LINE error=%v", err)
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "system",
				Level:    "warn",
				NodeID:   l.NodeID,
				NodeName: l.NodeName,
				Event:    "復帰通知をLINEへ送信できません",
			})
			return
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "system",
			Level:    "info",
			NodeID:   l.NodeID,
			NodeName: l.NodeName,
			Event:    "復帰通知をLINEへ送信しました",
		})
		return
	}
	np := getLevelNum(l.Level)
	if np > nl {
		return
	}
	title, message := getChatMessage(l, false)
	if err := SendLine(&datastore.NotifyConf, title+"\n"+message, 0, 0); err != nil {
		log.Printf("send discord error=%v", err)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "system",
			Level:    "warn",
			NodeID:   l.NodeID,
			NodeName: l.NodeName,
			Event:    "障害通知をLINEへ送信できません",
		})
		return
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "system",
		Level:    "info",
		NodeID:   l.NodeID,
		NodeName: l.NodeName,
		Event:    "障害通知をLINEへ送信しました",
	})
}
