package datastore

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/twsnmp/twsnmpfc/security"
	"go.etcd.io/bbolt"
)

type backImage struct {
	Path   string
	X      int
	Y      int
	Width  int
	Height int
}

// MapConfEnt :  マップ設定
type MapConfEnt struct {
	MapName        string
	BackImage      backImage
	UserID         string
	Password       string
	PollInt        int
	Timeout        int
	Retry          int
	LogDays        int
	LogDispSize    int
	SnmpMode       string
	Community      string
	SnmpUser       string
	SnmpPassword   string
	PublicKey      string
	PrivateKey     string
	TLSCert        string
	EnableSyslogd  bool
	EnableTrapd    bool
	EnableNetflowd bool
	AILevel        string
	AIThreshold    int
}

type NotifyConfEnt struct {
	MailServer         string
	User               string
	Password           string
	InsecureSkipVerify bool
	MailTo             string
	MailFrom           string
	Subject            string
	Interval           int
	Level              string
	Report             bool
	CheckUpdate        bool
	NotifyRepair       bool
}

type DiscoverConfEnt struct {
	StartIP string
	EndIP   string
	Timeout int
	Retry   int
	X       int
	Y       int
}

func (ds *DataStore) initConf() {
	ds.MapConf.Community = "public"
	ds.MapConf.PollInt = 60
	ds.MapConf.Retry = 1
	ds.MapConf.Timeout = 1
	ds.MapConf.LogDispSize = 5000
	ds.MapConf.LogDays = 14
	ds.MapConf.AILevel = "info"
	ds.MapConf.AIThreshold = 81
	ds.MapConf.Community = "public"
	ds.MapConf.UserID = "twsnmp"
	ds.MapConf.Password = security.PasswordHash("twsnmp")
	ds.DiscoverConf.Retry = 1
	ds.DiscoverConf.Timeout = 1
	ds.NotifyConf.InsecureSkipVerify = true
	ds.NotifyConf.Interval = 60
	ds.NotifyConf.Subject = "TWSNMPからの通知"
	ds.NotifyConf.Level = "none"
	ds.InfluxdbConf.DB = "twsnmp"
}

func (ds *DataStore) loadConfFromDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	bSaveConf := false
	err := ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("mapConf"))
		if v == nil {
			bSaveConf = true
			return nil
		}
		if err := json.Unmarshal(v, &ds.MapConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal mapConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("discoverConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &ds.DiscoverConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal discoverConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("notifyConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &ds.NotifyConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal notifyConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("backup"))
		if v != nil {
			if err := json.Unmarshal(v, &ds.Backup); err != nil {
				log.Println(fmt.Sprintf("Unmarshal mainWinbackupdowInfo from DB error=%v", err))
			}
		}
		v = b.Get([]byte("influxdbConf"))
		if v != nil {
			if err := json.Unmarshal(v, &ds.InfluxdbConf); err != nil {
				log.Println(fmt.Sprintf("Unmarshal influxdbConf from DB error=%v", err))
			}
		}
		return nil
	})
	if err == nil && ds.MapConf.PrivateKey == "" {
		ds.initSecurityKey()
	}
	if err == nil && bSaveConf {
		if err := ds.SaveMapConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := ds.SaveNotifyConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := ds.SaveDiscoverConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := ds.SaveInfluxdbConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
	}
	return err
}

func (ds *DataStore) SaveBackImage(img []byte) error {
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("backImage"), img)
	})
}

func (ds *DataStore) GetBackImage() ([]byte, error) {
	var r []byte
	if ds.db == nil {
		return r, ErrDBNotOpen
	}
	return r, ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		r = b.Get([]byte("backImage"))
		return nil
	})
}

func (ds *DataStore) initSecurityKey() {
	key, err := security.GenPrivateKey(4096)
	if err != nil {
		log.Printf("initSecurityKey err=%v", err)
		return
	}
	pubkey, err := security.GetSSHPublicKey(key)
	if err != nil {
		log.Printf("initSecurityKey err=%v", err)
		return
	}
	cert, err := security.GenSelfSignCert(key)
	if err != nil {
		log.Printf("initSecurityKey err=%v", err)
		return
	}
	ds.MapConf.PrivateKey = key
	ds.MapConf.PublicKey = pubkey
	ds.MapConf.TLSCert = cert
	log.Printf("initSecurityKey Public Key=%v", pubkey)
	_ = ds.SaveMapConfToDB()
}

func (ds *DataStore) GetPrivateKey() string {
	return security.GetRawKeyPem(ds.MapConf.PrivateKey)
}

func (ds *DataStore) SaveMapConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.MapConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("mapConf"), s)
	})
}

func (ds *DataStore) SaveNotifyConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.NotifyConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("notifyConf"), s)
	})
}

func (ds *DataStore) SaveDiscoverConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.DiscoverConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("discoverConf"), s)
	})
}
