package webapi

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/notify"
)

func getNotifyConf(c echo.Context) error {
	r := new(datastore.NotifyConfEnt)
	r.MailServer = datastore.NotifyConf.MailServer
	r.User = datastore.NotifyConf.User
	r.InsecureSkipVerify = datastore.NotifyConf.InsecureSkipVerify
	r.InsecureCipherSuites = datastore.NotifyConf.InsecureCipherSuites
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
	r.WebHookNotify = datastore.NotifyConf.WebHookNotify
	r.WebHookReport = datastore.NotifyConf.WebHookReport
	r.Provider = datastore.NotifyConf.Provider
	r.ClientID = datastore.NotifyConf.ClientID
	r.ClientSecret = datastore.NotifyConf.ClientSecret
	r.MSTenant = datastore.NotifyConf.MSTenant
	return c.JSON(http.StatusOK, r)
}

func postNotifyConf(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	delOAuth2Token := false
	if nc.Provider != datastore.NotifyConf.Provider ||
		nc.ClientID != datastore.NotifyConf.ClientID ||
		nc.ClientSecret != datastore.NotifyConf.ClientSecret ||
		nc.MSTenant != datastore.NotifyConf.MSTenant {
		delOAuth2Token = true
	}
	datastore.NotifyConf.MailServer = nc.MailServer
	datastore.NotifyConf.User = nc.User
	datastore.NotifyConf.InsecureSkipVerify = nc.InsecureSkipVerify
	datastore.NotifyConf.InsecureCipherSuites = nc.InsecureCipherSuites
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
	datastore.NotifyConf.WebHookNotify = nc.WebHookNotify
	datastore.NotifyConf.WebHookReport = nc.WebHookReport
	datastore.NotifyConf.Provider = nc.Provider
	datastore.NotifyConf.ClientID = nc.ClientID
	datastore.NotifyConf.ClientSecret = nc.ClientSecret
	datastore.NotifyConf.MSTenant = nc.MSTenant
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
	if delOAuth2Token {
		datastore.DeleteNotifyOAuth2Token()
	}
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
	title := fmt.Sprintf("%s（試験メッセージ）", nc.Subject)
	if err := notify.SendChat(nc, title, "info", "テストです。"); err != nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "user",
			Level: "warn",
			Event: fmt.Sprintf("チャットへの試験メッセージの送信に失敗しました err=%v", err),
		})
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "チャットへ試験メッセージを送信しました",
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

func postNotifyWebhookTest(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	if err := notify.WebHookTest(nc); err != nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "user",
			Level: "warn",
			Event: fmt.Sprintf("Webhookの試験に失敗しました err=%v", err),
		})
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "Webhookの試験に成功しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

// 通知スケジュール
func getNotifySchedule(c echo.Context) error {
	return c.JSON(http.StatusOK, datastore.NotifySchedule)
}

var notifySchedulePat = regexp.MustCompile(`(\S+)\s+(\d{2}):(\d{2})-(\d{2}):(\d{2})`)

func postNotifySchedule(c echo.Context) error {
	type notifyScheduleEnt struct {
		NodeID   string
		Schedule string
	}
	ns := new(notifyScheduleEnt)
	if err := c.Bind(ns); err != nil || ns.Schedule == "" {
		return echo.ErrBadRequest
	}
	for _, sc := range strings.Split(ns.Schedule, ",") {
		if !notifySchedulePat.MatchString(sc) {
			return echo.ErrBadRequest
		}
	}
	datastore.NotifySchedule[ns.NodeID] = ns.Schedule
	if err := datastore.SaveNotifySchedule(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "通知除外スケジュールを更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteNotifySchedule(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		id = ""
	}
	delete(datastore.NotifySchedule, id)
	if err := datastore.SaveNotifySchedule(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "通知除外スケジュールを削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNotifyHasValidOAuth2Token(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	if datastore.HasValidNotifyOAuth2Token(nc) {
		return c.JSON(http.StatusOK, map[string]string{"valid": "true"})
	}
	return c.JSON(http.StatusOK, map[string]string{"valid": "false"})
}

func getNotifyOAuth2TokenInfo(c echo.Context) error {
	token := datastore.GetNotifyOAuth2Token()
	if token == nil {
		return c.String(http.StatusOK, "トークンなし")
	}
	return c.String(http.StatusOK, fmt.Sprintf("期限:%s リフレッシュ:%v",
		token.Expiry.Format(time.DateTime), token.RefreshToken != ""))
}

func getNotifyGetOAuth2Token(c echo.Context) error {
	url, err := notify.GetNotifyOAuth2TokenStep1()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"url": url})
}

func getNotifyOAuth2Callback(c echo.Context) error {
	state := c.FormValue("state")
	code := c.FormValue("code")
	err := notify.GetNotifyOAuth2TokenStep2(state, code)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "OAuth2の認証が完了しました。このWindowを閉じてください。")
}
