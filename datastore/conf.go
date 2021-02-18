package datastore

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
		var p DBBackupParamEnt
		v = b.Get([]byte("backup"))
		if v != nil {
			if err := json.Unmarshal(v, &p); err != nil {
				log.Println(fmt.Sprintf("Unmarshal mainWinbackupdowInfo from DB error=%v", err))
			} else {
				if p.BackupFile != "" && p.Daily {
					ds.DBStats.BackupConfigOnly = p.ConfigOnly
					ds.DBStats.BackupFile = p.BackupFile
					ds.DBStats.BackupDaily = p.Daily
					now := time.Now()
					d := 0
					if now.Hour() > 2 {
						d = 1
					}
					ds.nextBackup = time.Date(now.Year(), now.Month(), now.Day()+d, 3, 0, 0, 0, time.Local).UnixNano()
				}
			}
		}
		v = b.Get([]byte("influxdbConf"))
		if v != nil {
			if err := json.Unmarshal(v, &ds.InfluxdbConf); err != nil {
				log.Println(fmt.Sprintf("Unmarshal influxdbConf from DB error=%v", err))
			}
		}
		v = b.Get([]byte("restAPIConf"))
		if v != nil {
			if err := json.Unmarshal(v, &ds.RestAPIConf); err != nil {
				log.Println(fmt.Sprintf("Unmarshal restAPIConf from DB error=%v", err))
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

func (ds *DataStore) SaveInfluxdbConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.InfluxdbConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("influxdbConf"), s)
	})
}

func (ds *DataStore) SaveRestAPIConfToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.RestAPIConf)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("restAPIConf"), s)
	})
}

func (ds *DataStore) SaveBackupParamToDB(p *DBBackupParamEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("backup"), s)
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

func (ds *DataStore) loadPollingTemplateFromDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	err := ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			var pt PollingTemplateEnt
			if err := json.Unmarshal(v, &pt); err == nil {
				ds.pollingTemplates[pt.ID] = &pt
			}
			return nil
		})
		return nil
	})
	return err
}

func getSha1KeyForTemplate(s string) string {
	h := sha1.New()
	if _, err := h.Write([]byte(s)); err != nil {
		log.Printf("getSha1KeyForTemplate err=%v", err)
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func (ds *DataStore) AddPollingTemplate(pt *PollingTemplateEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	pt.ID = getSha1KeyForTemplate(pt.Name + ":" + pt.Type + ":" + pt.NodeType + ":" + pt.Polling)
	if _, ok := ds.pollingTemplates[pt.ID]; ok {
		return fmt.Errorf("duplicate template")
	}
	s, err := json.Marshal(pt)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		return b.Put([]byte(pt.ID), s)
	})
	ds.pollingTemplates[pt.ID] = pt
	return nil
}

func (ds *DataStore) UpdatePollingTemplate(pt *PollingTemplateEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.pollingTemplates[pt.ID]; !ok {
		return ErrInvalidID
	}
	newID := getSha1KeyForTemplate(pt.Name + ":" + pt.Type + ":" + pt.NodeType + ":" + pt.Polling)
	if newID != pt.ID {
		// 更新後に同じ内容のテンプレートがないか確認する
		if _, ok := ds.pollingTemplates[newID]; ok {
			return fmt.Errorf("duplicate template")
		}
	}
	// 削除してから追加する
	_ = ds.DeletePollingTemplate(pt.ID)
	pt.ID = newID
	return ds.AddPollingTemplate(pt)
}

func (ds *DataStore) DeletePollingTemplate(id string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if _, ok := ds.pollingTemplates[id]; !ok {
		return ErrInvalidID
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		return b.Delete([]byte(id))
	})
	delete(ds.pollingTemplates, id)
	return nil
}
