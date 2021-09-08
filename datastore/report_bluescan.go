package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

// type=Device,address=%s,name=%s,rssi=%d,addrType=%s,vendor=%s,md=%s
type RSSIEnt struct {
	Time  int64
	Value int
}
type BlueDeviceEnt struct {
	ID          string // Host + Address
	Host        string
	Address     string
	Name        string
	AddressType string
	RSSI        []RSSIEnt
	Vendor      string
	ExtData     string
	Count       int64
	FirstTime   int64
	LastTime    int64
}

func GetBlueDevice(id string) *BlueDeviceEnt {
	if v, ok := blueDevice.Load(id); ok {
		return v.(*BlueDeviceEnt)
	}
	return nil
}

func AddBlueDevice(e *BlueDeviceEnt) {
	blueDevice.Store(e.ID, e)
}

func ForEachBludeDevice(f func(*BlueDeviceEnt) bool) {
	blueDevice.Range(func(k, v interface{}) bool {
		e := v.(*BlueDeviceEnt)
		return f(e)
	})
}

// type=OMRONEnv,address=%s,name=%s,rssi=%d,seq=%d,temp=%.02f,hum=%.02f,lx=%d,press=%.02f,sound=%.02f,eTVOC=%d,eCO2=%d
type EnvDataEnt struct {
	Time               int64
	RSSI               int
	Temp               float64
	Humidity           float64
	Illuminance        float64
	BarometricPressure float64
	Sound              float64
	ETVOC              float64
	ECo2               float64
}

type EnvMonitorEnt struct {
	ID        string // Host + Address
	Host      string
	Name      string
	Address   string
	EnvData   []EnvDataEnt
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetEnvMonitor(id string) *EnvMonitorEnt {
	if v, ok := envMonitor.Load(id); ok {
		return v.(*EnvMonitorEnt)
	}
	return nil
}

func AddEnvMonitor(e *EnvMonitorEnt) {
	envMonitor.Store(e.ID, e)
}

func ForEachEnvMonitor(f func(*EnvMonitorEnt) bool) {
	envMonitor.Range(func(k, v interface{}) bool {
		e := v.(*EnvMonitorEnt)
		return f(e)
	})
}

func loadBlueDevice(r *bbolt.Bucket) {
	b := r.Bucket([]byte("blueDevice"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e BlueDeviceEnt
			if err := json.Unmarshal(v, &e); err == nil {
				blueDevice.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func saveBlueDevice(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("blueDevice"))
	blueDevice.Range(func(k, v interface{}) bool {
		e, ok := v.(*BlueDeviceEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("Save Report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("Save Report err=%v", err)
		}
		return true
	})
}

func loadEnvMonitor(r *bbolt.Bucket) {
	b := r.Bucket([]byte("envMonitor"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e EnvMonitorEnt
			if err := json.Unmarshal(v, &e); err == nil {
				envMonitor.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func saveEnvMonitor(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("envMonitor"))
	envMonitor.Range(func(k, v interface{}) bool {
		e, ok := v.(*EnvMonitorEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("Save Report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("Save Report err=%v", err)
		}
		return true
	})
}
