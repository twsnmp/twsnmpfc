// Package webapi : WEB API
package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/discover"
	"github.com/twsnmp/twsnmpfc/logger"
	"github.com/twsnmp/twsnmpfc/notify"
	"github.com/twsnmp/twsnmpfc/ping"
	"github.com/twsnmp/twsnmpfc/polling"
	"github.com/twsnmp/twsnmpfc/report"
)

type WebAPI struct {
	DataStore *datastore.DataStore
	Backend   *backend.Backend
	Notify    *notify.Notify
	Report    *report.Report
	Ping      *ping.Ping
	Polling   *polling.Polling
	Discover  *discover.Discover
	Logger    *logger.Logger
	Statik    http.Handler
	Password  string
}

// 削除のためにIDだけ受け取る
type idWebAPI struct {
	ID string
}

type selectEntWebAPI struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

func Init(e *echo.Echo, p *WebAPI) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middle(p))
	// Route
	e.POST("/login", login)
	e.GET("/backimage", getBackImage)
	// JWT保護されたRoute
	r := e.Group("/api")
	r.Use(middleware.JWT([]byte(p.Password)))
	r.GET("/me", getMe)
	r.GET("/conf/map", getMapConf)
	r.POST("/conf/map", postMapConf)
	r.POST("/conf/backimage", postBackImage)
	r.DELETE("/conf/backimage", deleteBackImage)
	r.GET("/conf/notify", getNotifyConf)
	r.POST("/conf/notify", postNotifyConf)
	r.POST("/notify/test", postNotifyTest)
	r.GET("/conf/influxdb", getInfluxdb)
	r.POST("/conf/influxdb", postInfluxdb)
	r.DELETE("/conf/influxdb", deleteInfluxdb)
	r.GET("/conf/datastore", getDataStore)
	r.POST("/conf/backup", postBackup)
	r.DELETE("/report", deleteReport)
	r.DELETE("/logs", deleteLogs)
	r.DELETE("/ai", deleteAIResult)
	r.GET("/discover", getDiscover)
	r.POST("/discover/start", postDiscoverStart)
	r.POST("/discover/stop", postDiscoverStop)
	r.GET("/nodes", getNodes)
	r.POST("/node/delete", postNodeDelete)
	r.POST("/node/update", postNodeUpdate)
	r.GET("/node/:id", getNode)
	r.GET("/map", getMap)
	r.POST("/map/update", postMapUpdate)
	r.POST("/map/delete", postMapDelete)
	r.POST("/line/delete", postLineDelete)
	r.POST("/line/add", postLineAdd)

	r.GET("/pollings", getPollings)
	r.POST("/polling/:id", postPolling)
	r.POST("/polling/add", postPollingAdd)
	r.POST("/polling/update", postPollingUpdate)
	r.POST("/polling/delete", postPollingDelete)
	r.POST("/eventlogs", postEventLogs)
	r.POST("/syslog", postSyslog)
	r.POST("/snmptrap", postSnmpTrap)
	r.POST("/netflow", postNetFlow)
	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", p.Statik)))
}

func middle(p *WebAPI) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("api", p)
			if err := next(c); err != nil {
				return err
			}
			return nil
		}
	}
}
