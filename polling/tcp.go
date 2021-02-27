package polling

// TCP/HTTP(S)/TLSのポーリングを行う。

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
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
		log.Printf("node not found nodeID=%s", pe.NodeID)
		return
	}
	ok := false
	var rTime int64
	lr := make(map[string]string)
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := net.DialTimeout("tcp", n.IP+":"+pe.Polling, time.Duration(pe.Timeout)*time.Second)
		endTime := time.Now().UnixNano()
		if err != nil {
			log.Printf("doPollingTCP err=%v", err)
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		ok = true
	}
	pe.LastVal = float64(rTime)
	if ok {
		delete(lr, "error")
		lr["rtt"] = fmt.Sprintf("%f", pe.LastVal)
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	pe.LastResult = makeLastResult(lr)
}

func doPollingHTTP(pe *datastore.PollingEnt) {
	var ok bool
	var err error
	cmd := splitCmd(pe.Polling)
	if len(cmd) < 1 {
		setPollingError("http", pe, fmt.Errorf("no url"))
		return
	}
	url := cmd[0]
	ok = false
	var rTime int64
	body := ""
	status := ""
	code := 0
	lr := make(map[string]string)
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		status, body, code, err = doHTTPGet(pe, url)
		endTime := time.Now().UnixNano()
		if err != nil {
			log.Printf("doPollingHTTP err=%v", err)
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	pe.LastVal = float64(rTime)
	if len(cmd) > 2 {
		ok, lr, err = checkHTTPResp(pe, cmd[1], cmd[2], status, body, code)
		if err != nil {
			setPollingError("http", pe, err)
			return
		}
		if v, ok := lr["numVal"]; ok {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				pe.LastVal = f
			}
		}
	} else {
		lr["rtt"] = fmt.Sprintf("%f", pe.LastVal)
		lr["status"] = status
	}
	if ok {
		delete(lr, "error")
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
	pe.LastResult = makeLastResult(lr)
}

func checkHTTPResp(pe *datastore.PollingEnt, extractor, script, status, body string, code int) (bool, map[string]string, error) {
	lr := make(map[string]string)
	vm := otto.New()
	lr["status"] = status
	lr["code"] = fmt.Sprintf("%d", code)
	lr["rtt"] = fmt.Sprintf("%f", pe.LastVal)
	_ = vm.Set("status", status)
	_ = vm.Set("code", code)
	_ = vm.Set("rtt", pe.LastVal)
	if extractor == "" {
		value, err := vm.Run(script)
		if err != nil {
			return false, lr, err
		}
		pe.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); ok {
			return true, lr, nil
		}
		return false, lr, nil
	}
	grokEnt := datastore.GetGrokEnt(extractor)
	if grokEnt == nil {
		return false, lr, fmt.Errorf("no grok pattern")
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
		return false, lr, fmt.Errorf("no grok pattern err=%v", err)
	}
	cap := fmt.Sprintf("%%{%s}", extractor)
	values, err := g.Parse(cap, body)
	if err != nil {
		return false, lr, err
	}
	for k, v := range values {
		_ = vm.Set(k, v)
		lr[k] = v
	}
	value, err := vm.Run(script)
	if err != nil {
		return false, lr, err
	}
	if ok, _ := value.ToBoolean(); ok {
		return true, lr, nil
	}
	return false, lr, nil
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
	body := make([]byte, 64*1024)
	if pe.Type == "https" {
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return "", "", 0, err
		}
		defer resp.Body.Close()
		_, err = resp.Body.Read(body)
		if err == io.EOF {
			err = nil
		}
		return resp.Status, string(body), resp.StatusCode, err
	}
	resp, err := insecureClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", "", 0, err
	}
	defer resp.Body.Close()
	_, err = resp.Body.Read(body)
	if err == io.EOF {
		err = nil
	}
	return resp.Status, string(body), resp.StatusCode, err
}

func doPollingTLS(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("tls", pe, fmt.Errorf("node not found"))
		return
	}
	cmd := splitCmd(pe.Polling)
	mode := "verify"
	target := n.IP + ":443"
	script := ""
	if len(cmd) > 2 {
		mode = cmd[0]
		target = cmd[1]
		script = cmd[2]
	}

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
			log.Printf("doPollingTLS err=%v", err)
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		cs = conn.ConnectionState()
		ok = true
	}
	pe.LastVal = float64(rTime)
	if ok {
		lr = getTLSConnectioStateInfo(n.Name, &cs)
		lr["rtt"] = fmt.Sprintf("%f", pe.LastVal)
		if mode == "expire" {
			var d int
			if _, err := fmt.Sscanf(script, "%d", &d); err == nil && d > 0 {
				a := strings.SplitN(target, ":", 2)
				cert := getServerCert(a[0], &cs)
				if cert != nil {
					na := cert.NotAfter.Unix()
					lr["notafter"] = cert.NotAfter.Format("2006/01/02")
					lr["issuer"] = cert.Issuer.String()
					lr["subject"] = cert.Subject.String()
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
	pe.LastResult = makeLastResult(lr)
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

func getTLSConnectioStateInfo(host string, cs *tls.ConnectionState) map[string]string {
	ret := make(map[string]string)
	switch cs.Version {
	case tls.VersionSSL30:
		ret["version"] = "SSLv3"
	case tls.VersionTLS10:
		ret["version"] = "TLSv1.0"
	case tls.VersionTLS11:
		ret["version"] = "TLSv1.1"
	case tls.VersionTLS12:
		ret["version"] = "TLSv1.2"
	case tls.VersionTLS13:
		ret["version"] = "TLSv1.3"
	default:
		ret["version"] = "Unknown"
	}
	id := fmt.Sprintf("%04x", cs.CipherSuite)
	if n, ok := datastore.GetCipherSuiteName(id); ok {
		ret["cipherSuite"] = n
	} else {
		ret["cipherSuite"] = id
	}
	if len(cs.VerifiedChains) > 0 {
		ret["valid"] = "true"
	} else {
		ret["valid"] = "false"
	}
	if cert := getServerCert(host, cs); cert != nil {
		ret["issuer"] = cert.Issuer.String()
		ret["subject"] = cert.Subject.String()
		ret["notAfter"] = cert.NotAfter.Format("2006/01/02")
		ret["subjectKeyID"] = fmt.Sprintf("%x", cert.SubjectKeyId)
	}
	return ret
}
