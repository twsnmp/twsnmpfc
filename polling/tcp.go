package polling

// TCP/HTTP(S)/TLSのポーリングを行う。

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
)

func doPollingTCP(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		return
	}
	ok := false
	var rTime int64
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := net.DialTimeout("tcp", n.IP+":"+pe.Params, time.Duration(pe.Timeout)*time.Second)
		endTime := time.Now().UnixNano()
		if err != nil {
			pe.Result["error"] = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		ok = true
	}
	pe.Result["rtt"] = float64(rTime)
	if ok {
		delete(pe.Result, "error")
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

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
	vm.Set("setResult", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() && call.Argument(1).IsNumber() {
			n := call.Argument(0).String()
			if v, err := call.Argument(1).ToFloat(); err == nil {
				pe.Result[n] = v
			}
		}
		return otto.Value{}
	})
	vm.Set("status", status)
	vm.Set("code", code)
	vm.Set("rtt", rTime)
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
		for k, v := range getMetrics(body) {
			pe.Result[k] = v
			vm.Set(k, v)
		}
	}
	extractor := pe.Extractor
	script := pe.Script
	if extractor == "" {
		value, err := vm.Run(script)
		if err != nil {
			return false, err
		}
		if ok, _ := value.ToBoolean(); ok {
			return true, nil
		}
		return false, nil
	}
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
	value, err := vm.Run(script)
	if err != nil {
		return false, err
	}
	if ok, _ := value.ToBoolean(); ok {
		return true, nil
	}
	return false, nil
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

func getMetrics(body string) map[string]any {
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
		}
	} else {
		if err := json.Unmarshal([]byte(body), &r); err != nil {
			log.Printf("getMetrics other err=%v", err)
		}
	}
	return r
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

func doPollingTLS(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("tls", pe, fmt.Errorf("node not found"))
		return
	}
	mode := pe.Mode
	if mode == "" {
		mode = "verify"
	}
	target := pe.Params
	if target == "" {
		target = n.IP + ":443"
	} else if !strings.Contains(target, ":") {
		target = n.IP + ":" + target
	}
	script := pe.Script
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	switch mode {
	case "verify":
		conf.InsecureSkipVerify = false
	case "version":
		if strings.Contains(script, "1.0") {
			conf.MaxVersion = tls.VersionTLS10
		} else if strings.Contains(script, "1.1") {
			conf.MinVersion = tls.VersionTLS11
			conf.MaxVersion = tls.VersionTLS11
		} else if strings.Contains(script, "1.2") {
			conf.MinVersion = tls.VersionTLS12
			conf.MaxVersion = tls.VersionTLS12
		} else if strings.Contains(script, "1.3") {
			conf.MinVersion = tls.VersionTLS13
			conf.MaxVersion = tls.VersionTLS13
		}
	}
	d := &net.Dialer{
		Timeout: time.Duration(pe.Timeout) * time.Second,
	}
	ok := false
	var rTime int64
	var cs tls.ConnectionState
	lr := make(map[string]string)
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := tls.DialWithDialer(d, "tcp", target, conf)
		endTime := time.Now().UnixNano()
		if err != nil {
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		cs = conn.ConnectionState()
		ok = true
	}
	pe.Result["rtt"] = float64(rTime)
	if ok {
		getTLSConnectioStateInfo(pe, n.Name, &cs)
		if mode == "expire" {
			var d int
			if _, err := fmt.Sscanf(script, "%d", &d); err == nil && d > 0 {
				a := strings.SplitN(target, ":", 2)
				cert := getServerCert(a[0], &cs)
				if cert != nil {
					na := cert.NotAfter.Unix()
					pe.Result["notafter"] = cert.NotAfter.Format("2006/01/02")
					pe.Result["issuer"] = cert.Issuer.String()
					pe.Result["subject"] = cert.Subject.String()
					ct := time.Now().AddDate(0, 0, d).Unix()
					if ct > na {
						ok = false
					}
				} else {
					ok = false
				}
			}
		}
	}
	if (ok && !strings.Contains(script, "!")) || (!ok && strings.Contains(script, "!")) {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func getServerCert(host string, cs *tls.ConnectionState) *x509.Certificate {
	for _, cl := range cs.VerifiedChains {
		for _, c := range cl {
			if c.VerifyHostname(host) == nil {
				return c
			}
		}
	}
	for _, c := range cs.PeerCertificates {
		if c.VerifyHostname(host) == nil {
			return c
		}
	}
	return nil
}

func getTLSConnectioStateInfo(pe *datastore.PollingEnt, host string, cs *tls.ConnectionState) {
	switch cs.Version {
	case tls.VersionSSL30:
		pe.Result["version"] = "SSLv3"
	case tls.VersionTLS10:
		pe.Result["version"] = "TLSv1.0"
	case tls.VersionTLS11:
		pe.Result["version"] = "TLSv1.1"
	case tls.VersionTLS12:
		pe.Result["version"] = "TLSv1.2"
	case tls.VersionTLS13:
		pe.Result["version"] = "TLSv1.3"
	default:
		pe.Result["version"] = "Unknown"
	}
	id := fmt.Sprintf("%04x", cs.CipherSuite)
	if n, ok := datastore.GetCipherSuiteName(id); ok {
		pe.Result["cipherSuite"] = n
	} else {
		pe.Result["cipherSuite"] = id
	}
	if len(cs.VerifiedChains) > 0 {
		pe.Result["valid"] = "true"
	} else {
		pe.Result["valid"] = "false"
	}
	if cert := getServerCert(host, cs); cert != nil {
		pe.Result["issuer"] = cert.Issuer.String()
		pe.Result["subject"] = cert.Subject.String()
		pe.Result["notAfter"] = cert.NotAfter.Format("2006/01/02")
		pe.Result["subjectKeyID"] = fmt.Sprintf("%x", cert.SubjectKeyId)
	}
}

func autoAddTCPPolling(n *datastore.NodeEnt, pt *datastore.PollingTemplateEnt) {
	ports := strings.Split(pt.AutoMode, ",")
	for _, port := range ports {
		if !checkTCPConnect(n, port) {
			continue
		}
		p := new(datastore.PollingEnt)
		p.NodeID = n.ID
		p.Type = pt.Type
		if pt.Type == "http" {
			p.Name = pt.Name + " : " + port
			p.Params = "http"
			if pt.Mode == "https" {
				p.Params += "s"
			}
			p.Params += "://" + n.IP + ":" + port
		} else {
			sn := "tcp/" + port
			if nport, err := strconv.ParseInt(port, 10, 64); err == nil {
				if s, ok := datastore.GetServiceName(6, int(nport)); ok {
					sn = s
				}
			}
			p.Name = pt.Name + " : " + sn
			p.Params = port
		}
		if hasSameNamePolling(n.ID, p.Name) {
			continue
		}
		p.Mode = pt.Mode
		p.Script = pt.Script
		p.Extractor = pt.Extractor
		p.Filter = pt.Filter
		p.Level = pt.Level
		p.PollInt = datastore.MapConf.PollInt
		p.Timeout = datastore.MapConf.Timeout
		p.Retry = datastore.MapConf.Timeout
		p.LogMode = 0
		p.NextTime = 0
		p.State = "unknown"
		if err := datastore.AddPolling(p); err != nil {
			return
		}
	}
}

func checkTCPConnect(n *datastore.NodeEnt, port string) bool {
	conn, err := net.DialTimeout("tcp", n.IP+":"+port, time.Duration(datastore.MapConf.Timeout)*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
