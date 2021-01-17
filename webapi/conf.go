package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/security"
)

type mapConfWebAPI struct {
	MapName        string
	UserID         string
	Password       string
	PollInt        int
	Timeout        int
	Retry          int
	LogDays        int
	LogDispSize    int
	SnmpMode       string
	Community      string
	SnmpUser       string
	SnmpPassword   string
	EnableSyslogd  bool
	EnableTrapd    bool
	EnableNetflowd bool
	AILevel        string
	AIThreshold    int
}

func getMapConf(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := new(mapConfWebAPI)
	r.MapName = api.DataStore.MapConf.MapName
	r.UserID = api.DataStore.MapConf.UserID
	//	r.Password = api.DataStore.MapConf.Password
	r.PollInt = api.DataStore.MapConf.PollInt
	r.Timeout = api.DataStore.MapConf.Timeout
	r.Retry = api.DataStore.MapConf.Retry
	r.LogDays = api.DataStore.MapConf.LogDays
	r.LogDispSize = api.DataStore.MapConf.LogDispSize
	r.SnmpMode = api.DataStore.MapConf.SnmpMode
	r.Community = api.DataStore.MapConf.Community
	r.SnmpUser = api.DataStore.MapConf.SnmpUser
	//	r.SnmpPassword = api.DataStore.MapConf.SmmpPassword
	r.EnableSyslogd = api.DataStore.MapConf.EnableSyslogd
	r.EnableTrapd = api.DataStore.MapConf.EnableTrapd
	r.EnableNetflowd = api.DataStore.MapConf.EnableNetflowd
	r.AILevel = api.DataStore.MapConf.AILevel
	r.AIThreshold = api.DataStore.MapConf.AIThreshold
	return c.JSON(http.StatusOK, r)
}

func postMapConf(c echo.Context) error {
	mc := new(mapConfWebAPI)
	if err := c.Bind(mc); err != nil {
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	api.DataStore.MapConf.MapName = mc.MapName
	api.DataStore.MapConf.UserID = mc.UserID
	if mc.Password != "" {
		api.DataStore.MapConf.Password = security.PasswordHash(mc.Password)
	}
	api.DataStore.MapConf.PollInt = mc.PollInt
	api.DataStore.MapConf.Timeout = mc.Timeout
	api.DataStore.MapConf.Retry = mc.Retry
	api.DataStore.MapConf.LogDays = mc.LogDays
	api.DataStore.MapConf.LogDispSize = mc.LogDispSize
	api.DataStore.MapConf.SnmpMode = mc.SnmpMode
	api.DataStore.MapConf.Community = mc.Community
	api.DataStore.MapConf.SnmpUser = mc.SnmpUser
	if mc.SnmpPassword != "" {
		api.DataStore.MapConf.SnmpPassword = mc.SnmpPassword
	}
	api.DataStore.MapConf.EnableSyslogd = mc.EnableSyslogd
	api.DataStore.MapConf.EnableTrapd = mc.EnableTrapd
	api.DataStore.MapConf.EnableNetflowd = mc.EnableNetflowd
	api.DataStore.MapConf.AILevel = mc.AILevel
	api.DataStore.MapConf.AIThreshold = mc.AIThreshold
	api.DataStore.SaveMapConfToDB()
	return c.JSON(http.StatusOK, mc)
}