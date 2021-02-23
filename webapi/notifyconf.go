package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getNotifyConf(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := new(datastore.NotifyConfEnt)
	r.MailServer = api.DataStore.NotifyConf.MailServer
	r.User = api.DataStore.NotifyConf.User
	//	r.Password = api.DataStore.NotifyConf.Password
	r.InsecureSkipVerify = api.DataStore.NotifyConf.InsecureSkipVerify
	r.MailTo = api.DataStore.NotifyConf.MailTo
	r.MailFrom = api.DataStore.NotifyConf.MailFrom
	r.Subject = api.DataStore.NotifyConf.Subject
	r.Interval = api.DataStore.NotifyConf.Interval
	r.Level = api.DataStore.NotifyConf.Level
	r.Report = api.DataStore.NotifyConf.Report
	r.CheckUpdate = api.DataStore.NotifyConf.CheckUpdate
	r.NotifyRepair = api.DataStore.NotifyConf.NotifyRepair
	return c.JSON(http.StatusOK, r)
}

func postNotifyConf(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	api.DataStore.NotifyConf.MailServer = nc.MailServer
	api.DataStore.NotifyConf.User = nc.User
	api.DataStore.NotifyConf.InsecureSkipVerify = nc.InsecureSkipVerify
	api.DataStore.NotifyConf.MailTo = nc.MailTo
	api.DataStore.NotifyConf.MailFrom = nc.MailFrom
	api.DataStore.NotifyConf.Subject = nc.Subject
	api.DataStore.NotifyConf.Interval = nc.Interval
	api.DataStore.NotifyConf.Level = nc.Level
	api.DataStore.NotifyConf.Report = nc.Report
	api.DataStore.NotifyConf.CheckUpdate = nc.CheckUpdate
	api.DataStore.NotifyConf.NotifyRepair = nc.NotifyRepair
	if nc.Password != "" {
		api.DataStore.NotifyConf.Password = nc.Password
	}
	if err := api.DataStore.SaveNotifyConfToDB(); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNotifyTest(c echo.Context) error {
	nc := new(datastore.NotifyConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	if err := api.Notify.SendTestMail(nc); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
