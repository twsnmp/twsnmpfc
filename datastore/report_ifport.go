package datastore

import (
	"encoding/json"
	"log"
	"sync"

	"go.etcd.io/bbolt"
)

var ifPortTable sync.Map

type IfPortEnt struct {
	IfIndex         int
	MAC             string
	Name            string
	Descr           string
	Type            string
	Mtu             uint64
	Speed           uint64
	AdminStatus     int
	OperStatus      int
	InOctets        uint64
	InBPS           uint64
	OutOctets       uint64
	OutBPS          uint64
	FirstCheckTime  int64
	LastCheckTime   int64
	LastChangedTime int64
	Changed         int
}

func GetIfPortTable(id string) *[]IfPortEnt {
	if v, ok := ifPortTable.Load(id); ok {
		return v.(*[]IfPortEnt)
	}
	return nil
}

func UpdateIfPortTable(id string, l *[]IfPortEnt) {
	ifPortTable.Store(id, l)
}

func ForEachIfPortTable(f func(string, *[]IfPortEnt) bool) {
	ifPortTable.Range(func(k, v interface{}) bool {
		if e, ok := v.(*[]IfPortEnt); ok {
			return f(k.(string), e)
		}
		return true
	})
}

func loadIfPortTable(r *bbolt.Bucket) {
	b := r.Bucket([]byte("ifPortTable"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var l []IfPortEnt
			if err := json.Unmarshal(v, &l); err == nil {
				ifPortTable.Store(string(k), &l)
			}
			return nil
		})
	}
}

func saveIfPortTable(r *bbolt.Bucket) {
	b := r.Bucket([]byte("ifPortTable"))
	ifPortTable.Range(func(k, v interface{}) bool {
		if e, ok := v.(*[]IfPortEnt); ok {
			s, err := json.Marshal(e)
			if err != nil {
				log.Printf("save ifPortTable report err=%v", err)
				return true
			}
			err = b.Put([]byte(k.(string)), s)
			if err != nil {
				log.Printf("save ifPortTable report err=%v", err)
			}
		}
		return true
	})
}
