package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

// RSSIEnt represents a Bluetooth device RSSI entry.
type RSSIEnt struct {
	Time  int64
	Value int
}

// BlueDeviceEnt represents a Bluetooth device entry.
type BlueDeviceEnt struct {
	ID          string // Host + Address
	Host        string
	Address     string
	Name        string
	AddressType string
	RSSI        []RSSIEnt
	Info        string
	Vendor      string
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

func SetBlueDeviceName(id, name string) bool {
	if v, ok := blueDevice.Load(id); ok {
		if e, ok := v.(*BlueDeviceEnt); ok {
			e.Name = name
			return true
		}
	}
	return false
}

func AddBlueDevice(e *BlueDeviceEnt) {
	blueDevice.Store(e.ID, e)
}

func ForEachBlueDevice(f func(*BlueDeviceEnt) bool) {
	blueDevice.Range(func(k, v interface{}) bool {
		e := v.(*BlueDeviceEnt)
		return f(e)
	})
}

// EnvDataEnt represents environmental data from OMRON environment sensors.
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
	Battery            int
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

func SetEnvMonitorName(id, name string) bool {
	if v, ok := envMonitor.Load(id); ok {
		if e, ok := v.(*EnvMonitorEnt); ok {
			e.Name = name
			return true
		}
	}
	return false
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
			log.Printf("save bluetooth report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save bluetooth report err=%v", err)
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
			log.Printf("save env monitor report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save env monitor report err=%v", err)
		}
		return true
	})
}

type PowerMonitorDataEnt struct {
	Time   int64
	Switch bool
	Over   bool
	Load   float64
	RSSI   int
}

type PowerMonitorEnt struct {
	ID        string // Host + Address
	Host      string
	Name      string
	Address   string
	Data      []PowerMonitorDataEnt
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetPowerMonitor(id string) *PowerMonitorEnt {
	if v, ok := powerMonitor.Load(id); ok {
		return v.(*PowerMonitorEnt)
	}
	return nil
}

func SetPowerMonitorName(id, name string) bool {
	if v, ok := powerMonitor.Load(id); ok {
		if e, ok := v.(*PowerMonitorEnt); ok {
			e.Name = name
			return true
		}
	}
	return false
}

func AddPowerMonitor(e *PowerMonitorEnt) {
	powerMonitor.Store(e.ID, e)
}

func ForEachPowerMonitor(f func(*PowerMonitorEnt) bool) {
	powerMonitor.Range(func(k, v interface{}) bool {
		e := v.(*PowerMonitorEnt)
		return f(e)
	})
}

func loadPowerMonitor(r *bbolt.Bucket) {
	b := r.Bucket([]byte("powerMonitor"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e PowerMonitorEnt
			if err := json.Unmarshal(v, &e); err == nil {
				powerMonitor.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func savePowerMonitor(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("powerMonior"))
	if r == nil {
		return
	}
	powerMonitor.Range(func(k, v interface{}) bool {
		e, ok := v.(*PowerMonitorEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("save powerMonitor report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save powerMonitor report err=%v", err)
		}
		return true
	})
}

type MotionSensorDataEnt struct {
	Time         int64
	Moving       bool
	Light        bool
	Battery      int64
	LastMove     int64
	LastMoveDiff int64
	Event        string
	RSSI         int
}

type MotionSensorEnt struct {
	ID        string // Host + Address
	Host      string
	Name      string
	Address   string
	Data      []MotionSensorDataEnt
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetMotionSensor(id string) *MotionSensorEnt {
	if v, ok := motionSensor.Load(id); ok {
		return v.(*MotionSensorEnt)
	}
	return nil
}

func SetMotionSensorName(id, name string) bool {
	if v, ok := motionSensor.Load(id); ok {
		if e, ok := v.(*MotionSensorEnt); ok {
			e.Name = name
			return true
		}
	}
	return false
}

func AddMotionSensor(e *MotionSensorEnt) {
	motionSensor.Store(e.ID, e)
}

func ForEachMotionSensor(f func(*MotionSensorEnt) bool) {
	motionSensor.Range(func(k, v interface{}) bool {
		e := v.(*MotionSensorEnt)
		return f(e)
	})
}

func loadMotionSensor(r *bbolt.Bucket) {
	b := r.Bucket([]byte("motionSensor"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e MotionSensorEnt
			if err := json.Unmarshal(v, &e); err == nil {
				motionSensor.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func saveMotionSensor(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("motionSensor"))
	if r == nil {
		return
	}
	motionSensor.Range(func(k, v interface{}) bool {
		e, ok := v.(*MotionSensorEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("save motionSensor report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save motionSensor report err=%v", err)
		}
		return true
	})
}
