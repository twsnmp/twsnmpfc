package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getInfluxdb(c echo.Context) error {
	r := new(datastore.InfluxdbConfEnt)
	r.URL = datastore.InfluxdbConf.URL
	r.User = datastore.InfluxdbConf.User
	r.DB = datastore.InfluxdbConf.DB
	r.Duration = datastore.InfluxdbConf.Duration
	r.AIScore = datastore.InfluxdbConf.AIScore
	r.PollingLog = datastore.InfluxdbConf.PollingLog
	return c.JSON(http.StatusOK, r)
}

func postInfluxdb(c echo.Context) error {
	ic := new(datastore.InfluxdbConfEnt)
	if err := c.Bind(ic); err != nil {
		return echo.ErrBadRequest
	}
	datastore.InfluxdbConf.URL = ic.URL
	datastore.InfluxdbConf.User = ic.User
	if ic.Password != "" {
		datastore.MapConf.Password = ic.Password
	}
	datastore.InfluxdbConf.DB = ic.DB
	datastore.InfluxdbConf.Duration = ic.Duration
	datastore.InfluxdbConf.PollingLog = ic.PollingLog
	datastore.InfluxdbConf.AIScore = ic.AIScore
	if err := datastore.SaveInfluxdbConf(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "Influxdbの設定を更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteInfluxdb(c echo.Context) error {
	if err := datastore.InitInfluxdb(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "Influxdbのデータを削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
