package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

type CertEnt struct {
	ID           string // Target:PORT
	Target       string
	Port         uint16
	Subject      string
	Issuer       string
	SerialNumber string
	Verify       bool
	NotAfter     int64
	NotBefore    int64
	Error        string
	Score        float64
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

func GetCert(id string) *CertEnt {
	if v, ok := certs.Load(id); ok {
		if p, ok := v.(*CertEnt); ok {
			return p
		}
	}
	return nil
}

func AddCert(c *CertEnt) {
	certs.Store(c.ID, c)
}

func ForEachCerts(f func(*CertEnt) bool) {
	certs.Range(func(k, v interface{}) bool {
		if c, ok := v.(*CertEnt); ok {
			return f(c)
		}
		return true
	})
}

// internal use
func loadCert(r *bbolt.Bucket) {
	b := r.Bucket([]byte("cert"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e CertEnt
			if err := json.Unmarshal(v, &e); err == nil {
				certs.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func saveCert(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("cert"))
	certs.Range(func(k, v interface{}) bool {
		e := v.(*CertEnt)
		if e.UpdateTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("save cert report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save cert report err=%v", err)
		}
		return true
	})
}
