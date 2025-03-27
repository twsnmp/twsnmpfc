package polling

// TCP/HTTP(S)/TLSのポーリングを行う。

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func doPollingHTTP(pe *datastore.PollingEnt) {
	var ok bool
	var err error
	url := pe.Params
	ok = false
	var rTime int64
	body := ""
	status := ""
	code := 0
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		status, body, code, err = doHTTPGet(pe, url)
		endTime := time.Now().UnixNano()
		if err != nil {
			pe.Result["error"] = fmt.Sprintf("%v", err)
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	pe.Result["rtt"] = float64(rTime)
	pe.Result["status"] = status
	pe.Result["code"] = float64(code)
	if pe.Script != "" {
		ok, err = checkHTTPResp(pe, status, body, code, rTime)
		if err != nil {
			setPollingError("http", pe, err)
			return
		}
	}
	if ok {
		delete(pe.Result, "error")
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func checkHTTPResp(pe *datastore.PollingEnt, status, body string, code int, rTime int64) (bool, error) {
	vm := otto.New()
	addJavaScriptFunctions(pe, vm)
	vm.Set("status", status)
	vm.Set("code", code)
	vm.Set("rtt", rTime)
	if code < 200 || code >= 300 || body == "" {
		return false, fmt.Errorf("checkHTTPResp resp error code=%d,len(body)=%d", code, len(body))
	}
	if strings.Contains(pe.Mode, "metrics") {
		//前回の値と間隔をJavaScriptで処理できるようにする
		for _, k := range []string{"accepts", "handled", "requests"} {
			if i, ok := pe.Result[k]; ok {
				if v, ok := i.(float64); ok {
					vm.Set(k+"_last", v)
				}
			}
		}
		vm.Set("interval", pe.PollInt)
		if m, err := getMetrics(body); err == nil {
			delete(pe.Result, "error")
			for k, v := range m {
				pe.Result[k] = v
				vm.Set(k, v)
			}
		} else {
			return false, err
		}
	} else if pe.Mode == "hash" {
		nh := getHash(body)
		if oh, ok := pe.Result["sha256"]; !ok {
			pe.Result["first_sha256"] = nh
			pe.Result["sha256"] = nh
			vm.Set("last_sha256", nh)
			vm.Set("first_sha256", nh)
			vm.Set("sha256", nh)
		} else {
			pe.Result["sha256"] = nh
			vm.Set("sha256", nh)
			vm.Set("last_sha256", oh)
			if fh, ok := pe.Result["first_sha256"]; ok {
				vm.Set("first_sha256", fh)
			} else {
				pe.Result["first_sha256"] = oh
				vm.Set("first_sha256", oh)
			}
		}
	}
	extractor := pe.Extractor
	script := pe.Script
	if extractor == "goquery" {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
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
			if r, err := otto.ToValue(body); err == nil {
				return r
			}
			return otto.UndefinedValue()
		})
	} else if extractor == "jsonpath" {
		var res map[string]interface{}
		if err := json.Unmarshal([]byte(body), &res); err != nil {
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
		values, err := g.Parse(cap, body)
		if err != nil {
			return false, err
		}
		for k, v := range values {
			vm.Set(k, v)
			pe.Result[k] = v
		}
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

func getHash(body string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(body)))
}

type fiberMetricsEnt struct {
	PID struct {
		CPU   float64 `json:"cpu"`
		RAM   float64 `json:"ram"`
		Conns float64 `json:"conns"`
	} `json:"pid"`
	OS struct {
		CPU      float64 `json:"cpu"`
		RAM      float64 `json:"ram"`
		Conns    float64 `json:"conns"`
		TotalRAM float64 `json:"total_ram"`
		LoadAvg  float64 `json:"load_avg"`
	} `json:"os"`
}

var numReg = regexp.MustCompile(`[\.0-9]`)
var nginxAConnsReg = regexp.MustCompile(`Active connections: (\d+)`)
var nginxAHRReg = regexp.MustCompile(`\s*(\d+)\s+(\d+)\s+(\d+)`)
var nginxRWWReg = regexp.MustCompile(`Reading:\s*(\d+)\s+Writing:\s*(\d+)Waiting:\s*(\d+)`)

func getMetrics(body string) (map[string]any, error) {
	r := make(map[string]any)
	if strings.Contains(body, "Apache") {
		// Apache mod_status
		for _, l := range strings.Split(body, "\n") {
			l = strings.TrimSpace(l)
			a := strings.SplitN(l, ": ", 2)
			if len(a) != 2 {
				continue
			}
			if numReg.MatchString(a[1]) {
				if v, err := strconv.ParseFloat(a[1], 64); err == nil {
					r[a[0]] = v
					continue
				}
			}
			r[a[0]] = a[1]
		}
	} else if strings.Contains(body, "Active conn") {
		// NGINX
		if a := nginxAConnsReg.FindStringSubmatch(body); len(a) == 2 && a[1] != "" {
			if v, err := strconv.ParseFloat(a[1], 64); err == nil {
				r["active_connectios"] = v
			}
		}
		if a := nginxAHRReg.FindStringSubmatch(body); len(a) == 4 && a[1] != "" {
			if v, err := strconv.ParseFloat(a[1], 64); err == nil {
				r["accepts"] = v
			}
			if v, err := strconv.ParseFloat(a[2], 64); err == nil {
				r["handled"] = v
			}
			if v, err := strconv.ParseFloat(a[3], 64); err == nil {
				r["requests"] = v
			}
		}
		if a := nginxRWWReg.FindStringSubmatch(body); len(a) == 4 && a[1] != "" {
			if v, err := strconv.ParseFloat(a[1], 64); err == nil {
				r["Reading"] = v
			}
			if v, err := strconv.ParseFloat(a[2], 64); err == nil {
				r["Writing"] = v
			}
			if v, err := strconv.ParseFloat(a[3], 64); err == nil {
				r["Waiting"] = v
			}
		}
	} else if strings.Contains(body, "{\"pid\"") {
		var s fiberMetricsEnt
		body = strings.TrimSpace(body)
		if err := json.Unmarshal([]byte(body), &s); err == nil {
			r["pid_cpu"] = s.PID.CPU
			r["pid_ram"] = s.PID.RAM
			r["pid_conns"] = s.PID.Conns
			r["os_cpu"] = s.OS.CPU
			r["os_ram"] = s.OS.RAM
			r["os_conns"] = s.OS.Conns
			r["os_load_avg"] = s.OS.LoadAvg
			r["os_total_ram"] = s.OS.TotalRAM
		} else {
			log.Printf("getMetrics fiber err=%v", err)
			return r, err
		}
	} else {
		if err := json.Unmarshal([]byte(body), &r); err != nil {
			log.Printf("getMetrics other '%s' err=%v", body, err)
			return r, err
		}
	}
	return r, nil
}

var insecureTransport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var insecureClient = &http.Client{Transport: insecureTransport}

func doHTTPGet(pe *datastore.PollingEnt, url string) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout)*time.Second)
	defer cancel()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", 0, err
	}
	if pe.Mode == "metrics/json" {
		req.Header.Set("Accept", "application/json")
	}
	if pe.Mode == "https" {
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return "", "", 0, err
		}
		defer resp.Body.Close()
		// メモリー不足をおこさないための64MBまで
		if resp.ContentLength > 1024*1024*64 {
			return "", "", 0, fmt.Errorf("http rest seize over len=%d", resp.ContentLength)
		}
		if body, err := io.ReadAll(resp.Body); err == nil {
			return resp.Status, string(body), resp.StatusCode, err
		} else {
			if err == io.EOF {
				err = nil
			}
			return resp.Status, "", resp.StatusCode, err
		}
	}
	resp, err := insecureClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", "", 0, err
	}
	defer resp.Body.Close()
	if resp.ContentLength > 1024*1024*64 {
		return "", "", 0, fmt.Errorf("http resp seize over len=%d", resp.ContentLength)
	}
	if body, err := io.ReadAll(resp.Body); err == nil {
		return resp.Status, string(body), resp.StatusCode, err
	} else {
		if err == io.EOF {
			err = nil
		}
		return resp.Status, "", resp.StatusCode, err
	}
}
