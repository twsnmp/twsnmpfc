package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

// CreateCAReq represents a request from the frontend.
type CreateCAReq struct {
	RootCAKeyType string `json:"RootCAKeyType"`
	Name          string `json:"Name"`
	SANs          string `json:"SANs"`
	AcmePort      int    `json:"AcmePort"`
	HttpBaseURL   string `json:"HttpBaseURL"`
	AcmeBaseURL   string `json:"AcmeBaseURL"`
	HttpPort      int    `json:"HttpPort"`
	RootCATerm    int    `json:"RootCATerm"`
	CrlInterval   int    `json:"CrlInterval"`
	CertTerm      int    `json:"CertTerm"`
}

type PKIControlEnt struct {
	AcmeBaseURL string `json:"AcmeBaseURL"`
	EnableAcme  bool   `json:"EnableAcme"`
	EnableHttp  bool   `json:"EnableHttp"`
	AcmeStatus  string `json:"AcmeStatus"`
	HttpStatus  string `json:"HttpStatus"`
	CrlInterval int    `json:"CrlInterval"`
	CertTerm    int    `json:"CertTerm"`
}

// PKIConfEnt is the configuration data for a CA stored in the DB.
type PKIConfEnt struct {
	Name           string `json:"Name"`
	SANs           string `json:"AcmeSANs"`
	RootCAKeyType  string `json:"RootCAKeyType"`
	RootCAKey      string `json:"RootCAKey"`
	RootCACert     string `json:"RootCACert"`
	RootCATerm     int    `json:"RootCATerm"`
	CertTerm       int    `json:"CertTerm"`
	Serial         int64  `json:"Serial"`
	AcmeServerKey  string `json:"AcmeServerKey"`
	AcmeServerCert string `json:"AcmeServerCert"`
	AcmeBaseURL    string `json:"AcmeBaseURL"`
	AcmePort       int    `json:"AcmePort"`
	HttpBaseURL    string `json:"HttpBaseURL"`
	HttpPort       int    `json:"HttpPort"`
	ScepCAKey      string `json:"ScepCAKey"`
	ScepCACert     string `json:"ScepCACert"`
	CrlNumber      int64  `json:"CrlNumber"`
	CrlInterval    int    `json:"CrlInterval"`
	EnableAcme     bool   `json:"EnableAcme"`
	EnableHttp     bool   `json:"EnableHttp"`
}

type PKICertEnt struct {
	ID          string            `json:"ID"`
	Subject     string            `json:"Subject"`
	NodeID      string            `json:"NodeID"`
	Created     int64             `json:"Created"`
	Revoked     int64             `json:"Revoked"`
	Expire      int64             `json:"Expire"`
	Type        string            `json:"Type"`
	Certificate string            `json:"Certificate"`
	Info        map[string]string `json:"Info"`
}

var PKIConf PKIConfEnt

func loadPKIConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	setDefaultPKIConf()
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("pkiConf"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &PKIConf)
	})
}

func setDefaultPKIConf() {
	PKIConf = PKIConfEnt{
		CrlInterval:   24,
		RootCAKeyType: "ecdsa-256",
		RootCATerm:    10,
		CertTerm:      365 * 24,
		HttpPort:      8082,
		AcmePort:      8083,
		Serial:        time.Now().UnixNano(),
		CrlNumber:     1,
		SANs:          getDefaultSANs(),
	}
}

func InitCAConf(req *CreateCAReq) error {
	PKIConf.RootCAKeyType = req.RootCAKeyType
	PKIConf.AcmePort = req.AcmePort
	PKIConf.HttpPort = req.HttpPort
	PKIConf.HttpBaseURL = req.HttpBaseURL
	PKIConf.Name = req.Name
	if req.SANs == "" {
		PKIConf.SANs = getDefaultSANs()
	} else {
		PKIConf.SANs = req.SANs
	}
	if req.AcmeBaseURL == "" {
		baseURL := "https://"
		if a := strings.Split(PKIConf.SANs, ","); len(a) > 0 {
			baseURL += a[0]
		} else {
			if h, err := os.Hostname(); err == nil {
				baseURL += h
			} else {
				baseURL += "localhost"
			}
		}
		baseURL += fmt.Sprintf(":%d", PKIConf.AcmePort)
		PKIConf.AcmeBaseURL = baseURL
	} else {
		PKIConf.AcmeBaseURL = req.AcmeBaseURL
	}
	PKIConf.RootCATerm = req.RootCATerm
	PKIConf.CertTerm = req.CertTerm
	PKIConf.CrlInterval = req.CrlInterval
	if PKIConf.AcmePort < 1 || PKIConf.AcmePort > 0xfffe {
		PKIConf.AcmePort = 8083
	}
	if PKIConf.HttpPort < 1 || PKIConf.HttpPort > 0xfffe {
		PKIConf.HttpPort = 8082
	}
	if PKIConf.CertTerm < 1 {
		PKIConf.CertTerm = 24 * 365
	}
	SavePKIConf()
	return nil
}

func ClearCAData() {
	setDefaultPKIConf()
	SavePKIConf()
	db.Batch(func(tx *bbolt.Tx) error {
		tx.DeleteBucket([]byte("certs"))
		_, err := tx.CreateBucket([]byte("certs"))
		return err
	})
}

func ForEachCert(cb func(c *PKICertEnt) bool) {
	if db == nil {
		return
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("certs"))
		b.ForEach(func(k []byte, v []byte) error {
			var cert PKICertEnt
			if err := json.Unmarshal(v, &cert); err != nil {
				log.Printf("loadPKI err=%v", err)
				return nil
			}
			if !cb(&cert) {
				return fmt.Errorf("cancel by reader")
			}
			return nil
		})
		return nil
	})
}

func UpdateCert(cert *PKICertEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("certs"))
		j, err := json.Marshal(cert)
		if err != nil {
			return err
		}
		b.Put([]byte(cert.ID), j)
		return nil
	})
}

func RevokeCert(cert *PKICertEnt) error {
	cert.Revoked = time.Now().UnixNano()
	AddEventLog(&EventLogEnt{
		Time:  cert.Revoked,
		Level: "low",
		Type:  "ca",
		Event: fmt.Sprintf("証明書を失効しました subject=%s serial=%s", cert.Subject, cert.ID),
	})
	return UpdateCert(cert)
}

func RevokeCertByID(id string) error {
	cert := FindCert(id)
	if cert == nil {
		return nil
	}
	return RevokeCert(cert)
}

func FindCert(id string) *PKICertEnt {
	var cert PKICertEnt
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("certs"))
		if j := b.Get([]byte(id)); j != nil {
			return json.Unmarshal(j, &cert)
		}
		return fmt.Errorf("cert not found")
	})
	if err != nil {
		return nil
	}
	return &cert
}

func SavePKIConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	log.Println("savePKIConf")
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if j, err := json.Marshal(&PKIConf); err == nil {
			b.Put([]byte("pkiConf"), j)
		}
		return nil
	})
}

// getDefaultSANs retunrs  my IP and Host name
func getDefaultSANs() string {
	sans := []string{}
	if n, err := os.Hostname(); err == nil {
		sans = append(sans, n)
	}
	if ifs, err := net.Interfaces(); err == nil {
		for _, i := range ifs {
			if (i.Flags&net.FlagLoopback) == net.FlagLoopback ||
				(i.Flags&net.FlagUp) != net.FlagUp ||
				(i.Flags&net.FlagPointToPoint) == net.FlagPointToPoint ||
				len(i.HardwareAddr) != 6 {
				continue
			}
			addrs, err := i.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addrs {
				cidr := a.String()
				ip, _, err := net.ParseCIDR(cidr)
				if err != nil {
					continue
				}
				ipv4 := ip.To4()
				if ipv4 == nil {
					continue
				}
				sans = append(sans, ipv4.String())
			}
		}
	}
	return strings.Join(sans, ",")
}
