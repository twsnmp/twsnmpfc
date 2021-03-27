// Package webapi : WEB API
package webapi

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type WebAPI struct {
	Statik        http.Handler
	Password      string
	DataStorePath string
}

type selectEntWebAPI struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

func Init(e *echo.Echo, p *WebAPI) {
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	logger := middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} ${method} ${status} ${uri} ${remote_ip} ${bytes_in} ${bytes_out} ${latency_human}\n",
		Output:           os.Stdout,
		CustomTimeFormat: "2006-01-02T15:04:05.000",
	})
	e.Use(logger)
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
	r.POST("/conf/geoip", postGeoIP)
	r.DELETE("/conf/geoip", deleteGeoIP)
	r.GET("/conf/notify", getNotifyConf)
	r.POST("/conf/notify", postNotifyConf)
	r.POST("/notify/test", postNotifyTest)
	r.GET("/conf/influxdb", getInfluxdb)
	r.POST("/conf/influxdb", postInfluxdb)
	r.DELETE("/conf/influxdb", deleteInfluxdb)
	r.GET("/conf/datastore", getDataStore)
	r.POST("/conf/backup", postBackup)
	r.GET("/conf/report", getReportConf)
	r.POST("/conf/report", postReportConf)
	r.DELETE("/report", deleteReport)
	r.DELETE("/logs", deleteLogs)
	r.DELETE("/arp", deleteArp)
	r.GET("/discover", getDiscover)
	r.POST("/discover/start", postDiscoverStart)
	r.POST("/discover/stop", postDiscoverStop)
	r.GET("/nodes", getNodes)
	r.POST("/nodes/delete", deleteNodes)
	r.POST("/node/update", postNodeUpdate)
	r.GET("/node/log/:id", getNodeLog)
	r.GET("/node/polling/:id", getNodePolling)
	r.POST("/mibbr", postMIBBr)
	r.GET("/mibbr/:id", getMIBBr)
	r.GET("/map", getMap)
	r.POST("/map/update", postMapUpdate)
	r.POST("/line/delete", deleteLine)
	r.POST("/line/add", postLineAdd)

	r.GET("/pollings", getPollings)
	r.GET("/polling/template", getPollingTemplate)
	r.POST("/polling/add", postPollingAdd)
	r.POST("/polling/auto", postPollingAutoAdd)
	r.POST("/polling/:id", postPolling)
	r.POST("/polling/update", postPollingUpdate)
	r.POST("/pollings/delete", deletePollings)
	r.POST("/pollings/setlevel", setPollingLevel)
	r.GET("/polling/check/:id", getPollingCheck)
	// log
	r.POST("/log/eventlogs", postEventLogs)
	r.GET("/log/lastlogs/:st", postLastEventLogs)
	r.POST("/log/syslog", postSyslog)
	r.POST("/log/snmptrap", postSnmpTrap)
	r.POST("/log/netflow", postNetFlow)
	r.POST("/log/arp", postArp)
	// report
	r.GET("/report/devices", getDevices)
	r.DELETE("/report/device/:id", deleteDevice)
	r.POST("/report/devices/reset", resetDevices)
	r.GET("/report/users", getUsers)
	r.DELETE("/report/user/:id", deleteUser)
	r.POST("/report/users/reset", resetUsers)
	r.GET("/report/servers", getServers)
	r.DELETE("/report/server/:id", deleteServer)
	r.POST("/report/servers/reset", resetServers)
	r.GET("/report/flows", getFlows)
	r.DELETE("/report/flow/:id", deleteFlow)
	r.POST("/report/flows/reset", resetFlows)
	r.GET("/report/address/:addr", getAddressInfo)
	// AI
	r.GET("/report/ailist", getAIList)
	r.GET("/report/ai/:id", getAIResult)
	r.DELETE("/report/ai/:id", deleteAIResult)
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
