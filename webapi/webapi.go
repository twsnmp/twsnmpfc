// Package webapi : WEB API
package webapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/security"
)

type WebAPI struct {
	Statik        http.Handler
	Port          string
	UseTLS        bool
	Host          string
	IP            string
	Local         bool
	EnableMCP     bool
	MCPFrom       string
	Timeout       int
	Password      string
	DataStorePath string
	Version       string
	QuitSignal    chan os.Signal
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
		log.Printf("start webapi err=%v", err)
	}
}

func Stop() {
	ctxStopWeb, cancelWeb := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelWeb()
	stopMCPServer(ctxStopWeb)
	if err := e.Shutdown(ctxStopWeb); err != nil {
		log.Printf("shutdown webapi err=%v", err)
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
	mime.AddExtensionType(".js", "application/javascript")
	e.Use(logger)
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middle(p))
	// Route
	e.POST("/login", login)
	e.GET("/backimage", getBackImage)
	e.GET("/image/:path", getImage)
	e.GET("/version", getVersion)
	e.GET("/imageIcon/:id", getImageIcon)
	if p.EnableMCP {
		startMCPServer(e, p.MCPFrom)
	}
	// JWT保護されたRoute
	r := e.Group("/api")
	r.Use(echojwt.JWT([]byte(p.Password)))
	r.POST("/feedback", postFeedback)
	r.POST("/stop", postStop)
	r.GET("/checkupdate", getCheckUpdate)
	r.GET("/me", getMe)
	r.GET("/conf/map", getMapConf)
	r.POST("/conf/map", postMapConf)
	r.GET("/conf/sshPublicKey", getSSHPublicKey)
	r.POST("/conf/sshPublicKey", postSSHPublicKey)
	r.POST("/conf/sshkey", postReGenarateSSHKey)
	r.GET("/conf/icons", getIcons)
	r.POST("/conf/icon", postIcon)
	r.DELETE("/conf/icon/:icon", deleteIcon)
	r.POST("/conf/backimage", postBackImage)
	r.DELETE("/conf/backimage", deleteBackImage)
	r.POST("/image", postImage)
	r.DELETE("/image/:path", deleteImage)
	r.POST("/conf/geoip", postGeoIP)
	r.DELETE("/conf/geoip", deleteGeoIP)
	r.GET("/conf/notify", getNotifyConf)
	r.POST("/conf/notify", postNotifyConf)
	r.GET("/conf/notifySchedule", getNotifySchedule)
	r.POST("/conf/notifySchedule", postNotifySchedule)
	r.DELETE("/conf/notifySchedule/:id", deleteNotifySchedule)
	r.POST("/notify/test", postNotifyTest)
	r.POST("/notify/chat/test", postNotifyChatTest)
	r.POST("/notify/exec/test", postNotifyExecTest)
	r.GET("/conf/influxdb", getInfluxdb)
	r.POST("/conf/influxdb", postInfluxdb)
	r.DELETE("/conf/influxdb", deleteInfluxdb)
	r.GET("/conf/grok", getGrok)
	r.GET("/export/grok", getExportGrok)
	r.POST("/conf/grok", postGrok)
	r.POST("/conf/defgrok", postDefGrok)
	r.POST("/test/grok", postTestGrok)
	r.POST("/import/grok", postImportGrok)
	r.DELETE("/conf/grok/:id", deleteGrok)
	r.GET("/conf/datastore", getDataStore)
	r.POST("/conf/backup", postBackup)
	r.POST("/stop/backup", postStopBackup)
	r.GET("/conf/report", getReportConf)
	r.POST("/conf/report", postReportConf)
	r.DELETE("/logs", deleteLogs)
	r.DELETE("/arp", deleteArp)
	r.GET("/discover", getDiscover)
	r.GET("/discover/range", getDiscoverIPRange)
	r.DELETE("/discover/stat", deleteDiscoverStat)
	r.POST("/discover/start", postDiscoverStart)
	r.POST("/discover/stop", postDiscoverStop)
	r.GET("/nodes", getNodes)
	r.POST("/nodes/delete", deleteNodes)
	r.POST("/node/update", postNodeUpdate)
	r.POST("/nodes/delete_items", deleteDrawItems)
	r.POST("/item/update", postItemUpdate)
	r.GET("/node/log/:id", getNodeLog)
	r.GET("/node/polling/:id", getNodePolling)
	r.GET("/node/vpanel/:id", getVPanel)
	r.GET("/node/hostResource/:id", getHostResource)
	r.GET("/node/rmon/:id/:type", getRMON)
	r.GET("/node/port/:id", getPortList)
	r.GET("/node/memo/:id", getNodeMemo)
	r.POST("/node/memo", postNodeMemo)

	r.POST("/mibbr", postMIBBr)
	r.GET("/mibbr/:id", getMIBBr)
	r.POST("/gnmi", postGNMI)
	r.GET("/gnmi/:id", getGNMI)
	r.GET("/map", getMap)
	r.POST("/map/update", postNodePos)
	r.POST("/map/update_item", postItemPos)
	r.POST("/line/delete", deleteLine)
	r.DELETE("/line/:id", deleteLineByID)
	r.POST("/line/add", postLine)
	r.DELETE("/network/:id", deleteNetwork)
	r.GET("/findNeighborNetworksAndLines/:id", getFindNeighborNetworksAndLines)
	r.GET("/checkNetwork/:id", getCheckNetwork)
	r.POST("/network/update", postNetwork)
	r.POST("/map/update_network", postNetworkPos)
	r.POST("/wol/:id", postWOL)
	// Ping画面
	r.POST("/ping", postPing)

	r.GET("/pollings", getPollings)
	r.GET("/polling/template", getPollingTemplate)
	r.POST("/polling/add", postPollingAdd)
	r.POST("/polling/auto", postPollingAutoAdd)
	r.GET("/polling/:id", getPolling)
	r.POST("/pollingLogs/:id", postPollingLogs)
	r.GET("/aidata/:id", getPollingAIData)
	r.POST("/polling/update", postPollingUpdate)
	r.POST("/pollings/delete", deletePollings)
	r.POST("/pollings/setlevel", setPollingLevel)
	r.POST("/pollings/setlogmode", setPollingLogMode)
	r.POST("/pollings/setParams", setPollingParams)
	r.GET("/polling/check/:id", getPollingCheck)
	r.GET("/polling/TimeAnalyze/:id", getPollingLogTimeAnalyze)
	r.DELETE("/polling/clear/:id", deletePollingLog)
	// log
	r.POST("/log/eventlogs", postEventLogs)
	r.GET("/log/lastlogs/:st", postLastEventLogs)
	r.POST("/log/syslog", postSyslog)
	r.POST("/log/snmptrap", postSnmpTrap)
	r.POST("/log/netflow", postNetFlow)
	r.POST("/log/sflow", postSFlow)
	r.POST("/log/sflowCounter", postSFlowCounter)
	r.POST("/log/ipfix", postIPFIX)
	r.POST("/log/arp", postArp)
	r.DELETE("/log/:id", deleteLog)
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
	r.GET("/report/ipam", getIPAM)
	r.GET("/report/ips", getIPReport)
	r.DELETE("/report/ip/:ip", deleteIPReport)
	r.POST("/report/ips/reset", resetIPReport)
	r.GET("/report/address/:addr", getAddressInfo)
	r.GET("/report/unknownport", getUnknownPortList)
	r.GET("/report/ether", getEtherType)
	r.DELETE("/report/ether", deleteEtherType)
	r.GET("/report/dnsq", getDNSQ)
	r.DELETE("/report/dnsq", deleteDNSQ)
	r.GET("/report/radius", getRADIUSFlows)
	r.DELETE("/report/radius/:id", deleteRADIUSFlow)
	r.POST("/report/radius/reset", resetRADIUSFlows)
	r.GET("/report/tls", getTLSFlows)
	r.DELETE("/report/tls/:id", deleteTLSFlow)
	r.POST("/report/tls/reset", resetTLSFlows)
	r.GET("/report/cert", getCert)
	r.POST("/report/cert", postCert)
	r.DELETE("/report/cert/:id", deleteCert)
	r.POST("/report/cert/reset", resetCert)
	r.GET("/report/sensors", getSensors)
	r.GET("/report/sensor/stats/:id", getSensorStats)
	r.GET("/report/sensor/monitors/:id", getSensorMonitors)
	r.POST("/report/sensor/:id", postSensor)
	r.DELETE("/report/sensor/:id", deleteSensor)
	r.GET("/report/WinEventIDs", getWinEventID)
	r.DELETE("/report/WinEventID/:id", deleteWinEventID)
	r.GET("/report/WinLogon", getWinLogon)
	r.POST("/report/WinLogon/reset", resetWinLogon)
	r.DELETE("/report/WinLogon/:id", deleteWinLogon)
	r.GET("/report/WinAccount", getWinAccount)
	r.DELETE("/report/WinAccount/:id", deleteWinAccount)
	r.GET("/report/WinKerberos", getWinKerberos)
	r.POST("/report/WinKerberos/reset", resetWinKerberos)
	r.DELETE("/report/WinKerberos/:id", deleteWinKerberos)
	r.GET("/report/WinPrivilege", getWinPrivilege)
	r.DELETE("/report/WinPrivilege/:id", deleteWinPrivilege)
	r.GET("/report/WinProcess", getWinProcess)
	r.DELETE("/report/WinProcess/:id", deleteWinProcess)
	r.GET("/report/WinTask", getWinTask)
	r.DELETE("/report/WinTask/:id", deleteWinTask)
	r.GET("/report/WifiAP", getWifiAP)
	r.DELETE("/report/WifiAP/:id", deleteWifiAP)
	r.POST("/report/BlueScan/name", postBlueScanName)
	r.GET("/report/BlueDevice", getBlueDevice)
	r.DELETE("/report/BlueDevice/:id", deleteBlueDevice)
	r.GET("/report/EnvMonitor", getEnvMonitor)
	r.DELETE("/report/EnvMonitor/:id", deleteEnvMonitor)
	r.GET("/report/PowerMonitor", getPowerMonitor)
	r.DELETE("/report/PowerMonitor/:id", deletePowerMonitor)
	r.GET("/report/sdrPowerKeys", getSdrPowerKeys)
	r.POST("/report/sdrPowerData", getSdrPowerData)
	r.POST("/report/sdrPower/delete", deleteSdrPower)
	r.GET("/report/MotionSensor", getMotionSensor)
	r.DELETE("/report/MotionSensor/:id", deleteMotionSensor)
	// AI
	r.GET("/report/ailist", getAIList)
	r.GET("/report/ai/:id", getAIResult)
	r.DELETE("/report/ai/:id", deleteAIResult)
	r.GET("/monitor", getMonitor)
	r.GET("/mibmods", getMibMods)
	r.GET("/mibtree", getMibTree)
	r.GET("/imageIconList", getImageIconList)
	// PKI
	r.GET("/pki/hasCA", getHasCA)
	r.GET("/pki/certs", getPKICerts)
	r.GET("/pki/createCA", getDefaultCreateCAReq)
	r.POST("/pki/createCA", postCreateCA)
	r.POST("/pki/destroyCA", postDestroyCA)
	r.POST("/pki/createCSR", postCreateCertificateRequest)
	r.POST("/pki/createCRT", postCreateCertificate)
	r.DELETE("/pki/revoke/:id", deleteRevokeCert)
	r.GET("/pki/cert/:id", getExportCert)
	r.GET("/pki/control", getPKIControl)
	r.POST("/pki/control", postPKIControl)
	r.GET("/otel/metrics", getOTelMetrics)
	r.GET("/otel/metric/:id", getOTelMetric)
	r.GET("/otel/traceBucketList", getOTelTraceBucketList)
	r.POST("/otel/traces", postOTelTraces)
	r.POST("/otel/trace", postOTelTrace)
	r.POST("/otel/dag", postOTelDAG)
	r.GET("/otel/logs", getOTelLastLog)
	r.DELETE("/otel/alldata", deleteOTelAllData)

	// Mobile API
	m := e.Group("/mobile")
	m.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if datastore.MapConf.EnableMobileAPI && username == datastore.MapConf.UserID &&
			security.PasswordVerify(datastore.MapConf.Password, password) {
			log.Printf("auth ok user=%s", username)
			return true, nil
		}
		log.Printf("auth failed user=%s password=%s", username, password)
		return false, nil
	}))
	m.GET("/api/mapstatus", getMobileMapStatus)
	m.GET("/api/mapdata", getMobileMapData)
	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", p.Statik)))
}

func getMonitor(c echo.Context) error {
	return c.JSON(http.StatusOK, datastore.MonitorDataes)
}

func getMibMods(c echo.Context) error {
	return c.JSON(http.StatusOK, datastore.MIBModules)
}

func getMibTree(c echo.Context) error {
	return c.JSON(http.StatusOK, datastore.MIBTree)
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
		log.Println("not tls")
		return sv
	}
	if cert, err := getServerCert(p); err == nil {
		sv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
			},
			MinVersion: tls.VersionTLS13,
		}
	} else {
		log.Printf("getServerCert err=%v", err)
	}
	return sv
}

func getServerCert(p *WebAPI) (tls.Certificate, error) {
	//証明書、秘密鍵ファイルがある場合
	kpath := filepath.Join(p.DataStorePath, "key.pem")
	cpath := filepath.Join(p.DataStorePath, "cert.pem")
	keyPem, err := os.ReadFile(kpath)
	if err == nil {
		certPem, err := os.ReadFile(cpath)
		if err == nil {
			if strings.Contains(string(keyPem), "RSA PRIVATE") {
				keyPem = []byte(security.GetRawKeyPem(string(keyPem), p.Password))
			}
			cert, err := tls.X509KeyPair(certPem, keyPem)
			if err == nil {
				log.Println("use old cert")
				return cert, nil
			}
		}
	}
	// 秘密鍵と証明書を自動作成する
	certPem, keyPem, err := security.MakeWebAPICert(p.Host, p.IP)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPemRaw := []byte(security.GetRawKeyPem(string(keyPem), ""))
	cert, err := tls.X509KeyPair(certPem, keyPemRaw)
	if err != nil {
		return cert, err
	}
	if err := os.WriteFile(kpath, keyPem, 0600); err != nil {
		return cert, err
	}
	if err := os.WriteFile(cpath, certPem, 0600); err != nil {
		return cert, err
	}
	log.Println("new cert")
	return cert, nil
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

func postStop(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	go func() {
		time.Sleep(5 * time.Second)
		api.QuitSignal <- os.Interrupt
	}()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
