package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

type EtherTypeEnt struct {
	ID        string // ID Host:EtherType
	Host      string
	Type      string
	Name      string
	Count     int64
	FirstTime int64
	LastTime  int64
}

type DNSQEnt struct {
	ID           string // ID Host:Server:Type:Name
	Host         string
	Server       string
	Type         string
	Name         string
	Count        int64
	Change       int64
	ServerName   string
	ServerNodeID string
	ServerLoc    string
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type RADIUSFlowEnt struct {
	ID           string // ID Client:Server
	Client       string
	Server       string
	Count        int64
	Request      int64
	Challenge    int64
	Accept       int64
	Reject       int64
	ClientName   string
	ClientNodeID string
	ServerName   string
	ServerNodeID string
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type TLSFlowEnt struct {
	ID           string // ID Client:Server:Service
	Client       string
	Server       string
	Service      string
	Count        int64
	Version      string
	Cipher       string
	ClientName   string
	ClientNodeID string
	ClientLoc    string
	ServerName   string
	ServerNodeID string
	ServerLoc    string
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

func GetEtherType(id string) *EtherTypeEnt {
	if v, ok := etherType.Load(id); ok {
		return v.(*EtherTypeEnt)
	}
	return nil
}

func AddEtherType(s *EtherTypeEnt) {
	etherType.Store(s.ID, s)
}

func ForEachEtherType(f func(*EtherTypeEnt) bool) {
	etherType.Range(func(k, v interface{}) bool {
		s := v.(*EtherTypeEnt)
		return f(s)
	})
}

func GetDNSQ(id string) *DNSQEnt {
	if v, ok := dnsq.Load(id); ok {
		return v.(*DNSQEnt)
	}
	return nil
}

func AddDNSQ(s *DNSQEnt) {
	dnsq.Store(s.ID, s)
}

func ForEachDNSQ(f func(*DNSQEnt) bool) {
	dnsq.Range(func(k, v interface{}) bool {
		s := v.(*DNSQEnt)
		return f(s)
	})
}

func GetRADIUSFlow(id string) *RADIUSFlowEnt {
	if v, ok := radiusFlows.Load(id); ok {
		return v.(*RADIUSFlowEnt)
	}
	return nil
}

func AddRADIUSFlow(f *RADIUSFlowEnt) {
	radiusFlows.Store(f.ID, f)
}

func ForEachRADIUSFlows(f func(*RADIUSFlowEnt) bool) {
	radiusFlows.Range(func(k, v interface{}) bool {
		fl := v.(*RADIUSFlowEnt)
		return f(fl)
	})
}

func GetTLSFlow(id string) *TLSFlowEnt {
	if v, ok := tlsFlows.Load(id); ok {
		return v.(*TLSFlowEnt)
	}
	return nil
}

func AddTLSFlow(f *TLSFlowEnt) {
	tlsFlows.Store(f.ID, f)
}

func ForEachTLSFlows(f func(*TLSFlowEnt) bool) {
	tlsFlows.Range(func(k, v interface{}) bool {
		fl := v.(*TLSFlowEnt)
		return f(fl)
	})
}

func loadEther(r *bbolt.Bucket) {
	b := r.Bucket([]byte("ether"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e EtherTypeEnt
			if err := json.Unmarshal(v, &e); err == nil {
				etherType.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadDNS(r *bbolt.Bucket) {
	b := r.Bucket([]byte("dns"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e DNSQEnt
			if err := json.Unmarshal(v, &e); err == nil {
				dnsq.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadRADIUS(r *bbolt.Bucket) {
	b := r.Bucket([]byte("radius"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e RADIUSFlowEnt
			if err := json.Unmarshal(v, &e); err == nil {
				radiusFlows.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadTLS(r *bbolt.Bucket) {
	b := r.Bucket([]byte("tls"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e TLSFlowEnt
			if err := json.Unmarshal(v, &e); err == nil {
				tlsFlows.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func saveEther(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("ether"))
	etherType.Range(func(k, v interface{}) bool {
		e := v.(*EtherTypeEnt)
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("save ether report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save ether report err=%v", err)
		}
		return true
	})
}

func saveDNS(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("dns"))
	dnsq.Range(func(k, v interface{}) bool {
		e := v.(*DNSQEnt)
		if e.UpdateTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("save dns report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save dns report err=%v", err)
		}
		return true
	})
}

func saveRADIUS(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("radius"))
	radiusFlows.Range(func(k, v interface{}) bool {
		e := v.(*RADIUSFlowEnt)
		if e.UpdateTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("save radius report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save radius report err=%v", err)
		}
		return true
	})
}

func saveTLS(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("tls"))
	tlsFlows.Range(func(k, v interface{}) bool {
		e := v.(*TLSFlowEnt)
		if e.UpdateTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("save tls report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save tls report err=%v", err)
		}
		return true
	})
}
