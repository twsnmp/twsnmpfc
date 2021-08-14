package datastore

import (
	"encoding/json"
	"log"
	"sync"

	"go.etcd.io/bbolt"
)

func LoadReport() error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		loadDevices(r)
		loadUsers(r)
		loadServers(r)
		loadFlows(r)
		loadIPs(r)
		loadEther(r)
		loadDNS(r)
		loadRADIUS(r)
		loadTLS(r)
		loadCert(r)
		loadSensor(r)
		loadWinEventID(r)
		loadWinLogon(r)
		loadWinAccount(r)
		loadWinKerberosTGT(r)
		loadWinKerberosST(r)
		loadWinPrivilege(r)
		loadWinProcess(r)
		loadWinTask(r)
		return nil
	})
}

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

func loadUsers(r *bbolt.Bucket) {
	b := r.Bucket([]byte("users"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var u UserEnt
			if err := json.Unmarshal(v, &u); err == nil {
				users.Store(u.ID, &u)
			}
			return nil
		})
	}
}

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

func loadIPs(r *bbolt.Bucket) {
	b := r.Bucket([]byte("ips"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var i IPReportEnt
			if err := json.Unmarshal(v, &i); err == nil {
				ips.Store(i.IP, &i)
			}
			return nil
		})
	}
}

func loadEther(r *bbolt.Bucket) {
	b := r.Bucket([]byte("ether"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e EtherTypeEnt
			if err := json.Unmarshal(v, &e); err == nil {
				etherType.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadDNS(r *bbolt.Bucket) {
	b := r.Bucket([]byte("dns"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e DNSQEnt
			if err := json.Unmarshal(v, &e); err == nil {
				dnsq.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadRADIUS(r *bbolt.Bucket) {
	b := r.Bucket([]byte("radius"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e RADIUSFlowEnt
			if err := json.Unmarshal(v, &e); err == nil {
				radiusFlows.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadTLS(r *bbolt.Bucket) {
	b := r.Bucket([]byte("tls"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e TLSFlowEnt
			if err := json.Unmarshal(v, &e); err == nil {
				tlsFlows.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadCert(r *bbolt.Bucket) {
	b := r.Bucket([]byte("cert"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e CertEnt
			if err := json.Unmarshal(v, &e); err == nil {
				certs.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadSensor(r *bbolt.Bucket) {
	b := r.Bucket([]byte("sensor"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e SensorEnt
			if err := json.Unmarshal(v, &e); err == nil {
				sensors.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadWinEventID(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winEventID"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinEventIDEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winEventID.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadWinLogon(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winLogon"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinLogonEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winLogon.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadWinAccount(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winAccount"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinAccountEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winAccount.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadWinKerberosTGT(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winKerberosTGT"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinKerberosTGTEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winKerberosTGT.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadWinKerberosST(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winKerberosST"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinKerberosSTEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winKerberosST.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadWinPrivilege(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winPrivilege"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinPrivilegeEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winPrivilege.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadWinProcess(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winProcess"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinProcessEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winProcess.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func loadWinTask(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winTask"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinTaskEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winTask.Store(e.ID, &e)
			}
			return nil
		})
	}
}

func SaveReport(last int64) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		saveDevices(b, last)
		saveUsers(b, last)
		saveServers(b, last)
		saveFlows(b, last)
		saveIPs(b, last)
		saveEther(b, last)
		saveDNS(b, last)
		saveRADIUS(b, last)
		saveTLS(b, last)
		saveCert(b, last)
		saveSensor(b, last)
		saveWinEventID(b, last)
		saveWinLogon(b, last)
		saveWinAccount(b, last)
		saveWinKerberosTGT(b, last)
		saveWinKerberosST(b, last)
		saveWinPrivilege(b, last)
		saveWinProcess(b, last)
		saveWinTask(b, last)
		return nil
	})
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

func saveUsers(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("users"))
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
			log.Printf("Save Report err=%v", err)
			return true
		}
		err = r.Put([]byte(s.ID), js)
		if err != nil {
			log.Printf("Save Report err=%v", err)
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
			log.Printf("Save Report err=%v", err)
			return true
		}
		err = r.Put([]byte(f.ID), s)
		if err != nil {
			log.Printf("Save Report err=%v", err)
		}
		return true
	})
}

func saveIPs(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("ips"))
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
}

func saveEther(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("ether"))
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
}

func saveDNS(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("dns"))
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
}

func saveRADIUS(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("radius"))
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
}

func saveTLS(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("tls"))
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
}

func saveCert(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("cert"))
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
}

func saveSensor(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("sensor"))
	sensors.Range(func(k, v interface{}) bool {
		e, ok := v.(*SensorEnt)
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

func saveWinEventID(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("WinEventID"))
	winEventID.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinEventIDEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("saveWinEventID err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("saveWinEventID err=%v", err)
		}
		return true
	})
}

func saveWinLogon(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("WinLogon"))
	winLogon.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinLogonEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("saveWinLogon err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("saveWinLogon err=%v", err)
		}
		return true
	})
}

func saveWinAccount(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("WinAccount"))
	winAccount.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinAccountEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("saveWinAccount err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("saveWinAccount err=%v", err)
		}
		return true
	})
}

func saveWinKerberosTGT(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("WinKerberosTGT"))
	winKerberosTGT.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinKerberosTGTEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("saveWinKerberosTGT err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("saveWinKerberosTGT err=%v", err)
		}
		return true
	})
}

func saveWinKerberosST(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("WinKerberosST"))
	winKerberosST.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinKerberosSTEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("saveWinKerberosST err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("saveWinKerberosST err=%v", err)
		}
		return true
	})
}

func saveWinPrivilege(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("WinPrivilege"))
	winPrivilege.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinPrivilegeEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("saveWinPrivilege err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("saveWinPrivilege err=%v", err)
		}
		return true
	})
}

func saveWinProcess(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("WinProcess"))
	winProcess.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinProcessEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("saveWinProcess err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("saveWinProcess err=%v", err)
		}
		return true
	})
}

func saveWinTask(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("WinTask"))
	winTask.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinTaskEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("saveWinTask err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("saveWinTask err=%v", err)
		}
		return true
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
	case "sensor":
		sensors.Delete(id)
	case "WinEventID":
		winEventID.Delete(id)
	case "WinLogon":
		winLogon.Delete(id)
	case "WinAccount":
		winAccount.Delete(id)
	case "WinKerberosTGT":
		winKerberosTGT.Delete(id)
	case "WinKebrosST":
		winKerberosST.Delete(id)
	case "WinPrivilege":
		winPrivilege.Delete(id)
	case "WinProcess":
		winProcess.Delete(id)
	case "WinTask":
		winTask.Delete(id)
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
	switch r {
	case "devices":
		deleteSyncMapAllData(&devices)
	case "ips":
		deleteSyncMapAllData(&ips)
	case "users":
		deleteSyncMapAllData(&users)
	case "flows":
		deleteSyncMapAllData(&flows)
	case "servers":
		deleteSyncMapAllData(&servers)
	case "ether":
		deleteSyncMapAllData(&etherType)
	case "dns":
		deleteSyncMapAllData(&dnsq)
	case "radius":
		deleteSyncMapAllData(&radiusFlows)
	case "tls":
		deleteSyncMapAllData(&tlsFlows)
	case "cert":
		deleteSyncMapAllData(&certs)
	case "sensor":
		deleteSyncMapAllData(&sensors)
	case "WinEventID":
		deleteSyncMapAllData(&winEventID)
	case "WinLogon":
		deleteSyncMapAllData(&winLogon)
	case "WinAccount":
		deleteSyncMapAllData(&winAccount)
	case "WinKeberosTGT":
		deleteSyncMapAllData(&winKerberosTGT)
	case "WinKerberosST":
		deleteSyncMapAllData(&winKerberosST)
	case "WinPrivilege":
		deleteSyncMapAllData(&winPrivilege)
	case "WinProcess":
		deleteSyncMapAllData(&winProcess)
	case "WinTask":
		deleteSyncMapAllData(&winTask)
	}
	return nil
}

func deleteSyncMapAllData(m *sync.Map) {
	m.Range(func(k, v interface{}) bool {
		m.Delete(k)
		return true
	})
}
