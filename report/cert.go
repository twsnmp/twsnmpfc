package report

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func DoCheckCert() {
	checkCertCh <- true
}

func checkCerts() {
	ct := time.Now().Add(time.Hour * -24).UnixNano()
	datastore.ForEachCerts(func(c *datastore.CertEnt) bool {
		if c.LastTime < ct {
			go getCert(c)
		}
		return true
	})
}

func setCertPenalty(c *datastore.CertEnt) {
	now := time.Now().Unix()
	c.Penalty = 0
	if c.Error != "" {
		c.Penalty++
	}
	if c.LastTime == 0 {
		c.Penalty++
		return
	}
	if !c.Verify {
		c.Penalty++
	}
	if c.NotAfter < now+3600*24*30 {
		c.Penalty++
		if c.NotAfter < now+3600*24*7 {
			c.Penalty++
			if c.NotAfter < now {
				log.Println("expire cert")
				c.Penalty++
			}
		}
	} else if c.NotAfter-c.NotBefore > 3600*24*825 {
		// 証明書の期間は２年
		log.Println("long tem cert")
		c.Penalty++
	}
	if c.Subject == c.Issuer {
		// 自己署名
		log.Println("self sign cert")
		c.Penalty++
	}
}

func getCert(c *datastore.CertEnt) {
	target := fmt.Sprintf("%s:%d", c.Target, c.Port)
	c.Verify = false
	c.Error = ""
	conf := &tls.Config{
		InsecureSkipVerify: false,
	}
	d := &net.Dialer{
		Timeout: time.Duration(datastore.MapConf.Timeout) * time.Second,
	}
	for i := 0; i <= datastore.MapConf.Retry; i++ {
		conn, err := tls.DialWithDialer(d, "tcp", target, conf)
		if err != nil {
			c.Error = fmt.Sprintf("%v", err)
			switch err := err.(type) {
			case *net.OpError:
				log.Printf("getCert err=%v", err)
			default:
				conf.InsecureSkipVerify = true
				log.Printf("getCert set skip err=%v", err)
			}
			continue
		}
		defer conn.Close()
		cs := conn.ConnectionState()
		if cs.HandshakeComplete {
			if cert := getServerCert(c.Target, &cs); cert != nil {
				c.SerialNumber = cert.SerialNumber.String()
				c.Subject = cert.Subject.String()
				c.Issuer = cert.Issuer.String()
				c.NotAfter = cert.NotAfter.Unix()
				c.NotBefore = cert.NotBefore.Unix()
				c.Verify = !conf.InsecureSkipVerify
				if c.FirstTime == 0 {
					c.FirstTime = time.Now().UnixNano()
				}
				c.LastTime = time.Now().UnixNano()
			} else {
				c.Error = "no cert"
			}
		} else {
			c.Error = "TLS error"
		}
		c.UpdateTime = time.Now().UnixNano()
		setCertPenalty(c)
		return
	}
}

// getServerCert : サーバー証明書を取得する
func getServerCert(t string, cs *tls.ConnectionState) *x509.Certificate {
	if len(cs.VerifiedChains) > 0 && cs.ServerName != "" {
		for _, cl := range cs.VerifiedChains {
			for _, c := range cl {
				if c.VerifyHostname(cs.ServerName) == nil {
					return c
				}
			}
		}
	}
	if ip := net.ParseIP(t); ip != nil {
		t = "[" + t + "]"
	}
	for _, c := range cs.PeerCertificates {
		if c.VerifyHostname(t) == nil {
			return c
		}
	}
	if len(cs.PeerCertificates) > 0 {
		log.Println("not hit return PeerCertificates[0]")
		return cs.PeerCertificates[0]
	}
	return nil
}

func ResetCertScore() {
	datastore.ForEachCerts(func(c *datastore.CertEnt) bool {
		c.Penalty = 0
		setCertPenalty(c)
		c.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcCertScore()
}
