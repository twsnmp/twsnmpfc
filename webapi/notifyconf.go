package webapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/notify"
)

func getNotifyConf(c echo.Context) error {
	r := new(datastore.NotifyConfEnt)
	r.MailServer = datastore.NotifyConf.MailServer
	r.User = datastore.NotifyConf.User
	r.InsecureSkipVerify = datastore.NotifyConf.InsecureSkipVerify
	r.MailTo = datastore.NotifyConf.MailTo
	r.MailFrom = datastore.NotifyConf.MailFrom
	r.Subject = datastore.NotifyConf.Subject
	r.AddNodeName = datastore.NotifyConf.AddNodeName
	r.Interval = datastore.NotifyConf.Interval
	r.Level = datastore.NotifyConf.Level
	r.Report = datastore.NotifyConf.Report
	r.CheckUpdate = datastore.NotifyConf.CheckUpdate
	r.NotifyRepair = datastore.NotifyConf.NotifyRepair
	r.NotifyLowScore = datastore.NotifyConf.NotifyLowScore
	r.NotifyNewInfo = datastore.NotifyConf.NotifyNewInfo
	r.URL = datastore.NotifyConf.URL
	r.HTMLMail = datastore.NotifyConf.HTMLMail
	r.ChatType = datastore.NotifyConf.ChatType
	r.ChatWebhookURL = datastore.NotifyConf.ChatWebhookURL
	r.ExecCmd = datastore.NotifyConf.ExecCmd
	return c.JSON(http.StatusOK, r)
}

func postNotifyConf(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	datastore.NotifyConf.MailServer = nc.MailServer
	datastore.NotifyConf.User = nc.User
	datastore.NotifyConf.InsecureSkipVerify = nc.InsecureSkipVerify
	datastore.NotifyConf.MailTo = nc.MailTo
	datastore.NotifyConf.MailFrom = nc.MailFrom
	datastore.NotifyConf.Subject = nc.Subject
	datastore.NotifyConf.AddNodeName = nc.AddNodeName
	datastore.NotifyConf.Interval = nc.Interval
	datastore.NotifyConf.Level = nc.Level
	datastore.NotifyConf.Report = nc.Report
	datastore.NotifyConf.CheckUpdate = nc.CheckUpdate
	datastore.NotifyConf.NotifyRepair = nc.NotifyRepair
	datastore.NotifyConf.NotifyLowScore = nc.NotifyLowScore
	datastore.NotifyConf.NotifyNewInfo = nc.NotifyNewInfo
	datastore.NotifyConf.URL = nc.URL
	datastore.NotifyConf.HTMLMail = nc.HTMLMail
	datastore.NotifyConf.ChatType = nc.ChatType
	datastore.NotifyConf.ChatWebhookURL = nc.ChatWebhookURL
	datastore.NotifyConf.ExecCmd = nc.ExecCmd
	if nc.Password != "" {
		datastore.NotifyConf.Password = nc.Password
	}
	if nc.URL == "" {
		datastore.NotifyConf.URL = fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)
	}
	if err := datastore.SaveNotifyConf(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "通知設定を更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNotifyTest(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	if nc.URL == "" {
		nc.URL = fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)
	}
	if err := notify.SendTestMail(nc); err != nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "user",
			Level: "warn",
			Event: fmt.Sprintf("試験メールの送信に失敗しました err=%v", err),
		})
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "試験メールを送信しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNotifyChatTest(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	if nc.URL == "" {
		nc.URL = fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)
	}
	title := fmt.Sprintf("%s（試験メッセージ）", datastore.NotifyConf.Subject)
	if err := notify.SendChat(nc, title, "info", "テストです。"); err != nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "user",
			Level: "warn",
			Event: fmt.Sprintf("試験メッセージの送信に失敗しました err=%v", err),
		})
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "試験メッセージ通知を送信しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNotifyExecTest(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	if err := notify.ExecNotifyCmd(nc.ExecCmd, 1); err != nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "user",
			Level: "warn",
			Event: fmt.Sprintf("通知コマンド実行の試験に失敗しました err=%v", err),
		})
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "通知コマンド実行の試験に成功しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
