// Package webapi : WEB API
package webapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/security"
)

type WebAPI struct {
	Statik        http.Handler
	Port          string
	UseTLS        bool
	Host          string
	IP            string
	Local         bool
	Password      string
	DataStorePath string
}

type selectEntWebAPI struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

var e *echo.Echo

func Start(p *WebAPI) {
	e = echo.New()
	setup(p)
	if err := e.StartServer(makeServer(p)); err != nil {
		log.Println(err)
	}
}

func Stop() {
	ctxStopWeb, cancelWeb := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelWeb()
	if err := e.Shutdown(ctxStopWeb); err != nil {
		log.Printf("webui shutdown err=%v", err)
	}
}

func setup(p *WebAPI) {
	e.HideBanner = true
	e.HidePort = true
	e.Validator = &twsnmpfcValidator{validator: validator.New()}
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
	r.GET("/conf/grok", getGrok)
	r.GET("/export/grok", getExportGrok)
	r.POST("/conf/grok", postGrok)
	r.POST("/test/grok", postTestGrok)
	r.POST("/import/grok", postImportGrok)
	r.DELETE("/conf/grok/:id", deleteGrok)
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
	r.POST("/pollings/setlogmode", setPollingLogMode)
	r.GET("/polling/check/:id", getPollingCheck)
	// log
	r.POST("/log/eventlogs", postEventLogs)
	r.GET("/log/lastlogs/:st", postLastEventLogs)
	r.POST("/log/syslog", postSyslog)
	r.POST("/log/snmptrap", postSnmpTrap)
	r.POST("/log/netflow", postNetFlow)
	r.POST("/log/ipfix", postIPFIX)
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
	r.GET("/monitor", getMonitor)
	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", p.Statik)))
}

func getMonitor(c echo.Context) error {
	return c.JSON(http.StatusOK, backend.MonitorDataes)
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

func makeServer(p *WebAPI) *http.Server {
	sv := &http.Server{}
	if p.Local {
		sv.Addr = fmt.Sprintf("127.0.0.1:%s", p.Port)
		return sv
	}
	sv.Addr = fmt.Sprintf(":%s", p.Port)
	if !p.UseTLS {
		return sv
	}
	cert := getServerCert(p)
	sv.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
		},
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		InsecureSkipVerify:       true,
	}
	return sv
}

func getServerCert(p *WebAPI) tls.Certificate {
	//証明書、秘密鍵ファイルがある場合
	kpath := filepath.Join(p.DataStorePath, "key.pem")
	cpath := filepath.Join(p.DataStorePath, "cert.pem")
	keyPem, err := ioutil.ReadFile(kpath)
	if err == nil {
		certPem, err := ioutil.ReadFile(cpath)
		if err == nil {
			keyPem = []byte(security.GetRawKeyPem(string(keyPem), p.Password))
			cert, err := tls.X509KeyPair(certPem, keyPem)
			if err == nil {
				return cert
			}
		}
	}
	// 秘密鍵と証明書を自動作成する
	certPem, keyPem, err := security.MakeWebAPICert(p.Host, p.Password, p.IP)
	if err != nil {
		log.Fatalf("getServerCert err=%v", err)
	}
	keyPemRaw := []byte(security.GetRawKeyPem(string(keyPem), p.Password))
	cert, err := tls.X509KeyPair(certPem, keyPemRaw)
	if err != nil {
		log.Fatalf("getServerCert err=%v", err)
	}
	if err := ioutil.WriteFile(kpath, keyPem, 0600); err != nil {
		log.Printf("getServerCert err=%v", err)
	}
	if err := ioutil.WriteFile(cpath, certPem, 0600); err != nil {
		log.Printf("getServerCert err=%v", err)
	}
	return cert
}

type twsnmpfcValidator struct {
	validator *validator.Validate
}

func (v *twsnmpfcValidator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return nil
}
