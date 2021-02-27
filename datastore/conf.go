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
	EnableArpWatch bool
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

func initConf() {
	MapConf.Community = "public"
	MapConf.PollInt = 60
	MapConf.Retry = 1
	MapConf.Timeout = 1
	MapConf.LogDispSize = 5000
	MapConf.LogDays = 14
	MapConf.AILevel = "info"
	MapConf.AIThreshold = 81
	MapConf.Community = "public"
	MapConf.UserID = "twsnmp"
	MapConf.Password = security.PasswordHash("twsnmp")
	DiscoverConf.Retry = 1
	DiscoverConf.Timeout = 1
	NotifyConf.InsecureSkipVerify = true
	NotifyConf.Interval = 60
	NotifyConf.Subject = "TWSNMPからの通知"
	NotifyConf.Level = "none"
	InfluxdbConf.DB = "twsnmp"
}

func loadConfFromDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	bSaveConf := false
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("mapConf"))
		if v == nil {
			bSaveConf = true
			return nil
		}
		if err := json.Unmarshal(v, &MapConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal mapConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("discoverConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &DiscoverConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal discoverConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("notifyConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &NotifyConf); err != nil {
			log.Println(fmt.Sprintf("Unmarshal notifyConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("backup"))
		if v != nil {
			if err := json.Unmarshal(v, &Backup); err != nil {
				log.Println(fmt.Sprintf("Unmarshal mainWinbackupdowInfo from DB error=%v", err))
			}
		}
		v = b.Get([]byte("influxdbConf"))
		if v != nil {
			if err := json.Unmarshal(v, &InfluxdbConf); err != nil {
				log.Println(fmt.Sprintf("Unmarshal influxdbConf from DB error=%v", err))
			}
		}
		return nil
	})
	if err == nil && MapConf.PrivateKey == "" {
		initSecurityKey()
	}
	if err == nil && bSaveConf {
		if err := SaveMapConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := SaveNotifyConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := SaveDiscoverConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
		if err := SaveInfluxdbConfToDB(); err != nil {
			log.Printf("loadConfFromDB err=%v", err)
		}
	}
	return err
}

func SaveBackImage(img []byte) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("backImage"), img)
	})
}

func GetBackImage() ([]byte, error) {
	var r []byte
	if db == nil {
		return r, ErrDBNotOpen
	}
	return r, db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		r = b.Get([]byte("backImage"))
		return nil
	})
}

func initSecurityKey() {
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
	MapConf.PrivateKey = key
	MapConf.PublicKey = pubkey
	MapConf.TLSCert = cert
	log.Printf("initSecurityKey Public Key=%v", pubkey)
	_ = SaveMapConfToDB()
}

func GetPrivateKey() string {
	return security.GetRawKeyPem(MapConf.PrivateKey)
}

func SaveMapConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(MapConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("mapConf"), s)
	})
}

func SaveNotifyConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(NotifyConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("notifyConf"), s)
	})
}

func SaveDiscoverConfToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(DiscoverConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("discoverConf"), s)
	})
}
