package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
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
	EnableSshd      bool
	EnableSflowd    bool
	EnableTcpd      bool
	EnableOTel      bool
	EnableMobileAPI bool
	AILevel         string
	AIThreshold     int
	AIMode          string
	GeoIPInfo       string
	FontSize        int
	AutoCharCode    bool
	DisableOperLog  bool
	MapSize         int
	IconSize        int
	ArpWatchRange   string
	OTelRetention   int
	OTelFrom        string
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
	MapConf.AIMode = "iforest"
	MapConf.Community = "public"
	MapConf.UserID = "twsnmp"
	MapConf.Password = security.PasswordHash("twsnmp")
	MapConf.EnableArpWatch = true
	MapConf.FontSize = 12
	MapConf.OTelRetention = 3
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
		v = b.Get([]byte("notifySchedule"))
		if v != nil {
			if err := json.Unmarshal(v, &NotifySchedule); err != nil {
				log.Printf("load conf err=%v", err)
			}
		}
		loadNotifyOAuth2Token(b)
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
	if MapConf.ArpWatchRange == "" {
		checkArpWatchRange()
		SaveMapConf()
	}
	return err
}

func SaveBackImage(img []byte) error {
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveBackImage dur=%v", time.Since(st))
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
	st := time.Now()
	imageListCache = []string{}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveImage dur=%v", time.Since(st))
		return b.Put([]byte(path), img)
	})
}

func DeleteImage(path string) error {
	st := time.Now()
	imageListCache = []string{}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("DeleteImage dur=%v", time.Since(st))
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
	key, err := security.GenPrivateKey(4096)
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
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	checkArpWatchRange()
	s, err := json.Marshal(MapConf)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveMapConf dur=%v", time.Since(st))
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
	st := time.Now()
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
		log.Printf("saveIcons dur=%v", time.Since(st))
		return b.Put([]byte("icons"), s)
	})
}

func GetSshdPublicKeys() string {
	r := ""
	if db == nil {
		return r
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		r = string(b.Get([]byte("sshdPublicKeys")))
		return nil
	})
	return r
}

func SaveSshdPublicKeys(pk string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("sshdPublicKeys"), []byte(pk))
	})
}

// ResetPassword : set user:password to twsnmp:twsnmp
func ResetPassword(ds string) error {
	dbPath := filepath.Join(ds, "twsnmpfc.db")
	_, err := os.Stat(dbPath)
	if err != nil {
		return err
	}
	d, err := bbolt.Open(dbPath, 0444, nil)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("no config bucket")
		}
		v := b.Get([]byte("mapConf"))
		if v == nil {
			return fmt.Errorf("no map config")
		}
		if err := json.Unmarshal(v, &MapConf); err != nil {
			return err
		}
		MapConf.UserID = "twsnmp"
		MapConf.Password = security.PasswordHash("twsnmp")
		j, err := json.Marshal(&MapConf)
		if err != nil {
			return err
		}
		return b.Put([]byte("mapConf"), []byte(j))
	})
}

// GetImageIcon returns the image icon for the given id.
func GetImageIcon(id string) ([]byte, error) {
	return os.ReadFile(filepath.Join(dspath, "icons", id))
}

// ARP監視のIP範囲をネットワークインターフェースから取得する
func checkArpWatchRange() bool {
	if MapConf.ArpWatchRange != "" {
		return false
	}
	ifs, err := net.Interfaces()
	if err != nil {
		log.Printf("check app watch range err=%v", err)
		return false
	}
	cidrs := []string{}
	cidrMap := make(map[string]bool)
	for _, i := range ifs {
		if (i.Flags&net.FlagLoopback) == net.FlagLoopback ||
			(i.Flags&net.FlagUp) != net.FlagUp ||
			(i.Flags&net.FlagPointToPoint) == net.FlagPointToPoint ||
			len(i.HardwareAddr) != 6 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ip, ipnet, err := net.ParseCIDR(a.String())
			if err != nil {
				continue
			}
			if ip.To4() == nil || !ip.IsGlobalUnicast() {
				continue
			}
			if !strings.Contains(a.String(), ".") {
				continue
			}
			r := ipnet.String()
			if _, ok := cidrMap[r]; ok {
				//重複しないようにする
				continue
			}
			cidrMap[r] = true
			cidrs = append(cidrs, r)
		}
	}
	MapConf.ArpWatchRange = strings.Join(cidrs, ",")
	return MapConf.ArpWatchRange != ""
}
