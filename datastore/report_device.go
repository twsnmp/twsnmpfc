package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

type DeviceEnt struct {
	ID         string // MAC Addr
	Name       string
	IP         string
	NodeID     string
	Vendor     string
	Score      float64
	ValidScore bool
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

func GetDevice(id string) *DeviceEnt {
	if v, ok := devices.Load(id); ok {
		return v.(*DeviceEnt)
	}
	return nil
}

func AddDevice(d *DeviceEnt) {
	devices.Store(d.ID, d)
}

func ForEachDevices(f func(*DeviceEnt) bool) {
	devices.Range(func(k, v interface{}) bool {
		d := v.(*DeviceEnt)
		return f(d)
	})
}

// internal use

func loadDevices(r *bbolt.Bucket) {
	b := r.Bucket([]byte("devices"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var d DeviceEnt
			if err := json.Unmarshal(v, &d); err == nil {
				devices.Store(d.ID, &d)
			}
			return nil
		})
	}
}

func saveDevices(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("devices"))
	devices.Range(func(k, v interface{}) bool {
		d := v.(*DeviceEnt)
		if d.UpdateTime < last {
			return true
		}
		s, err := json.Marshal(d)
		if err != nil {
			log.Printf("Save Report err=%v", err)
			return true
		}
		err = r.Put([]byte(d.ID), s)
		if err != nil {
			log.Printf("Save Report err=%v", err)
		}
		return true
	})
}
