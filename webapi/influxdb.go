package webapi

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getInfluxdb(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := new(datastore.InfluxdbConfEnt)
	r.URL = api.DataStore.InfluxdbConf.URL
	r.User = api.DataStore.InfluxdbConf.User
	r.DB = api.DataStore.InfluxdbConf.DB
	r.Duration = api.DataStore.InfluxdbConf.Duration
	r.AIScore = api.DataStore.InfluxdbConf.AIScore
	r.PollingLog = api.DataStore.InfluxdbConf.PollingLog
	return c.JSON(http.StatusOK, r)
}

func postInfluxdb(c echo.Context) error {
	ic := new(datastore.InfluxdbConfEnt)
	if err := c.Bind(ic); err != nil {
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	api.DataStore.InfluxdbConf.URL = ic.URL
	api.DataStore.InfluxdbConf.User = ic.User
	if ic.Password != "" {
		api.DataStore.MapConf.Password = ic.Password
	}
	api.DataStore.InfluxdbConf.DB = ic.DB
	api.DataStore.InfluxdbConf.Duration = ic.Duration
	api.DataStore.InfluxdbConf.PollingLog = ic.PollingLog
	api.DataStore.InfluxdbConf.AIScore = ic.AIScore
	if err := api.DataStore.SaveInfluxdbConfToDB(); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteInfluxdb(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	if err := api.DataStore.InitInfluxdb(); err != nil {
		log.Printf("deleteInfluxdb err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
