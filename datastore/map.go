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
	Color  string
}

// MapConfEnt :  マップ設定
type MapConfEnt struct {
	MapName      string
	BackImage    backImage
	UserID       string
	Password     string
	PollInt      int
	Timeout      int
	Retry        int
	LogDays      int
	LogDispSize  int
	LogTimeout   int
	SnmpMode     string
	Community    string
	SnmpUser     string
	SnmpPassword string
	PublicKey    string
	PrivateKey   string
	//	TLSCert        string
	EnableSyslogd   bool
	EnableTrapd     bool
	EnableNetflowd  bool
	EnableArpWatch  bool
	EnableMobileAPI bool
	AILevel         string
	AIThreshold     int
	AIMode          string
	GeoIPInfo       string
	FontSize        int
	AutoCharCode    bool
}

func initConf() {
	MapConf.Community = "public"
	MapConf.PollInt = 60
	MapConf.Retry = 1
	MapConf.Timeout = 1
	MapConf.LogDispSize = 5000
	MapConf.LogTimeout = 15
	MapConf.LogDays = 14
	MapConf.AILevel = "info"
	MapConf.AIThreshold = 81
	MapConf.AIMode = "lof"
	MapConf.Community = "public"
	MapConf.UserID = "twsnmp"
	MapConf.Password = security.PasswordHash("twsnmp")
	MapConf.EnableArpWatch = true
	MapConf.FontSize = 12
	DiscoverConf.Retry = 1
	DiscoverConf.Timeout = 1
	NotifyConf.InsecureSkipVerify = true
	NotifyConf.Interval = 60
	NotifyConf.Subject = "TWSNMPからの通知"
	NotifyConf.Level = "none"
	InfluxdbConf.DB = "twsnmp"
	Backup.ConfigOnly = true
}

func loadConf() error {
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
			return err
		}
		v = b.Get([]byte("discoverConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &DiscoverConf); err != nil {
			return err
		}
		v = b.Get([]byte("notifyConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &NotifyConf); err != nil {
			return err
		}
		v = b.Get([]byte("backup"))
		if v != nil {
			if err := json.Unmarshal(v, &Backup); err != nil {
				log.Printf("load conf err=%v", err)
			}
		}
		v = b.Get([]byte("influxdbConf"))
		if v != nil {
			if err := json.Unmarshal(v, &InfluxdbConf); err != nil {
				log.Printf("load conf err=%v", err)
			}
		}
		v = b.Get([]byte("icons"))
		if v != nil {
			if err := json.Unmarshal(v, &icons); err != nil {
				log.Printf("load icons err=%v", err)
			}
		}
		return nil
	})
	if err == nil && MapConf.PrivateKey == "" {
		InitSecurityKey()
	}
	if err == nil && bSaveConf {
		if err := SaveMapConf(); err != nil {
			log.Printf("load conf err=%v", err)
		}
		if err := SaveNotifyConf(); err != nil {
			log.Printf("load conf err=%v", err)
		}
		if err := SaveDiscoverConf(); err != nil {
			log.Printf("load conf err=%v", err)
		}
		if err := SaveInfluxdbConf(); err != nil {
			log.Printf("load conf err=%v", err)
		}
	}
	return err
}

func SaveBackImage(img []byte) error {
	return db.Batch(func(tx *bbolt.Tx) error {
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

var imageListCache = []string{}

func SaveImage(path string, img []byte) error {
	imageListCache = []string{}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte(path), img)
	})
}

func DelteImage(path string) error {
	imageListCache = []string{}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Delete([]byte(path))
	})
}

func GetImageList() []string {
	if db == nil || len(imageListCache) > 0 {
		return imageListCache
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket iamges is nil")
		}
		return b.ForEach(func(k, v []byte) error {
			imageListCache = append(imageListCache, string(k))
			return nil
		})
	})
	return imageListCache
}

func GetImage(path string) ([]byte, error) {
	var r []byte
	if db == nil {
		return r, ErrDBNotOpen
	}
	return r, db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket iamges is nil")
		}
		r = b.Get([]byte(path))
		return nil
	})
}

func InitSecurityKey() {
	key, err := security.GenPrivateKey(4096, "")
	if err != nil {
		log.Printf("init security key err=%v", err)
		return
	}
	pubkey, err := security.GetSSHPublicKey(key)
	if err != nil {
		log.Printf("init security key err=%v", err)
		return
	}
	MapConf.PrivateKey = key
	MapConf.PublicKey = pubkey
	_ = SaveMapConf()
}

func GetPrivateKey() string {
	return security.GetRawKeyPem(MapConf.PrivateKey, "")
}

func SaveMapConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(MapConf)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("mapConf"), s)
	})
}

type IconEnt struct {
	Text string
	Icon string
	Code int64
}

var icons []*IconEnt

func GetIcons() []*IconEnt {
	return icons
}

func AddOrUpdateIcon(i *IconEnt) error {
	for _, e := range icons {
		if e.Icon == i.Icon {
			e.Text = i.Text
			e.Code = i.Code
			return saveIcons()
		}
	}
	icons = append(icons, i)
	return saveIcons()
}

func DeleteIcon(icon string) error {
	tmp := icons
	icons = []*IconEnt{}
	for _, i := range tmp {
		if i.Icon != icon {
			icons = append(icons, i)
		}
	}
	return saveIcons()
}

func saveIcons() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(icons)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("icons"), s)
	})
}
