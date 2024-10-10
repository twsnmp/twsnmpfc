package polling

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/openconfig/gnmic/pkg/api"
	"github.com/openconfig/gnmic/pkg/api/target"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingGNMI(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found id=%x", pe.NodeID)
		return
	}
	if pe.Script == "" {
		setPollingError("gnmi", pe, fmt.Errorf("gnmi no script"))
		return
	}
	target := pe.Params
	if target == "" {
		target = fmt.Sprintf("%s:57400", n.IP)
	} else if p, err := strconv.Atoi(target); err == nil && p > 0 && p < 65535 {
		target = fmt.Sprintf("%s:%d", n.IP, p)
	}
	if pe.Mode == "subscribe" {
		doPollingGNMISubscribe(pe, n, target)
		return
	}
	doPollingGNMIGet(pe, n, target)
}

func doPollingGNMIGet(pe *datastore.PollingEnt, n *datastore.NodeEnt, target string) {
	tg, err := api.NewTarget(
		api.Name(n.Name),
		api.Address(target),
		api.Username(n.User),
		api.Password(n.Password),
		api.SkipVerify(true),
	)
	if err != nil {
		setPollingError("gnmi", pe, err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = tg.CreateGNMIClient(ctx)
	if err != nil {
		setPollingError("gnmi", pe, err)
		return
	}
	defer tg.Close()
	getReq, err := api.NewGetRequest(
		api.Path(pe.Filter),
		api.Encoding("json_ietf"))
	if err != nil {
		setPollingError("gnmi", pe, err)
		return
	}
	getResp, err := tg.Get(ctx, getReq)
	if err != nil {
		setPollingError("gnmi", pe, err)
		return
	}
	data := []byte{}
	for _, not := range getResp.GetNotification() {
		for _, u := range not.GetUpdate() {
			data = u.Val.GetJsonIetfVal()
			if len(data) > 0 {
				break
			}
		}
		if len(data) > 0 {
			break
		}
	}
	if len(data) < 1 {
		setPollingError("gnmi", pe, fmt.Errorf("json data not found"))
		return
	}
	vm := otto.New()
	addJavaScriptFunctions(pe, vm)
	vm.Set("data", string(data))
	vm.Set("now", time.Now().UnixMilli())
	if v, ok := pe.Result["data"]; ok {
		if j, ok := v.(string); ok {
			vm.Set("last_data", string(j))
		}
		if v, ok := pe.Result["last"]; ok {
			if l, ok := v.(int64); ok {
				vm.Set("last", l)
			}
		}
	}
	pe.Result = make(map[string]interface{})
	pe.Result["data"] = string(data)
	pe.Result["last"] = time.Now().UnixMilli()
	value, err := vm.Run(pe.Script)
	if err != nil {
		log.Printf("gnmi polling err=%v", err)
		setPollingError("gnmi", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); !ok {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "normal")
}

var gNMISubscribeMap = sync.Map{}

func doPollingGNMISubscribe(pe *datastore.PollingEnt, n *datastore.NodeEnt, target string) {
	tg, err := api.NewTarget(
		api.Name(n.Name),
		api.Address(target),
		api.Username(n.User),
		api.Password(n.Password),
		api.SkipVerify(true),
	)
	if err != nil {
		setPollingError("gnmi", pe, err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = tg.CreateGNMIClient(ctx)
	if err != nil {
		setPollingError("gnmi", pe, err)
		return
	}
	defer tg.Close()
	subReq, err := api.NewSubscribeRequest(
		api.Encoding("json_ietf"),
		api.SubscriptionListMode("stream"),
		api.Subscription(
			api.Path(pe.Filter),
			api.SubscriptionMode("on_change"),
		))
	if err != nil {
		setPollingError("gnmi", pe, err)
		return
	}
	go tg.Subscribe(ctx, subReq, pe.ID)
	gNMISubscribeMap.Store(pe.ID, tg)
	subRspChan, subErrChan := tg.ReadSubscriptions()
	for {
		select {
		case rsp := <-subRspChan:
			if p := datastore.GetPolling(pe.ID); p == nil {
				log.Printf("stop deleted gnmi subscribe polling %s", pe.ID)
				GNMIStopSubscription(pe.ID)
				return
			}
			if rsp.Response.GetUpdate() != nil {
				oldState := pe.State
				gNMISetSubscribeResp(pe, rsp)
				gNMIUpdatePolling(pe, oldState)
			}
		case tgErr := <-subErrChan:
			if _, ok := gNMISubscribeMap.Load(pe.ID); ok {
				log.Printf("polling %s subscription %q stopped: %v", pe.Name, tgErr.SubscriptionName, tgErr.Err)
				oldState := pe.State
				setPollingError("gnmi", pe, tgErr.Err)
				gNMIUpdatePolling(pe, oldState)
				gNMISubscribeMap.Delete(pe.ID)
			}
			return
		}
	}
}

func gNMISetSubscribeResp(pe *datastore.PollingEnt, rsp *target.SubscribeResponse) {
	data := []byte{}
	not := rsp.Response.GetUpdate()
	for _, u := range not.GetUpdate() {
		data = u.GetVal().GetJsonIetfVal()
		if len(data) > 0 {
			break
		}
	}
	if len(data) < 1 {
		setPollingError("gnmi", pe, fmt.Errorf("json data not found"))
		return
	}
	vm := otto.New()
	addJavaScriptFunctions(pe, vm)
	vm.Set("data", string(data))
	vm.Set("now", time.Now().UnixMilli())
	if v, ok := pe.Result["data"]; ok {
		if j, ok := v.(string); ok {
			vm.Set("last_data", string(j))
		}
		if v, ok := pe.Result["last"]; ok {
			if l, ok := v.(int64); ok {
				vm.Set("last", l)
			}
		}
	}
	pe.Result = make(map[string]interface{})
	pe.Result["data"] = string(data)
	pe.Result["last"] = time.Now().UnixMilli()
	value, err := vm.Run(pe.Script)
	if err != nil {
		log.Printf("gnmi polling err=%v", err)
		setPollingError("gnmi", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); !ok {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "normal")
}

func gNMIUpdatePolling(pe *datastore.PollingEnt, oldState string) {
	datastore.UpdatePolling(pe)
	if pe.LogMode == datastore.LogModeAlways || pe.LogMode == datastore.LogModeAI || (pe.LogMode == datastore.LogModeOnChange && oldState != pe.State) {
		if err := datastore.AddPollingLog(pe); err != nil {
			log.Printf("add polling log err=%v %#v", err, pe)
		}
	}
	if datastore.InfluxdbConf.PollingLog != "" {
		if datastore.InfluxdbConf.PollingLog == "all" || pe.LogMode != datastore.LogModeNone {
			if err := datastore.SendPollingLogToInfluxdb(pe); err != nil {
				log.Printf("send polling log to influxdb1 err=%v", err)
			}
		}
	}
}

func gNMIStopAllSubscription() {
	log.Println("stop all gnmi subscribe")
	gNMISubscribeMap.Range(func(id, v any) bool {
		if tg, ok := v.(*target.Target); ok {
			tg.StopSubscription(id.(string))
		}
		gNMISubscribeMap.Delete(id)
		return true
	})
}

// GNMIStopSubscription : stop gNMI subscribe polling
func GNMIStopSubscription(id string) {
	if v, ok := gNMISubscribeMap.Load(id); ok {
		if tg, ok := v.(*target.Target); ok {
			log.Printf("stop gnmi subscribe %s", id)
			tg.StopSubscription(id)
		}
		gNMISubscribeMap.Delete(id)
	}
}
