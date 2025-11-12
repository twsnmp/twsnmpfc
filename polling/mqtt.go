package polling

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/PuerkitoBio/goquery"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func doPollingMqtt(pe *datastore.PollingEnt) bool {
	switch pe.Mode {
	case "connect":
		return doPollingMqttConnect(pe)
	default:
		doPollingMqttSubscribe(pe)
		return false
	}
}

func doPollingMqttConnect(pe *datastore.PollingEnt) bool {
	client, err := getMqttClient(pe)
	if err != nil {
		setPollingError("mqtt", pe, err)
		return false
	}
	st := time.Now().UnixNano()
	token := client.Connect()
	token.Wait()
	pe.Result["rtt"] = float64(time.Now().UnixNano() - st)
	err = token.Error()
	if err != nil {
		setPollingState(pe, pe.Level)
		return true
	}
	if pe.Script != "" {
		vm := otto.New()
		setVMFuncAndValues(pe, vm)
		vm.Set("rtt", pe.Result["rtt"])
		value, err := vm.Run(pe.Script)
		if err != nil {
			setPollingError("mqtt", pe, err)
			return false
		}
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(pe, pe.Level)
			return true
		}
	}
	setPollingState(pe, "normal")
	client.Disconnect(10)
	return true
}

var mqttSubscribeMap = sync.Map{}

func doPollingMqttSubscribe(pe *datastore.PollingEnt) {
	if _, ok := mqttSubscribeMap.Load(pe.ID); ok {
		return
	}
	client, err := getMqttClient(pe)
	if err != nil {
		setPollingError("mqtt", pe, err)
		return
	}
	token := client.Connect()
	token.Wait()
	err = token.Error()
	if err != nil {
		setPollingError("mqtt", pe, err)
		return
	}
	token = client.Subscribe(pe.Filter, 1, func(cl mqtt.Client, msg mqtt.Message) {
		p := datastore.GetPolling(pe.ID)
		if p == nil || p.Params != pe.Params || p.Filter != pe.Filter ||
			p.Extractor != pe.Extractor || p.Script != pe.Script {
			MqttStopSubscription(pe.ID)
			log.Printf("p=%+v  pe=%+v", p, pe)
			return
		}
		oldState := pe.State
		ok, err := checkMqttMessage(pe, msg.Topic(), string(msg.Payload()))
		if err != nil {
			setPollingError("mqtt", pe, err)
			return
		}
		if ok {
			delete(pe.Result, "error")
			setPollingState(pe, "normal")
		} else {
			setPollingState(pe, pe.Level)
		}
		subscribeUpdatePolling(pe, oldState)
	})
	token.Wait()
	err = token.Error()
	if err != nil {
		client.Disconnect(10)
		setPollingError("mqtt", pe, err)
		return
	}
	mqttSubscribeMap.Store(pe.ID, client)
}

func checkMqttMessage(pe *datastore.PollingEnt, topic, payload string) (bool, error) {
	pe.Result["topic"] = topic
	pe.Result["payload"] = payload
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	vm.Set("topic", topic)
	vm.Set("payload", payload)
	extractor := pe.Extractor
	script := pe.Script
	if extractor == "goquery" {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(payload))
		if err != nil {
			return false, err
		}
		vm.Set("goquery", func(call otto.FunctionCall) otto.Value {
			if call.Argument(0).IsString() {
				sel := call.Argument(0).String()
				if ov, err := otto.ToValue(doc.Find(sel).Text()); err == nil {
					return ov
				}
			}
			return otto.UndefinedValue()
		})
	} else if extractor == "getBody" {
		vm.Set("getBody", func(call otto.FunctionCall) otto.Value {
			if r, err := otto.ToValue(payload); err == nil {
				return r
			}
			return otto.UndefinedValue()
		})
	} else if extractor == "jsonpath" {
		var res map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &res); err != nil {
			return false, err
		}
		vm.Set("jsonpath", func(call otto.FunctionCall) otto.Value {
			if call.Argument(0).IsString() {
				sel := call.Argument(0).String()
				if v, err := jsonpath.Get(sel, res); err == nil {
					if ov, err := otto.ToValue(v); err == nil {
						return ov
					}
				}
			}
			return otto.UndefinedValue()
		})
	} else if extractor != "" {
		grokEnt := datastore.GetGrokEnt(extractor)
		if grokEnt == nil {
			return false, fmt.Errorf("no grok pattern")
		}
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
			return false, fmt.Errorf("no grok pattern err=%v", err)
		}
		cap := fmt.Sprintf("%%{%s}", extractor)
		values, err := g.Parse(cap, payload)
		if err != nil {
			return false, err
		}
		for k, v := range values {
			vm.Set(k, v)
			pe.Result[k] = v
		}
	}
	if script == "" {
		return true, nil
	}
	value, err := vm.Run(script)
	if err != nil {
		return false, err
	}
	if ok, _ := value.ToBoolean(); ok {
		return true, nil
	}
	return false, nil
}

func getMqttClient(pe *datastore.PollingEnt) (mqtt.Client, error) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		return nil, fmt.Errorf("node not found id=%x", pe.NodeID)
	}
	opts := mqtt.NewClientOptions()
	opts.SetClientID("TWSNMP_FC_" + pe.ID)
	if pe.Params == "" {
		opts.AddBroker(fmt.Sprintf("tcp://%s:1883", n.IP))
	} else {
		if strings.HasPrefix(pe.Params, "ssl:/") {
			a := strings.SplitN(pe.Params, ",", 2)
			insec := len(a) == 2 && a[1] == "insecure"
			opts.AddBroker(a[0])
			// TLS
			tlsConfig, err := getMqttTLSConfig(insec)
			if err != nil {
				return nil, err
			}
			opts.SetTLSConfig(tlsConfig)
		} else {
			opts.AddBroker(pe.Params)
		}
	}
	if n.User != "" && n.Password != "" {
		opts.SetUsername(n.User)
		opts.SetPassword(n.Password)
	}
	opts.SetConnectTimeout(time.Duration(pe.Timeout) * time.Second)
	opts.SetPingTimeout(time.Duration(pe.Timeout) * time.Second)
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("mqtt connection lost err=%v", err)
		MqttStopSubscription(pe.ID)
	}
	return mqtt.NewClient(opts), nil
}

func getMqttTLSConfig(insec bool) (*tls.Config, error) {
	if datastore.CACert == "" {
		return &tls.Config{
			InsecureSkipVerify: insec,
		}, nil
	}
	certpool := x509.NewCertPool()
	ca, err := os.ReadFile(datastore.CACert)
	if err != nil {
		return nil, err
	}
	certpool.AppendCertsFromPEM(ca)
	if datastore.ClientCert == "" && datastore.ClientKey == "" {
		return &tls.Config{
			RootCAs:            certpool,
			InsecureSkipVerify: insec,
		}, nil
	}
	// mTLS
	cert, err := tls.LoadX509KeyPair(datastore.ClientCert, datastore.ClientKey)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            certpool,
		InsecureSkipVerify: insec,
	}, nil

}

func mqttStopAllSubscription() {
	log.Println("stop all mqtt subscribe")
	mqttSubscribeMap.Range(func(id, v any) bool {
		if c, ok := v.(mqtt.Client); ok {
			c.Disconnect(100)
		}
		mqttSubscribeMap.Delete(id)
		return true
	})
}

func MqttStopSubscription(id string) {
	log.Println("check mqtt stop")
	if v, ok := mqttSubscribeMap.Load(id); ok {
		if c, ok := v.(mqtt.Client); ok {
			log.Printf("stop mqtt subscribe %s", id)
			c.Disconnect(100)
		}
		mqttSubscribeMap.Delete(id)
	}
}
