package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func WebHookTest(n *datastore.NotifyConfEnt) error {
	if n.WebHookNotify != "" {
		payload := webhookNotifyPayload{}
		payload.Log = append(payload.Log, webhookNotifyLog{
			Time:     time.Now().Format(time.RFC3339),
			Type:     "test",
			NodeName: "test node name",
			NodeID:   "test node ID",
			Event:    "test event",
			Level:    "info",
		})
		payload.Count = len(payload.Log)
		j, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		err = PostWebhook(n.WebHookNotify, j)
		if err != nil {
			return err
		}
	}
	if n.WebHookReport != "" {
		payload := webhookReportPayload{
			Title: "Test Report",
		}
		payload.Info = append(payload.Info, webhookReportInfo{
			Name:  "test name",
			Value: "test value",
		})
		payload.AI = append(payload.AI, webhookReportAI{
			Score:   1.0,
			Polling: "test polling",
			Node:    "test node",
			Time:    time.Now().Format(time.RFC3339),
		})
		payload.Sensors = append(payload.Sensors, webhookSensor{
			Host:      "test",
			Type:      "syslog",
			Total:     100,
			Send:      1,
			State:     "normal",
			FirstTime: time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
			LastTime:  time.Now().Format(time.RFC3339),
		})
		j, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		err = PostWebhook(n.WebHookReport, j)
		if err != nil {
			return err
		}
	}
	return nil
}

type webhookNotifyPayload struct {
	Count int                `json:"Count"`
	Log   []webhookNotifyLog `json:"Log"`
}

type webhookNotifyLog struct {
	Time     string `json:"Time"`
	Type     string `json:"Type"`
	Level    string `json:"Level"`
	NodeName string `json:"NodeName"`
	NodeID   string `json:"NodeID"`
	Event    string `json:"Event"`
}

func webhookNotify(list []*datastore.EventLogEnt) {
	if datastore.NotifyConf.WebHookNotify == "" {
		return
	}
	nl := getLevelNum(datastore.NotifyConf.Level)
	if nl == 3 {
		return
	}
	payload := webhookNotifyPayload{}
	ti := time.Now().Add(time.Duration(-datastore.NotifyConf.Interval) * time.Minute).UnixNano()
	for _, l := range list {
		if ti > l.Time {
			continue
		}
		np := getLevelNum(l.Level)
		if np > nl {
			continue
		}
		payload.Log = append(payload.Log, webhookNotifyLog{
			Time:     time.Unix(0, l.Time).Format(time.RFC3339),
			Type:     l.Type,
			NodeName: l.NodeName,
			NodeID:   l.NodeID,
			Event:    l.Event,
			Level:    l.Level,
		})
	}
	payload.Count = len(payload.Log)
	if payload.Count < 1 {
		return
	}
	j, err := json.Marshal(payload)
	if err != nil {
		log.Printf("webhookNotify err=%v", err)
		return
	}
	err = PostWebhook(datastore.NotifyConf.WebHookNotify, j)
	if err != nil {
		log.Printf("webhookNotify err=%v", err)
	}
}

type webhookReportPayload struct {
	Title   string              `json:"Title"`
	Info    []webhookReportInfo `json:"Info"`
	Sensors []webhookSensor     `json:"Sensors"`
	AI      []webhookReportAI   `json:"AI"`
}

type webhookReportInfo struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type webhookReportAI struct {
	Score   float64 `json:"Score"`
	Node    string  `json:"Node"`
	Polling string  `json:"Polling"`
	Count   int     `json:"Count"`
	Time    string  `json:"Time"`
}

type webhookSensor struct {
	Host      string `json:"Host"`
	Type      string `json:"Type"`
	Total     int64  `json:"Total"`
	Send      int64  `json:"Send"`
	State     string `json:"State"`
	FirstTime string `json:"FirstTime"`
	LastTime  string `json:"LastTime"`
}

func webhookReport(title string, info []reportInfoEnt, sensors []*datastore.SensorEnt, ai []aiResultEnt) {
	if datastore.NotifyConf.WebHookReport == "" {
		return
	}
	payload := webhookReportPayload{
		Title: title,
	}
	for _, i := range info {
		payload.Info = append(payload.Info,
			webhookReportInfo{
				Name:  i.Name,
				Value: i.Value,
			})
	}
	for _, s := range sensors {
		payload.Sensors = append(payload.Sensors, webhookSensor{
			Host:      s.Host,
			Type:      s.Type,
			Total:     s.Total,
			Send:      s.Send,
			State:     s.State,
			FirstTime: time.Unix(0, s.FirstTime).Format(time.RFC3339),
			LastTime:  time.Unix(0, s.LastTime).Format(time.RFC3339),
		})
	}
	for _, a := range ai {
		payload.AI = append(payload.AI, webhookReportAI{
			Score:   a.LastScore,
			Node:    a.NodeName,
			Polling: a.PollingName,
			Count:   a.Count,
			Time:    time.Unix(0, a.LastTime).Format(time.RFC3339),
		})
	}
	j, err := json.Marshal(payload)
	if err != nil {
		log.Printf("webhookNotify err=%v", err)
		return
	}
	err = PostWebhook(datastore.NotifyConf.WebHookNotify, j)
	if err != nil {
		log.Printf("webhookNotify err=%v", err)
	}
}

func PostWebhook(url string, j []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Second * time.Duration(2),
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook error %s", resp.Status)
	}
	return nil
}
