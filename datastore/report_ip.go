package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

type IPReportEnt struct {
	IP         string
	MAC        string
	Name       string
	NodeID     string
	Loc        string
	Vendor     string
	Count      int64
	Change     int64
	Score      float64
	ValidScore bool
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

func GetIPReport(id string) *IPReportEnt {
	if v, ok := ips.Load(id); ok {
		return v.(*IPReportEnt)
	}
	return nil
}

func AddIPReport(ip *IPReportEnt) {
	ips.Store(ip.IP, ip)
}

func ForEachIPReport(f func(*IPReportEnt) bool) {
	ips.Range(func(k, v interface{}) bool {
		ip := v.(*IPReportEnt)
		return f(ip)
	})
}

func findIPsFromMAC(mac string) []string {
	var ret = []string{}
	ips.Range(func(k, v interface{}) bool {
		ip := v.(*IPReportEnt)
		if ip.MAC == mac {
			ret = append(ret, ip.IP)
		}
		return true
	})
	return ret
}

// internal use
func loadIPs(r *bbolt.Bucket) {
	b := r.Bucket([]byte("ips"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var i IPReportEnt
			if err := json.Unmarshal(v, &i); err == nil {
				ips.Store(i.IP, &i)
			}
			return nil
		})
	}
}

func saveIPs(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("ips"))
	ips.Range(func(k, v interface{}) bool {
		i := v.(*IPReportEnt)
		if i.UpdateTime < last {
			return true
		}
		s, err := json.Marshal(i)
		if err != nil {
			log.Printf("Save Report err=%v", err)
			return true
		}
		err = r.Put([]byte(i.IP), s)
		if err != nil {
			log.Printf("Save Report err=%v", err)
		}
		return true
	})
}
