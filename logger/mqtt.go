package logger

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

var MqttTCPPort = 1883
var MqttWSPort = 1884
var MqttCert = ""
var MqttKey = ""
var MqttUsers = ""
var MqttFrom = ""

func mqttd(stopCh chan bool) {
	log.Printf("start mqttd")
	s := startMqttServer()
	<-stopCh
	s.Close()
	log.Printf("stop mqttd")
}

func startMqttServer() *mqtt.Server {
	// Create the new MQTT Server.
	server := mqtt.New(nil)
	err := server.AddHook(new(mqttHook), nil)
	if err != nil {
		log.Fatal(err)
	}
	if MqttFrom == "" && MqttUsers == "" {
		// Allow all connections.
		if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
			log.Fatal(err)
		}
	} else {
		authRules := &auth.Ledger{}
		for _, e := range strings.Split(MqttUsers, ",") {
			a := strings.SplitN(e, ":", 2)
			if len(a) == 2 {
				authRules.Auth = append(authRules.Auth, auth.AuthRule{
					Username: auth.RString(a[0]),
					Password: auth.RString(a[1]),
					Allow:    true,
				})
			}
		}
		for _, e := range strings.Split(MqttFrom, ",") {
			authRules.Auth = append(authRules.Auth, auth.AuthRule{
				Remote: auth.RString(e),
				Allow:  true,
			})
		}
		if err := server.AddHook(new(auth.Hook), &auth.Options{
			Ledger: authRules,
		}); err != nil {
			log.Fatal(err)
		}
	}
	var tlsConfig *tls.Config
	if MqttCert != "" && MqttKey != "" {
		cert, err := tls.LoadX509KeyPair(MqttCert, MqttKey)
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	if MqttTCPPort > 0 {
		tcp := listeners.NewTCP(listeners.Config{
			ID:        "tcp1",
			Address:   fmt.Sprintf(":%d", MqttTCPPort),
			TLSConfig: tlsConfig,
		})
		if err := server.AddListener(tcp); err != nil {
			log.Fatal(err)
		}
	}
	if MqttWSPort > 0 {
		ws := listeners.NewWebsocket(listeners.Config{
			ID:        "ws1",
			Address:   fmt.Sprintf(":%d", MqttWSPort),
			TLSConfig: tlsConfig,
		})
		if err := server.AddListener(ws); err != nil {
			log.Fatal(err)
		}
	}
	go func() {
		err := server.Serve()
		if err != nil {
			log.Println(err)
		}
	}()
	return server
}

type mqttHook struct {
	mqtt.HookBase
}

func (h *mqttHook) ID() string {
	return "twsnmpfc-mqttd"
}

func (h *mqttHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnect,
		mqtt.OnDisconnect,
		mqtt.OnSubscribed,
		mqtt.OnUnsubscribed,
		mqtt.OnPublished,
	}, []byte{b})
}

func (h *mqttHook) Init(config any) error {
	log.Println("mqtt hook initialised")
	return nil
}

func (h *mqttHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	log.Printf("mqtt client connected client=%s", cl.ID)
	if datastore.MapConf.MqttToSyslog && datastore.MapConf.EnableSyslogd {
		logMap := make(map[string]any)
		logMap["content"] = fmt.Sprintf("mqtt clinet connected client=%s remote=%s", cl.ID, cl.Net.Remote)
		logMap["tag"] = "mqtt:connect"
		logMap["severity"] = 6
		logMap["facility"] = float64(17)
		logMap["hostname"] = cl.ID
		if j, err := json.Marshal(&logMap); err == nil {
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "syslog",
				Log:  string(j),
			}
		}
	}
	return nil
}

func (h *mqttHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	log.Printf("mqtt client disconnected client=%s,expire=%v,err=%v", cl.ID, expire, err)
	if datastore.MapConf.MqttToSyslog && datastore.MapConf.EnableSyslogd {
		logMap := make(map[string]any)
		logMap["content"] = fmt.Sprintf("mqtt clinet disconnected client=%s remote=%s exire=%v err=%v",
			cl.ID, cl.Net.Remote, expire, err)
		logMap["tag"] = "mqtt:disconnect"
		logMap["severity"] = 6
		if err != nil {
			logMap["severity"] = 4
		}
		logMap["facility"] = float64(17)
		logMap["hostname"] = cl.ID
		if j, err := json.Marshal(&logMap); err == nil {
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "syslog",
				Log:  string(j),
			}
		}
	}
}

func (h *mqttHook) OnSubscribed(cl *mqtt.Client, pk packets.Packet, reasonCodes []byte) {
	log.Printf("mqtt subscribed client=%s qos=%v", cl.ID, reasonCodes)
}

func (h *mqttHook) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet) {
	log.Printf("mqtt unsubscribed client=%s", cl.ID)
}

func (h *mqttHook) OnPublished(cl *mqtt.Client, pk packets.Packet) {
	report.UpdateSensor(cl.ID, "mqtt", 1)
	if datastore.MapConf.MqttToSyslog && datastore.MapConf.EnableSyslogd {
		logMap := make(map[string]any)
		logMap["content"] = string(pk.Payload)
		logMap["tag"] = fmt.Sprintf("mqtt:%s", pk.TopicName)
		logMap["severity"] = 6
		logMap["facility"] = float64(17)
		logMap["hostname"] = cl.ID
		if j, err := json.Marshal(&logMap); err == nil {
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "syslog",
				Log:  string(j),
			}
		}
	}
}
