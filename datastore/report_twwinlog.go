package datastore

import (
	"encoding/json"
	"log"

	"go.etcd.io/bbolt"
)

// WinEventIDEnt represents a Windows Event ID entry.
// type=EventID,computer=%s,channel=%s,provider=%s,eventID=%d,total=%d,count=%d,ft=%s,lt=%s
type WinEventIDEnt struct {
	ID        string // Computer + Provider + EventID
	Level     string
	Computer  string
	Provider  string
	Channel   string
	EventID   int
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetWinEventID(id string) *WinEventIDEnt {
	if v, ok := winEventID.Load(id); ok {
		return v.(*WinEventIDEnt)
	}
	return nil
}

func AddWinEventID(e *WinEventIDEnt) {
	winEventID.Store(e.ID, e)
}

func ForEachWinEventID(f func(*WinEventIDEnt) bool) {
	winEventID.Range(func(k, v interface{}) bool {
		e := v.(*WinEventIDEnt)
		return f(e)
	})
}

// type=Logon,subject=@,target=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,ip=192.168.1.250,logonType=Network,time=2021-08-19T05:17:43+09:00
// type=LogonFailed,subject=@,target=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,ip=192.168.1.9,logonType=Network,failedCode=BadPassword,time=2021-08-19T04:46:28+09:00
// type=Logoff,subject=@,target=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,ip=,logonType=Network,time=2021-08-19T04:46:10+09:00

type WinLogonEnt struct {
	ID         string // target + computer + IP
	Target     string
	Computer   string
	IP         string
	Count      int64
	Logon      int64
	Logoff     int64
	Failed     int64
	LogonType  map[string]int
	FailedCode map[string]int
	Score      float64
	ValidScore bool
	Penalty    int64
	FirstTime  int64
	LastTime   int64
}

func GetWinLogon(id string) *WinLogonEnt {
	if v, ok := winLogon.Load(id); ok {
		return v.(*WinLogonEnt)
	}
	return nil
}

func AddWinLogon(e *WinLogonEnt) {
	winLogon.Store(e.ID, e)
}

func ForEachWinLogon(f func(*WinLogonEnt) bool) {
	winLogon.Range(func(k, v interface{}) bool {
		e := v.(*WinLogonEnt)
		return f(e)
	})
}

// WinAccountEnt represents a Windows account event entry.
// type=Account,subject=%s,target=%s,computer=%s,count=%d,edit=%d,password=%d,other=%d,ft=%s,lt=%s",
type WinAccountEnt struct {
	ID        string // subject + target + computer
	Subject   string
	Target    string
	Computer  string
	Count     int64
	Edit      int64
	Password  int64
	Other     int64
	FirstTime int64
	LastTime  int64
}

func GetWinAccount(id string) *WinAccountEnt {
	if v, ok := winAccount.Load(id); ok {
		return v.(*WinAccountEnt)
	}
	return nil
}

func AddWinAccount(e *WinAccountEnt) {
	winAccount.Store(e.ID, e)
}

func ForEachWinAccount(f func(*WinAccountEnt) bool) {
	winAccount.Range(func(k, v interface{}) bool {
		e := v.(*WinAccountEnt)
		return f(e)
	})
}

// WinKerberosEnt represents a Windows Kerberos event entry.
// type=Kerberos,target=%s,computer=%s,ip=%s,service=%s,ticketType=%s,count=%d,failed=%d,status=%s,cert=%s,ft=%s,lt=%s
// type=KerberosFailed,target=%s,computer=%s,ip=%s,service=%s,ticketType=%s,status=%s,time=%s
type WinKerberosEnt struct {
	ID         string // target + computer + ip  + service + ticketType
	Target     string
	Computer   string
	IP         string
	Service    string
	TicketType string
	Count      int64
	Failed     int64
	LastCert   string
	LastStatus string
	Score      float64
	ValidScore bool
	Penalty    int64
	FirstTime  int64
	LastTime   int64
}

func GetWinKerberos(id string) *WinKerberosEnt {
	if v, ok := winKerberos.Load(id); ok {
		return v.(*WinKerberosEnt)
	}
	return nil
}

func AddWinKerberos(e *WinKerberosEnt) {
	winKerberos.Store(e.ID, e)
}

func ForEachWinKerberos(f func(*WinKerberosEnt) bool) {
	winKerberos.Range(func(k, v interface{}) bool {
		e := v.(*WinKerberosEnt)
		return f(e)
	})
}

// WinPrivilegeEnt represents a Windows privilege event entry.
// type=Privilege,subject=%s,computer=%s,count=%d,ft=%s,lt=%s
type WinPrivilegeEnt struct {
	ID        string //  subject + computer
	Subject   string
	Computer  string
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetWinPrivilege(id string) *WinPrivilegeEnt {
	if v, ok := winPrivilege.Load(id); ok {
		return v.(*WinPrivilegeEnt)
	}
	return nil
}

func AddWinPrivilege(e *WinPrivilegeEnt) {
	winPrivilege.Store(e.ID, e)
}

func ForEachWinPrivilege(f func(*WinPrivilegeEnt) bool) {
	winPrivilege.Range(func(k, v interface{}) bool {
		e := v.(*WinPrivilegeEnt)
		return f(e)
	})
}

// WinProcessEnt represents a Windows process event entry.
// type=Process,computer=%s,process=%s,count=%d,start=%d,exit=%d,subject=%s,status=%s,parent=%s,ft=%s,lt=%s",
type WinProcessEnt struct {
	ID          string // Computer + Process
	Computer    string
	Process     string
	Count       int64
	Start       int64
	Exit        int64
	LastParent  string
	LastSubject string
	LastStatus  string
	FirstTime   int64
	LastTime    int64
}

func GetWinProcess(id string) *WinProcessEnt {
	if v, ok := winProcess.Load(id); ok {
		return v.(*WinProcessEnt)
	}
	return nil
}

func AddWinProcess(e *WinProcessEnt) {
	winProcess.Store(e.ID, e)
}

func ForEachWinProcess(f func(*WinProcessEnt) bool) {
	winProcess.Range(func(k, v interface{}) bool {
		e := v.(*WinProcessEnt)
		return f(e)
	})
}

// WinTaskEnt represents a Windows scheduled task event entry.
// type=Task,subject=%s,taskname=%s,computer=%s,count=%d,ft=%s,lt=%s",
type WinTaskEnt struct {
	ID        string // Computer + TaskName + Subject
	TaskName  string
	Computer  string
	Subject   string
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetWinTask(id string) *WinTaskEnt {
	if v, ok := winTask.Load(id); ok {
		return v.(*WinTaskEnt)
	}
	return nil
}

func AddWinTask(e *WinTaskEnt) {
	winTask.Store(e.ID, e)
}

func ForEachWinTask(f func(*WinTaskEnt) bool) {
	winTask.Range(func(k, v interface{}) bool {
		e := v.(*WinTaskEnt)
		return f(e)
	})
}

// internal use
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

func loadWinKerberos(r *bbolt.Bucket) {
	b := r.Bucket([]byte("winKerberos"))
	if b != nil {
		_ = b.ForEach(func(k, v []byte) error {
			var e WinKerberosEnt
			if err := json.Unmarshal(v, &e); err == nil {
				winKerberos.Store(e.ID, &e)
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

func saveWinEventID(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("winEventID"))
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
			log.Printf("save winEventID report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save winEventID report err=%v", err)
		}
		return true
	})
}

func saveWinLogon(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("winLogon"))
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
			log.Printf("save winLogon report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save winLogon report  err=%v", err)
		}
		return true
	})
}

func saveWinAccount(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("winAccount"))
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
			log.Printf("save winAccount report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save winAccount report err=%v", err)
		}
		return true
	})
}

func saveWinKerberos(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("winKerberos"))
	winKerberos.Range(func(k, v interface{}) bool {
		e, ok := v.(*WinKerberosEnt)
		if !ok {
			return true
		}
		if e.LastTime < last {
			return true
		}
		s, err := json.Marshal(e)
		if err != nil {
			log.Printf("save winKerberos report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save winKerberos report err=%v", err)
		}
		return true
	})
}

func saveWinPrivilege(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("winPrivilege"))
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
			log.Printf("save winPrivilege report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save winPrivilege report err=%v", err)
		}
		return true
	})
}

func saveWinProcess(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("winProcess"))
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
			log.Printf("save winProcess report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save winProcess report err=%v", err)
		}
		return true
	})
}

func saveWinTask(b *bbolt.Bucket, last int64) {
	r := b.Bucket([]byte("winTask"))
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
			log.Printf("save winTask report err=%v", err)
			return true
		}
		err = r.Put([]byte(e.ID), s)
		if err != nil {
			log.Printf("save winTask report err=%v", err)
		}
		return true
	})
}
