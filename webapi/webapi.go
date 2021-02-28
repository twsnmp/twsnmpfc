// Package webapi : WEB API
package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type WebAPI struct {
	Statik   http.Handler
	Password string
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
	r.DELETE("/arp", deleteArp)
	r.GET("/discover", getDiscover)
	r.POST("/discover/start", postDiscoverStart)
	r.POST("/discover/stop", postDiscoverStop)
	r.GET("/nodes", getNodes)
	r.POST("/node/delete", postNodeDelete)
	r.POST("/node/update", postNodeUpdate)
	r.GET("/node/:id", getNode)
	r.POST("/mibbr", postMIBBr)
	r.GET("/mibbr/:id", getMIBBr)
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
	// log
	r.POST("/log/eventlogs", postEventLogs)
	r.POST("/log/syslog", postSyslog)
	r.POST("/log/snmptrap", postSnmpTrap)
	r.POST("/log/netflow", postNetFlow)
	r.POST("/log/arp", postArp)
	// report
	r.GET("/report/devices", getDevices)
	r.DELETE("/report/device/:id", deleteDevice)
	r.POST("/report/devices/reset", resetDevices)
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
