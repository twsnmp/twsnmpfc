package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

type ServerEnt struct {
	ID           string //  ID Server
	Server       string
	Services     map[string]int64
	Count        int64
	Bytes        int64
	ServerName   string
	ServerNodeID string
	Loc          string
	Score        float64
	ValidScore   bool
	Penalty      int64
	TLSInfo      string
	NTPInfo      string
	DHCPInfo     string
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

type FlowEnt struct {
	ID           string // ID Client:Server
	Client       string
	Server       string
	Services     map[string]int64
	Count        int64
	Bytes        int64
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

func GetFlow(id string) *FlowEnt {
	if v, ok := flows.Load(id); ok {
		return v.(*FlowEnt)
	}
	return nil
}

func AddFlow(f *FlowEnt) {
	flows.Store(f.ID, f)
}

func ForEachFlows(f func(*FlowEnt) bool) {
	flows.Range(func(k, v interface{}) bool {
		fl := v.(*FlowEnt)
		return f(fl)
	})
}

func GetServer(id string) *ServerEnt {
	if v, ok := servers.Load(id); ok {
		return v.(*ServerEnt)
	}
	return nil
}

func AddServer(s *ServerEnt) {
	servers.Store(s.ID, s)
}

func ForEachServers(f func(*ServerEnt) bool) {
	servers.Range(func(k, v interface{}) bool {
		s := v.(*ServerEnt)
		return f(s)
	})
}

// internal use
func loadServers(r *bbolt.Bucket) {
	b := r.Bucket([]byte("servers"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var s ServerEnt
			if err := json.Unmarshal(v, &s); err == nil {
				servers.Store(s.ID, &s)
			}
			return nil
		})
	}
}

func loadFlows(r *bbolt.Bucket) {
	b := r.Bucket([]byte("flows"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var f FlowEnt
			if err := json.Unmarshal(v, &f); err == nil {
				flows.Store(f.ID, &f)
			}
			return nil
		})
	}
}

func saveServers(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("servers"))
	servers.Range(func(k, v interface{}) bool {
		s := v.(*ServerEnt)
		if s.UpdateTime < last {
			return true
		}
		js, err := json.Marshal(s)
		if err != nil {
			log.Printf("save server report err=%v", err)
			return true
		}
		err = r.Put([]byte(s.ID), js)
		if err != nil {
			log.Printf("save server report err=%v", err)
		}
		return true
	})
}

func saveFlows(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("flows"))
	flows.Range(func(k, v interface{}) bool {
		f := v.(*FlowEnt)
		if f.UpdateTime < last {
			return true
		}
		s, err := json.Marshal(f)
		if err != nil {
			log.Printf("save flow report err=%v", err)
			return true
		}
		err = r.Put([]byte(f.ID), s)
		if err != nil {
			log.Printf("save flow report err=%v", err)
		}
		return true
	})
}
