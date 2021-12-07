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
	r.Interval = datastore.NotifyConf.Interval
	r.Level = datastore.NotifyConf.Level
	r.Report = datastore.NotifyConf.Report
	r.CheckUpdate = datastore.NotifyConf.CheckUpdate
	r.NotifyRepair = datastore.NotifyConf.NotifyRepair
	r.NotifyLowScore = datastore.NotifyConf.NotifyLowScore
	r.NotifyNewInfo = datastore.NotifyConf.NotifyNewInfo
	r.URL = datastore.NotifyConf.URL
	r.HTMLMail = datastore.NotifyConf.HTMLMail
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
	datastore.NotifyConf.Interval = nc.Interval
	datastore.NotifyConf.Level = nc.Level
	datastore.NotifyConf.Report = nc.Report
	datastore.NotifyConf.CheckUpdate = nc.CheckUpdate
	datastore.NotifyConf.NotifyRepair = nc.NotifyRepair
	datastore.NotifyConf.NotifyLowScore = nc.NotifyLowScore
	datastore.NotifyConf.NotifyNewInfo = nc.NotifyNewInfo
	datastore.NotifyConf.URL = nc.URL
	datastore.NotifyConf.HTMLMail = nc.HTMLMail
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
	c.Scheme()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
