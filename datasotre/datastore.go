// Package datastore : データ保存
package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

var (
	db          *bbolt.DB
	prevDBStats bbolt.Stats
	dbOpenTime  time.Time
	// Data on Memory
	MapConf          MapConfEnt
	NotifyConf       NotifyConfEnt
	InfluxdbConf     InfluxdbConfEnt
	RestAPIConf      RestAPIConfEnt
	DiscoverConf     DiscoverConfEnt
	DBStats          DBStatsEnt
	Nodes            = sync.Map{}
	Lines            = sync.Map{}
	Pollings         = sync.Map{}
	PollingTemplates = make(map[string]*PollingTemplateEnt)
)

const (
	// MaxDispLog : ログの検索結果の最大値
	MaxDispLog = 20000
	// MaxDelLog : ログ削除処理の最大削除数
	MaxDelLog = 500000
)

// Define errors
var (
	ErrNoPayload     = fmt.Errorf("no payload")
	ErrInvalidNode   = fmt.Errorf("invalid node")
	ErrInvalidParams = fmt.Errorf("invald params")
	ErrDBNotOpen     = fmt.Errorf("db not open")
	ErrInvalidID     = fmt.Errorf("invalid id")
)

type NodeEnt struct {
	ID        string
	Name      string
	Descr     string
	Icon      string
	State     string
	X         int
	Y         int
	IP        string
	MAC       string
	SnmpMode  string
	Community string
	User      string
	Password  string
	PublicKey string
	URL       string
	Type      string
	AddrMode  string
}

type LineEnt struct {
	ID         string
	NodeID1    string
	PollingID1 string
	State1     string
	NodeID2    string
	PollingID2 string
	State2     string
}

type PollingEnt struct {
	ID         string
	Name       string
	NodeID     string
	Type       string
	Polling    string
	Level      string
	PollInt    int
	Timeout    int
	Retry      int
	LogMode    int
	NextTime   int64
	LastTime   int64
	LastResult string
	LastVal    float64
	State      string
}

type PollingTemplateEnt struct {
	ID       string
	Name     string
	Type     string
	Polling  string
	Level    string
	NodeType string
	Descr    string
}

type EventLogEnt struct {
	Time     int64 // UnixNano()
	Type     string
	Level    string
	NodeName string
	NodeID   string
	Event    string
}

type PollingLogEnt struct {
	Time      int64 // UnixNano()
	PollingID string
	State     string
	NumVal    float64
	StrVal    string
}

type LogEnt struct {
	Time int64 // UnixNano()
	Type string
	Log  string
}

// MapConfEnt :  マップ設定
type MapConfEnt struct {
	MapName        string
	PollInt        int
	Timeout        int
	Retry          int
	LogDays        int
	LogDispSize    int
	NodeSort       string
	SnmpMode       string
	Community      string
	User           string
	Password       string
	PublicKey      string
	PrivateKey     string
	TLSCert        string
	EnableSyslogd  bool
	EnableTrapd    bool
	EnableNetflowd bool
	BackImg        string
	GeoIPPath      string
	GrokPath       string
	ArpWatchLevel  string
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
	Report             string
	ExecCmd            string
	CheckUpdate        bool
	NotifyRepair       bool
}

type DiscoverConfEnt struct {
	SnmpMode string
	StartIP  string
	EndIP    string
	Timeout  int
	Retry    int
	X        int
	Y        int
}

type AIResult struct {
	PollingID string
	LastTime  int64
	LossData  [][]float64
	ScoreData [][]float64
}

type DBStatsEnt struct {
	Time             string
	Size             int64
	TotalWrite       int64
	LastWrite        int64
	PeakWrite        int64
	AvgWrite         float64
	StartTime        time.Time
	Speed            float64
	Peak             float64
	Rate             float64
	BackupConfigOnly bool
	BackupDaily      bool
	BackupFile       string
	BackupTime       string
}

type InfluxdbConfEnt struct {
	URL        string
	User       string
	Password   string
	DB         string
	Duration   string
	PollingLog string
	AIScore    string
}

type RestAPIConfEnt struct {
	Port     int
	User     string
	Password string
}

type DBBackupParamEnt struct {
	ConfigOnly bool
	Daily      bool
	BackupFile string
}

func CheckDB(path string) error {
	var err error
	d, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	defer d.Close()
	return nil
}

func OpenDB(path string) error {
	var err error
	db, err = bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	prevDBStats = db.Stats()
	dbOpenTime = time.Now()
	err = initDB()
	if err != nil {
		db.Close()
		return err
	}
	err = loadConfFromDB()
	if err != nil {
		db.Close()
		return err
	}
	err = loadMapDataFromDB()
	if err != nil {
		db.Close()
		return err
	}
	_ = loadPollingTemplateFromDB()
	return nil
}

func initDB() error {
	buckets := []string{"config", "nodes", "lines", "pollings", "logs", "pollingLogs",
		"syslog", "trap", "netflow", "ipfix", "arplog", "mibdb", "arp", "ai", "report", "pollingTemplates"}
	reports := []string{"devices", "users", "flows", "servers", "allows", "dennys"}
	initConf()
	return db.Update(func(tx *bbolt.Tx) error {
		for _, b := range buckets {
			pb, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
			if b == "report" {
				for _, r := range reports {
					if _, err := pb.CreateBucketIfNotExists([]byte(r)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

// CloseDB : DBをクローズする
func CloseDB() {
	if db == nil {
		return
	}
	saveLogList([]EventLogEnt{{
		Type:  "system",
		Level: "info",
		Time:  time.Now().UnixNano(),
		Event: "TWSNMP終了",
	}})
	db.Close()
	db = nil
}

func UpdateDBStats() {
	if db == nil {
		return
	}
	s := db.Stats()
	d := s.Sub(&prevDBStats)
	var dbSize int64
	_ = db.View(func(tx *bbolt.Tx) error {
		dbSize = tx.Size()
		return nil
	})
	DBStats.Size = dbSize
	DBStats.TotalWrite = int64(s.TxStats.Write)
	DBStats.LastWrite = int64(d.TxStats.Write)
	if DBStats.PeakWrite < DBStats.LastWrite {
		DBStats.PeakWrite = DBStats.LastWrite
	}
	// 初回は計算しない。
	if DBStats.PeakWrite > 0 && DBStats.Time != "" {
		DBStats.Rate = 100 * float64(d.TxStats.Write) / float64(DBStats.PeakWrite)
		DBStats.StartTime = dbOpenTime
		dbot := time.Since(dbOpenTime).Seconds()
		if dbot > 0 {
			DBStats.AvgWrite = float64(s.TxStats.Write) / dbot
		}
	}
	dt := d.TxStats.WriteTime.Seconds()
	if dt != 0 {
		DBStats.Speed = float64(d.TxStats.Write) / dt
		if DBStats.Peak < DBStats.Speed {
			DBStats.Peak = DBStats.Speed
		}
	} else {
		DBStats.Speed = 0.0
	}
	DBStats.Time = time.Now().Format("15:04:05")
	prevDBStats = s

	if DBStats.BackupFile != "" && nextBackup != 0 && nextBackup < time.Now().UnixNano() {
		nextBackup += (24 * 3600 * 1000 * 1000 * 1000)
		go func() {
			log.Printf("Backup start = %s", DBStats.BackupFile)
			AddEventLog(EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ開始:" + DBStats.BackupFile,
			})
			if err := BackupDB(); err != nil {
				log.Printf("backupDB err=%v", err)
			}
			log.Printf("Backup end = %s", DBStats.BackupFile)
			AddEventLog(EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ終了:" + DBStats.BackupFile,
			})
		}()
		DBStats.BackupTime = DBStats.Time
	}
}

// bboltに保存する場合のキーを時刻から生成する。
func makeKey() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}

var stopBackup = false
var nextBackup int64
var dbBackupSize int64
var dstDB *bbolt.DB
var dstTx *bbolt.Tx

func BackupDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	if dstDB != nil {
		return fmt.Errorf("backup in progress")
	}
	os.Remove(DBStats.BackupFile)
	var err error
	dstDB, err = bbolt.Open(DBStats.BackupFile, 0600, nil)
	if err != nil {
		return err
	}
	defer func() {
		dstDB.Close()
		dstDB = nil
	}()
	dstTx, err = dstDB.Begin(true)
	if err != nil {
		return err
	}
	err = db.View(func(srcTx *bbolt.Tx) error {
		return srcTx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			return walkBucket(b, nil, name, nil, b.Sequence())
		})
	})
	if err != nil {
		_ = dstTx.Rollback()
		return err
	}
	if !DBStats.BackupConfigOnly {
		mapConfTmp := MapConf
		mapConfTmp.EnableNetflowd = false
		mapConfTmp.EnableSyslogd = false
		mapConfTmp.EnableTrapd = false
		mapConfTmp.LogDays = 0
		if s, err := json.Marshal(mapConfTmp); err == nil {
			if b := dstTx.Bucket([]byte("config")); b != nil {
				return b.Put([]byte("mapConf"), s)
			}
		}
	}
	return dstTx.Commit()
}

var configBuckets = []string{"config", "nodes", "lines", "pollings", "mibdb"}

func walkBucket(b *bbolt.Bucket, keypath [][]byte, k, v []byte, seq uint64) error {
	if stopBackup {
		return fmt.Errorf("stop backup")
	}
	if DBStats.BackupConfigOnly && v == nil {
		c := false
		for _, cbn := range configBuckets {
			if k != nil && cbn == string(k) {
				c = true
				break
			}
		}
		if !c {
			return nil
		}
	}
	if dbBackupSize > 64*1024 {
		_ = dstTx.Commit()
		var err error
		dstTx, err = dstDB.Begin(true)
		if err != nil {
			return err
		}
		dbBackupSize = 0
	}
	// Execute callback.
	if err := walkFunc(keypath, k, v, seq); err != nil {
		return err
	}
	dbBackupSize += int64(len(k) + len(v))

	// If this is not a bucket then stop.
	if v != nil {
		return nil
	}

	// Iterate over each child key/value.
	keypath = append(keypath, k)
	return b.ForEach(func(k, v []byte) error {
		if v == nil {
			bkt := b.Bucket(k)
			return walkBucket(bkt, keypath, k, nil, bkt.Sequence())
		}
		return walkBucket(b, keypath, k, v, b.Sequence())
	})
}

func walkFunc(keys [][]byte, k, v []byte, seq uint64) error {
	// Create bucket on the root transaction if this is the first level.
	nk := len(keys)
	if nk == 0 {
		bkt, err := dstTx.CreateBucket(k)
		if err != nil {
			return err
		}
		if err := bkt.SetSequence(seq); err != nil {
			return err
		}
		return nil
	}
	// Create buckets on subsequent levels, if necessary.
	b := dstTx.Bucket(keys[0])
	if nk > 1 {
		for _, k := range keys[1:] {
			b = b.Bucket(k)
		}
	}
	// Fill the entire page for best compaction.
	b.FillPercent = 1.0
	// If there is no value then this is a bucket call.
	if v == nil {
		bkt, err := b.CreateBucket(k)
		if err != nil {
			return err
		}
		if err := bkt.SetSequence(seq); err != nil {
			return err
		}
		return nil
	}
	// Otherwise treat it as a key/value pair.
	return b.Put(k, v)
}
