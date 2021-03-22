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
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := net.DialTimeout("tcp", n.IP+":"+pe.Params, time.Duration(pe.Timeout)*time.Second)
		endTime := time.Now().UnixNano()
		if err != nil {
			log.Printf("doPollingTCP err=%v", err)
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
			log.Printf("doPollingHTTP err=%v", err)
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
	vm.Set("status", status)
	vm.Set("code", code)
	vm.Set("rtt", rTime)
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
	if pe.Mode == "https" {
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
	mode := pe.Mode
	if mode == "" {
		mode = "verify"
	}
	target := pe.Params
	if target == "" {
		target = n.IP + ":443"
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
			log.Printf("doPollingTLS err=%v", err)
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
