package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

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
		b = r.Bucket([]byte("ips"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var i IPReportEnt
				if err := json.Unmarshal(v, &i); err == nil {
					ips.Store(i.IP, &i)
				}
				return nil
			})
		}
		b = r.Bucket([]byte("ether"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var e EtherTypeEnt
				if err := json.Unmarshal(v, &e); err == nil {
					etherType.Store(e.ID, &e)
				}
				return nil
			})
		}
		b = r.Bucket([]byte("dns"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var e DNSQEnt
				if err := json.Unmarshal(v, &e); err == nil {
					dnsq.Store(e.ID, &e)
				}
				return nil
			})
		}
		b = r.Bucket([]byte("radius"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var e RADIUSFlowEnt
				if err := json.Unmarshal(v, &e); err == nil {
					radiusFlows.Store(e.ID, &e)
				}
				return nil
			})
		}
		b = r.Bucket([]byte("tls"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var e TLSFlowEnt
				if err := json.Unmarshal(v, &e); err == nil {
					tlsFlows.Store(e.ID, &e)
				}
				return nil
			})
		}
		b = r.Bucket([]byte("cert"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var e CertEnt
				if err := json.Unmarshal(v, &e); err == nil {
					certs.Store(e.ID, &e)
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
		r = b.Bucket([]byte("ips"))
		ips.Range(func(k, v interface{}) bool {
			i := v.(*IPReportEnt)
			if i.UpdateTime < last {
				return true
			}
			s, err := json.Marshal(i)
			if err != nil {
				log.Printf("Save Report err=%v", err)
				return true
			}
			err = r.Put([]byte(i.IP), s)
			if err != nil {
				log.Printf("Save Report err=%v", err)
			}
			return true
		})
		r = b.Bucket([]byte("ether"))
		etherType.Range(func(k, v interface{}) bool {
			e := v.(*EtherTypeEnt)
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
		r = b.Bucket([]byte("dns"))
		dnsq.Range(func(k, v interface{}) bool {
			e := v.(*DNSQEnt)
			if e.UpdateTime < last {
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
		r = b.Bucket([]byte("radius"))
		radiusFlows.Range(func(k, v interface{}) bool {
			e := v.(*RADIUSFlowEnt)
			if e.UpdateTime < last {
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
		r = b.Bucket([]byte("tls"))
		tlsFlows.Range(func(k, v interface{}) bool {
			e := v.(*TLSFlowEnt)
			if e.UpdateTime < last {
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
		r = b.Bucket([]byte("cert"))
		certs.Range(func(k, v interface{}) bool {
			e := v.(*CertEnt)
			if e.UpdateTime < last {
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
	switch report {
	case "devices":
		devices.Delete(id)
	case "users":
		users.Delete(id)
	case "servers":
		servers.Delete(id)
	case "flows":
		flows.Delete(id)
	case "ips":
		ips.Delete(id)
	case "ether":
		etherType.Delete(id)
	case "dns":
		dnsq.Delete(id)
	case "radius":
		radiusFlows.Delete(id)
	case "tls":
		tlsFlows.Delete(id)
	case "cert":
		certs.Delete(id)
	}
	return nil
}

func ClearReport(r string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			_ = b.DeleteBucket([]byte(r))
			_, _ = b.CreateBucketIfNotExists([]byte(r))
		}
		return nil
	})
	if r == "devices" {
		devices.Range(func(k, v interface{}) bool {
			devices.Delete(k)
			return true
		})
		return nil
	}
	if r == "ips" {
		ips.Range(func(k, v interface{}) bool {
			ips.Delete(k)
			return true
		})
		return nil
	}
	if r == "users" {
		users.Range(func(k, v interface{}) bool {
			users.Delete(k)
			return true
		})
		return nil
	}
	if r == "flows" {
		flows.Range(func(k, v interface{}) bool {
			flows.Delete(k)
			return true
		})
		return nil
	}
	if r == "servers" {
		servers.Range(func(k, v interface{}) bool {
			servers.Delete(k)
			return true
		})
	}
	if r == "ether" {
		etherType.Range(func(k, v interface{}) bool {
			etherType.Delete(k)
			return true
		})
	}
	if r == "dns" {
		dnsq.Range(func(k, v interface{}) bool {
			dnsq.Delete(k)
			return true
		})
	}
	if r == "radius" {
		radiusFlows.Range(func(k, v interface{}) bool {
			radiusFlows.Delete(k)
			return true
		})
	}
	if r == "tls" {
		tlsFlows.Range(func(k, v interface{}) bool {
			tlsFlows.Delete(k)
			return true
		})
	}
	return nil
}
