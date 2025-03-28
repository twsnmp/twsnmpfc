package polling

// Pi-Holeのポーリングを行う。

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/davidebianchi/go-jsonclient"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func docPollingPiHole(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("pihole", pe, fmt.Errorf("node not found"))
	}
	url := pe.Params
	if url == "" {
		url = fmt.Sprintf("http://%s", n.IP)
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	var err error
	var sid string
	var rTime int64
	for i := 0; ; i++ {
		startTime := time.Now().UnixNano()
		sid, err = loginToPiHole(url, n.Password, pe.Timeout)
		endTime := time.Now().UnixNano()
		if err == nil {
			break
		}
		if i > pe.Retry {
			setPollingError("pihole", pe, err)
			return
		}
		rTime = endTime - startTime
	}
	defer logoutFromPiHole(url, sid, pe.Timeout)
	res, err := getPiHole(url, pe.Mode, sid, pe.Timeout)
	if err != nil {
		setPollingError("pihole", pe, err)
		return
	}
	pe.Result["rtt"] = float64(rTime)
	vm.Set("rtt", rTime)
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

	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("pihole", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		delete(pe.Result, "error")
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func loginToPiHole(url, passwd string, timeout int) (string, error) {
	opts := jsonclient.Options{
		BaseURL: url + "/api/",
		HTTPClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
	client, err := jsonclient.New(opts)
	if err != nil {
		return "", err
	}
	var reqBody = map[string]interface{}{
		"password": passwd,
	}
	req, err := client.NewRequest(http.MethodPost, "auth", reqBody)
	if err != nil {
		return "", err
	}
	var respBody = make(map[string]interface{})
	resp, err := client.Do(req, &respBody)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %s", resp.Status)
	}
	v, err := jsonpath.Get("$.session.sid", respBody)
	if err != nil {
		return "", err
	}
	if sid, ok := v.(string); ok {
		return sid, nil
	}
	return "", fmt.Errorf("sid not found %v", v)
}

func logoutFromPiHole(url, sid string, timeout int) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest("DELETE", url+"/api/auth?sid="+sid, nil)
	if err != nil {
		log.Printf("logoutFromPiHole err=%v", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("logoutFromPiHole err=%v", err)
		return
	}
	defer resp.Body.Close()
}

func getPiHole(url, path, sid string, timeout int) (map[string]interface{}, error) {
	var respBody = make(map[string]interface{})
	opts := jsonclient.Options{
		BaseURL: url + "/api/",
		Headers: jsonclient.Headers{
			"sid": sid,
		},
		HTTPClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
	client, err := jsonclient.New(opts)
	if err != nil {
		return respBody, err
	}
	req, err := client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return respBody, err
	}
	_, err = client.Do(req, &respBody)
	return respBody, err
}
