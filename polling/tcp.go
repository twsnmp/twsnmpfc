package polling

// TCP/HTTP(S)/TLSのポーリングを行う。

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
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
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := tls.DialWithDialer(d, "tcp", target, conf)
		endTime := time.Now().UnixNano()
		if err != nil {
			pe.Result["error"] = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		cs = conn.ConnectionState()
		ok = true
	}
	pe.Result["rtt"] = float64(rTime)
	if ok {
		host := n.Name
		if a := strings.SplitN(target, ":", 2); len(a) > 1 {
			host = a[0]
		}
		getTLSConnectioStateInfo(pe, host, &cs)
		if mode == "expire" {
			var d int
			if _, err := fmt.Sscanf(script, "%d", &d); err == nil && d > 0 {
				cert := getServerCert(host, &cs)
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
		delete(pe.Result, "error")
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
	if ip := net.ParseIP(host); ip != nil {
		host = "[" + host + "]"
	}
	for _, c := range cs.PeerCertificates {
		if c.VerifyHostname(host) == nil {
			return c
		}
	}
	if len(cs.PeerCertificates) > 0 {
		return cs.PeerCertificates[0]
	}
	return nil
}

func getTLSConnectioStateInfo(pe *datastore.PollingEnt, host string, cs *tls.ConnectionState) {
	switch cs.Version {
	case 0x0300: //tls.VersionSSL30 : ワーニングがでる
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
	pe.Result["valid"] = "false"
	if cert := getServerCert(host, cs); cert != nil {
		pe.Result["issuer"] = cert.Issuer.String()
		pe.Result["subject"] = cert.Subject.String()
		pe.Result["notAfter"] = cert.NotAfter.Format("2006/01/02")
		pe.Result["subjectKeyID"] = fmt.Sprintf("%x", cert.SubjectKeyId)
		if cert.NotAfter.Unix() > time.Now().Unix() {
			pe.Result["valid"] = "true"
		}
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
		if err := datastore.AddPollingWithDupCheck(p); err != nil {
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
