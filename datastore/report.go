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
	NodeID     string
	Vendor     string
	Services   map[string]int64
	Score      float64
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
	Penalty      int64
	FirstTime    int64
	LastTime     int64
	UpdateTime   int64
}

// AllowRuleEnt : 特定のサービスは特定のサーバーに限定するルール
type AllowRuleEnt struct {
	Service string // Service
	Servers map[string]bool
}

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
					devices[d.ID] = &d
				}
				return nil
			})
		}
		b = r.Bucket([]byte("users"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var u UserEnt
				if err := json.Unmarshal(v, &u); err == nil {
					users[u.ID] = &u
				}
				return nil
			})
		}
		b = r.Bucket([]byte("servers"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var s ServerEnt
				if err := json.Unmarshal(v, &s); err == nil {
					servers[s.ID] = &s
				}
				return nil
			})
		}
		b = r.Bucket([]byte("flows"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var f FlowEnt
				if err := json.Unmarshal(v, &f); err == nil {
					flows[f.ID] = &f
				}
				return nil
			})
		}
		b = r.Bucket([]byte("dennys"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var en bool
				if err := json.Unmarshal(v, &en); err == nil {
					dennyRules[string(k)] = en
				}
				return nil
			})
		}
		b = r.Bucket([]byte("allows"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var as AllowRuleEnt
				if err := json.Unmarshal(v, &as); err == nil {
					allowRules[as.Service] = &as
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
		for _, d := range devices {
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
		for _, u := range users {
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
		for _, s := range servers {
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
		for _, f := range flows {
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

func DeleteReport(report, id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	db.Update(func(tx *bbolt.Tx) error {
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
		delete(devices, id)
	} else if report == "users" {
		delete(users, id)
	} else if report == "servers" {
		delete(servers, id)
	} else if report == "flows" {
		delete(flows, id)
	}
	return nil
}

func GetDevice(id string) *DeviceEnt {
	return devices[id]
}

func AddDevice(d *DeviceEnt) {
	devices[d.ID] = d
}

func DeleteDevice(id string) {
	delete(devices, id)
}

func ForEachDevices(f func(*DeviceEnt) bool) {
	for _, d := range devices {
		if !f(d) {
			return
		}
	}
}

func GetUser(id string) *UserEnt {
	return users[id]
}

func AddUser(u *UserEnt) {
	users[u.ID] = u
}

func ForEachUsers(f func(*UserEnt) bool) {
	for _, u := range users {
		if !f(u) {
			return
		}
	}
}

func DeleteUser(id string) {
	delete(devices, id)
}

func GetDennyRule(id string) bool {
	return dennyRules[id]
}

func GetAllowRule(id string) *AllowRuleEnt {
	return allowRules[id]
}

func GetFlow(id string) *FlowEnt {
	return flows[id]
}

func AddFlow(f *FlowEnt) {
	flows[f.ID] = f
}

func ForEachFlows(f func(*FlowEnt) bool) {
	for _, fl := range flows {
		if !f(fl) {
			return
		}
	}
}

func DeleteFlow(id string) {
	delete(flows, id)
}

func GetServer(id string) *ServerEnt {
	return servers[id]
}

func AddServer(s *ServerEnt) {
	servers[s.ID] = s
}

func LenServers() int {
	return len(servers)
}

func ForEachServers(f func(*ServerEnt) bool) {
	for _, s := range servers {
		if !f(s) {
			return
		}
	}
}

func DeleteServer(id string) {
	delete(servers, id)
}

func AddAllowRule(service, server string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	as, ok := allowRules[service]
	if ok {
		as.Servers[server] = true
	} else {
		as = &AllowRuleEnt{
			Service: service,
			Servers: map[string]bool{server: true},
		}
		allowRules[service] = as
	}
	js, err := json.Marshal(as)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
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

func DeleteAllowRule(id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	a := strings.Split(id, ":")
	if len(a) != 2 {
		return fmt.Errorf("deleteAllowRule bad id %s", id)
	}
	server := a[0]
	service := a[1]
	as, ok := allowRules[service]
	if !ok {
		return nil
	}
	delete(as.Servers, server)
	js := []byte{}
	if len(as.Servers) > 0 {
		js, _ = json.Marshal(as)
	}
	return db.Update(func(tx *bbolt.Tx) error {
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

func AddDennyRule(id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	dennyRules[id] = true
	js, err := json.Marshal(dennyRules[id])
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
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

func deleteDennyRule(id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	_, ok := dennyRules[id]
	if !ok {
		return nil
	}
	delete(dennyRules, id)
	return db.Update(func(tx *bbolt.Tx) error {
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
	devices = make(map[string]*DeviceEnt)
	users = make(map[string]*UserEnt)
	flows = make(map[string]*FlowEnt)
	servers = make(map[string]*ServerEnt)
	return nil
}
