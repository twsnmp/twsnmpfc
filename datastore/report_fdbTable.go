package datastore

import (
	"encoding/json"
	"log"
	"sync"

	"go.etcd.io/bbolt"
)

var fdbTable sync.Map

type FDBTableEnt struct {
	MAC             string
	VLanID          int
	Port            int
	IfIndex         int
	FirstCheckTime  int64
	LastCheckTime   int64
	LastChangedTime int64
	Changed         int
}

func GetFDBTable(id string) *[]FDBTableEnt {
	if v, ok := fdbTable.Load(id); ok {
		return v.(*[]FDBTableEnt)
	}
	return nil
}

func UpdateFDBTable(id string, l *[]FDBTableEnt) {
	fdbTable.Store(id, l)
}

func ForEachFDBTable(f func(string, *[]FDBTableEnt) bool) {
	fdbTable.Range(func(k, v interface{}) bool {
		if e, ok := v.(*[]FDBTableEnt); ok {
			return f(k.(string), e)
		}
		return true
	})
}

func loadFDBTable(r *bbolt.Bucket) {
	b := r.Bucket([]byte("fdbTable"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var l []FDBTableEnt
			if err := json.Unmarshal(v, &l); err == nil {
				fdbTable.Store(string(k), &l)
			}
			return nil
		})
	}
}

func saveFDBTable(r *bbolt.Bucket) {
	b := r.Bucket([]byte("fdbTable"))
	fdbTable.Range(func(k, v interface{}) bool {
		if e, ok := v.(*[]FDBTableEnt); ok {
			s, err := json.Marshal(e)
			if err != nil {
				log.Printf("save fdbTable report err=%v", err)
				return true
			}
			err = b.Put([]byte(k.(string)), s)
			if err != nil {
				log.Printf("save fdbTable report err=%v", err)
			}
		}
		return true
	})
}
