package datastore

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.etcd.io/bbolt"
)

type DeviceEnt struct {
	ID         string // MAC Addr
	Name       string
	IP         string
	Vendor     string
	Services   map[string]int64
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type UserEnt struct {
	ID         string // User ID + Server
	UserID     string
	Server     string
	ServerName string
	Clients    map[string]int64
	Total      int
	Ok         int
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type ServerEnt struct {
	ID         string //  ID Server
	Server     string
	Services   map[string]int64
	Count      int64
	Bytes      int64
	ServerName string
	Loc        string
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type FlowEnt struct {
	ID         string // ID Client:Server
	Client     string
	Server     string
	Services   map[string]int64
	Count      int64
	Bytes      int64
	ClientName string
	ClientLoc  string
	ServerName string
	ServerLoc  string
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

// AllowRuleEnt : 特定のサービスは特定のサーバーに限定するルール
type AllowRuleEnt struct {
	Service string // Service
	Servers map[string]bool
}

func (ds *DataStore) loadReport() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	return ds.db.View(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		b := r.Bucket([]byte("devices"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var d DeviceEnt
				if err := json.Unmarshal(v, &d); err == nil {
					ds.devices[d.ID] = &d
				}
				return nil
			})
		}
		b = r.Bucket([]byte("users"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var u UserEnt
				if err := json.Unmarshal(v, &u); err == nil {
					ds.users[u.ID] = &u
				}
				return nil
			})
		}
		b = r.Bucket([]byte("servers"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var s ServerEnt
				if err := json.Unmarshal(v, &s); err == nil {
					ds.servers[s.ID] = &s
				}
				return nil
			})
		}
		b = r.Bucket([]byte("flows"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var f FlowEnt
				if err := json.Unmarshal(v, &f); err == nil {
					ds.flows[f.ID] = &f
				}
				return nil
			})
		}
		b = r.Bucket([]byte("dennys"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var en bool
				if err := json.Unmarshal(v, &en); err == nil {
					ds.dennyRules[string(k)] = en
				}
				return nil
			})
		}
		b = r.Bucket([]byte("allows"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var as AllowRuleEnt
				if err := json.Unmarshal(v, &as); err == nil {
					ds.allowRules[as.Service] = &as
				}
				return nil
			})
		}
		return nil
	})
}

func (ds *DataStore) SaveReport(last int64) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	return ds.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		r := b.Bucket([]byte("devices"))
		for _, d := range ds.devices {
			if d.UpdateTime < last {
				continue
			}
			s, err := json.Marshal(d)
			if err != nil {
				return err
			}
			err = r.Put([]byte(d.ID), s)
			if err != nil {
				return err
			}
		}
		r = b.Bucket([]byte("users"))
		for _, u := range ds.users {
			if u.UpdateTime < last {
				continue
			}
			s, err := json.Marshal(u)
			if err != nil {
				return err
			}
			err = r.Put([]byte(u.ID), s)
			if err != nil {
				return err
			}
		}
		r = b.Bucket([]byte("servers"))
		for _, s := range ds.servers {
			if s.UpdateTime < last {
				continue
			}
			js, err := json.Marshal(s)
			if err != nil {
				return err
			}
			err = r.Put([]byte(s.ID), js)
			if err != nil {
				return err
			}
		}
		r = b.Bucket([]byte("flows"))
		for _, f := range ds.flows {
			if f.UpdateTime < last {
				continue
			}
			s, err := json.Marshal(f)
			if err != nil {
				return err
			}
			err = r.Put([]byte(f.ID), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ds *DataStore) DeleteReport(report, id string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte(report))
			if r != nil {
				_ = r.Delete([]byte(id))
			}
		}
		return nil
	})
	if report == "devices" {
		delete(ds.devices, id)
	} else if report == "users" {
		delete(ds.users, id)
	} else if report == "servers" {
		delete(ds.servers, id)
	} else if report == "flows" {
		delete(ds.flows, id)
	}
	return nil
}

func (ds *DataStore) GetDevice(id string) *DeviceEnt {
	return ds.devices[id]
}

func (ds *DataStore) AddDevice(d *DeviceEnt) {
	ds.devices[d.ID] = d
}

func (ds *DataStore) ForEachDevices(f func(*DeviceEnt) bool) {
	for _, d := range ds.devices {
		if !f(d) {
			return
		}
	}
}

func (ds *DataStore) GetUser(id string) *UserEnt {
	return ds.users[id]
}

func (ds *DataStore) AddUser(u *UserEnt) {
	ds.users[u.ID] = u
}

func (ds *DataStore) ForEachUsers(f func(*UserEnt) bool) {
	for _, u := range ds.users {
		if !f(u) {
			return
		}
	}
}

func (ds *DataStore) GetDennyRule(id string) bool {
	return ds.dennyRules[id]
}

func (ds *DataStore) GetAllowRule(id string) *AllowRuleEnt {
	return ds.allowRules[id]
}

func (ds *DataStore) GetFlow(id string) *FlowEnt {
	return ds.flows[id]
}

func (ds *DataStore) AddFlow(f *FlowEnt) {
	ds.flows[f.ID] = f
}

func (ds *DataStore) ForEachFlows(f func(*FlowEnt) bool) {
	for _, fl := range ds.flows {
		if !f(fl) {
			return
		}
	}
}

func (ds *DataStore) GetServer(id string) *ServerEnt {
	return ds.servers[id]
}

func (ds *DataStore) AddServer(s *ServerEnt) {
	ds.servers[s.ID] = s
}

func (ds *DataStore) LenServers() int {
	return len(ds.servers)
}

func (ds *DataStore) ForEachServers(f func(*ServerEnt) bool) {
	for _, s := range ds.servers {
		if !f(s) {
			return
		}
	}
}

func (ds *DataStore) AddAllowRule(service, server string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	as, ok := ds.allowRules[service]
	if ok {
		as.Servers[server] = true
	} else {
		as = &AllowRuleEnt{
			Service: service,
			Servers: map[string]bool{server: true},
		}
		ds.allowRules[service] = as
	}
	js, err := json.Marshal(as)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte("allows"))
			if r != nil {
				_ = r.Put([]byte(service), js)
			}
		}
		return nil
	})
}

func (ds *DataStore) DeleteAllowRule(id string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	a := strings.Split(id, ":")
	if len(a) != 2 {
		return fmt.Errorf("deleteAllowRule bad id %s", id)
	}
	server := a[0]
	service := a[1]
	as, ok := ds.allowRules[service]
	if !ok {
		return nil
	}
	delete(as.Servers, server)
	js := []byte{}
	if len(as.Servers) > 0 {
		js, _ = json.Marshal(as)
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte("allows"))
			if r != nil {
				if len(js) < 1 {
					_ = r.Delete([]byte(service))
				} else {
					_ = r.Put([]byte(service), js)
				}
			}
		}
		return nil
	})
}

func (ds *DataStore) AddDennyRule(id string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	ds.dennyRules[id] = true
	js, err := json.Marshal(ds.dennyRules[id])
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte("dennys"))
			if r != nil {
				_ = r.Put([]byte(id), js)
			}
		}
		return nil
	})
}

func (ds *DataStore) deleteDennyRule(id string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	_, ok := ds.dennyRules[id]
	if !ok {
		return nil
	}
	delete(ds.dennyRules, id)
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte("dennys"))
			if r != nil {
				_ = r.Delete([]byte(id))
			}
		}
		return nil
	})
}

func (ds *DataStore) ClearAllReport() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			for _, r := range []string{"devices", "flows", "users", "servers"} {
				_ = b.DeleteBucket([]byte(r))
				_, _ = b.CreateBucketIfNotExists([]byte(r))
			}
		}
		return nil
	})
	ds.devices = make(map[string]*DeviceEnt)
	ds.users = make(map[string]*UserEnt)
	ds.flows = make(map[string]*FlowEnt)
	ds.servers = make(map[string]*ServerEnt)
	return nil
}
