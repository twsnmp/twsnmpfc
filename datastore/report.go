package datastore

import (
	"encoding/json"
	"fmt"
	"log"

	"go.etcd.io/bbolt"
)

type ReportConfEnt struct {
	DenyCountries        []string
	DenyServices         []string
	AllowDNS             string
	AllowDHCP            string
	AllowMail            string
	AllowLDAP            string
	AllowLocalIP         string
	JapanOnly            bool
	DropFlowThTCPPacket  int
	RetentionTimeForSafe int
}

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

type UserEnt struct {
	ID           string // User ID + Server
	UserID       string
	Server       string
	ServerName   string
	ServerNodeID string
	Clients      map[string]int64
	Total        int
	Ok           int
	Score        float64
	ValidScore   bool
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

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

var ReportConf ReportConfEnt

func LoadReport() error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
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
		b = r.Bucket([]byte("users"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var u UserEnt
				if err := json.Unmarshal(v, &u); err == nil {
					users.Store(u.ID, &u)
				}
				return nil
			})
		}
		b = r.Bucket([]byte("servers"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var s ServerEnt
				if err := json.Unmarshal(v, &s); err == nil {
					servers.Store(s.ID, &s)
				}
				return nil
			})
		}
		b = r.Bucket([]byte("flows"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var f FlowEnt
				if err := json.Unmarshal(v, &f); err == nil {
					flows.Store(f.ID, &f)
				}
				return nil
			})
		}
		return nil
	})
}

func SaveReport(last int64) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
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
		r = b.Bucket([]byte("users"))
		users.Range(func(k, v interface{}) bool {
			u := v.(*UserEnt)
			if u.UpdateTime < last {
				return true
			}
			s, err := json.Marshal(u)
			if err != nil {
				log.Printf("Save Report err=%v", err)
				return true
			}
			err = r.Put([]byte(u.ID), s)
			if err != nil {
				log.Printf("Save Report err=%v", err)
			}
			return true
		})
		r = b.Bucket([]byte("servers"))
		servers.Range(func(k, v interface{}) bool {
			s := v.(*ServerEnt)
			if s.UpdateTime < last {
				return true
			}
			js, err := json.Marshal(s)
			if err != nil {
				log.Printf("Save Report err=%v", err)
				return true
			}
			err = r.Put([]byte(s.ID), js)
			if err != nil {
				log.Printf("Save Report err=%v", err)
			}
			return true
		})
		r = b.Bucket([]byte("flows"))
		flows.Range(func(k, v interface{}) bool {
			f := v.(*FlowEnt)
			if f.UpdateTime < last {
				return true
			}
			s, err := json.Marshal(f)
			if err != nil {
				log.Printf("Save Report err=%v", err)
				return true
			}
			err = r.Put([]byte(f.ID), s)
			if err != nil {
				log.Printf("Save Report err=%v", err)
			}
			return true
		})
		return nil
	})
}

func DeleteReport(report, id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte(report))
			if r != nil {
				r.Delete([]byte(id))
			}
		}
		return nil
	})
	if report == "devices" {
		devices.Delete(id)
	} else if report == "users" {
		users.Delete(id)
	} else if report == "servers" {
		servers.Delete(id)
	} else if report == "flows" {
		flows.Delete(id)
	}
	return nil
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

func DeleteDevice(id string) {
	devices.Delete(id)
}

func ForEachDevices(f func(*DeviceEnt) bool) {
	devices.Range(func(k, v interface{}) bool {
		d := v.(*DeviceEnt)
		return f(d)
	})
}

func GetUser(id string) *UserEnt {
	if v, ok := users.Load(id); ok {
		return v.(*UserEnt)
	}
	return nil
}

func AddUser(u *UserEnt) {
	users.Store(u.ID, u)
}

func ForEachUsers(f func(*UserEnt) bool) {
	users.Range(func(k, v interface{}) bool {
		u := v.(*UserEnt)
		return f(u)
	})
}

func DeleteUser(id string) {
	devices.Delete(id)
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

func DeleteFlow(id string) {
	flows.Delete(id)
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

func DeleteServer(id string) {
	servers.Delete(id)
}

func ClearAllReport() error {
	if db == nil {
		return ErrDBNotOpen
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			for _, r := range []string{"devices", "flows", "users", "servers"} {
				_ = b.DeleteBucket([]byte(r))
				_, _ = b.CreateBucketIfNotExists([]byte(r))
			}
		}
		return nil
	})
	devices.Range(func(k, v interface{}) bool {
		devices.Delete(k)
		return true
	})
	users.Range(func(k, v interface{}) bool {
		users.Delete(k)
		return true
	})
	flows.Range(func(k, v interface{}) bool {
		flows.Delete(k)
		return true
	})
	servers.Range(func(k, v interface{}) bool {
		servers.Delete(k)
		return true
	})
	return nil
}

// LaodReportConf :
func LaodReportConf() error {
	ReportConf.RetentionTimeForSafe = 24
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("report"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &ReportConf); err != nil {
			log.Printf("Unmarshal mapConf from DB error=%v", err)
			return err
		}
		return nil
	})
}

func SaveReportConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ReportConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("report"), s)
	})
}
