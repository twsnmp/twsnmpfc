// Package datastore : データ保存
package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"

	"github.com/oschwald/geoip2-golang"
	gomibdb "github.com/twsnmp/go-mibdb"
	"go.etcd.io/bbolt"
)

type DataStore struct {
	db          *bbolt.DB
	prevDBStats bbolt.Stats
	dbOpenTime  time.Time
	// Conf Data on Memory
	MapConf      MapConfEnt
	NotifyConf   NotifyConfEnt
	InfluxdbConf InfluxdbConfEnt
	RestAPIConf  RestAPIConfEnt
	DiscoverConf DiscoverConfEnt
	DBStats      DBStatsEnt
	// Map Data on Memory not export
	nodes            sync.Map
	lines            sync.Map
	pollings         sync.Map
	pollingTemplates map[string]*PollingTemplateEnt
	// Report Data on Memory not export
	devices    map[string]*DeviceEnt
	users      map[string]*UserEnt
	flows      map[string]*FlowEnt
	servers    map[string]*ServerEnt
	dennyRules map[string]bool
	allowRules map[string]*AllowRuleEnt
	// MAP Changed check
	stateChangedNodes sync.Map
	lastLogAdded      time.Time
	lastNodeChanged   time.Time
	//
	MIBDB        *gomibdb.MIBDB
	stopBackup   bool
	nextBackup   int64
	dbBackupSize int64
	dstDB        *bbolt.DB
	dstTx        *bbolt.Tx
	eventLogCh   chan *EventLogEnt
	delCount     int

	influxc   client.Client
	muInfluxc sync.Mutex

	protMap    map[int]string
	serviceMap map[string]string
	geoip      *geoip2.Reader
	geoipMap   map[string]string
	ouiMap     map[string]string
	tlsCSMap   map[string]string

	logSize     int64
	compLogSize int64
}

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

type PollingTemplateEnt struct {
	ID       string
	Name     string
	Type     string
	Polling  string
	Level    string
	NodeType string
	Descr    string
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

func NewDataStore(ctx context.Context, dspath string, fs http.FileSystem) *DataStore {
	ds := &DataStore{
		devices:          make(map[string]*DeviceEnt),
		users:            make(map[string]*UserEnt),
		flows:            make(map[string]*FlowEnt),
		servers:          make(map[string]*ServerEnt),
		dennyRules:       make(map[string]bool),
		allowRules:       make(map[string]*AllowRuleEnt),
		eventLogCh:       make(chan *EventLogEnt, 100),
		pollingTemplates: make(map[string]*PollingTemplateEnt),
		protMap: map[int]string{
			1:   "icmp",
			2:   "igmp",
			6:   "tcp",
			8:   "egp",
			17:  "udp",
			112: "vrrp",
		},
		serviceMap: make(map[string]string),
		geoipMap:   make(map[string]string),
		ouiMap:     make(map[string]string),
		tlsCSMap:   make(map[string]string),
	}
	ds.InitDataStore(dspath, fs)
	go ds.eventLogger(ctx)
	return ds
}

func (ds *DataStore) InitDataStore(dspath string, fs http.FileSystem) {
	if dspath == "" {
		log.Println("No DataStore Path Skip Init")
		return
	}
	// BBoltをオープン
	if err := ds.OpenDB(filepath.Join(dspath, "twsnmpfc.db")); err != nil {
		log.Fatalf("InitDataStore OpenDB err=%v", err)
	}
	// MIBDB
	if r, err := os.Open(filepath.Join(dspath, "mib.txt")); err == nil {
		ds.loadMIBDB(r)
	} else {
		if r, err := fs.Open("/conf/mib.txt"); err == nil {
			ds.loadMIBDB(r)
		} else {
			log.Fatalf("InitDataStore MIBDB err=%v", err)
		}
	}
	// 拡張MIBの読み込み
	ds.loadExtMIBs(filepath.Join(dspath, "extmibs"))
	// サービスの定義ファイル、ユーザー指定があれば利用、なければ内蔵
	if r, err := os.Open(filepath.Join(dspath, "services.txt")); err == nil {
		ds.loadServiceMap(r)
	} else {
		if r, err := fs.Open("/conf/services.txt"); err == nil {
			ds.loadServiceMap(r)
		} else {
			log.Fatalf("InitDataStore services.txt err=%v", err)
		}
	}
	// OUIの定義
	if r, err := os.Open(filepath.Join(dspath, "oui.txt")); err == nil {
		ds.loadOUIMap(r)
	} else {
		if r, err := fs.Open("/conf/oui.txt"); err == nil {
			ds.loadOUIMap(r)
		} else {
			log.Fatalf("InitDataStore oui.txt err=%v", err)
		}
	}
	// TLS暗号名の定義
	if r, err := os.Open(filepath.Join(dspath, "tlsparams.csv")); err == nil {
		ds.loadTLSCihperNameMap(r)
	} else {
		if r, err := fs.Open("/conf/tlsparams.csv"); err == nil {
			ds.loadTLSCihperNameMap(r)
		} else {
			log.Fatalf("InitDataStore tlsparams.csv err=%v", err)
		}
	}
	p := filepath.Join(dspath, "geoip.mmdb")
	if _, err := os.Stat(p); err == nil {
		ds.openGeoIP(p)
	}
	p = filepath.Join(dspath, "grok.txt")
	if _, err := os.Stat(p); err == nil {
		ds.loadGrokMap(p)
	}
}

func (ds *DataStore) OpenDB(path string) error {
	var err error
	ds.db, err = bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	ds.prevDBStats = ds.db.Stats()
	ds.dbOpenTime = time.Now()
	err = ds.initDB()
	if err != nil {
		ds.db.Close()
		return err
	}
	err = ds.loadConfFromDB()
	if err != nil {
		ds.db.Close()
		return err
	}
	err = ds.loadMapDataFromDB()
	if err != nil {
		ds.db.Close()
		return err
	}
	_ = ds.loadPollingTemplateFromDB()
	return nil
}

func (ds *DataStore) initDB() error {
	buckets := []string{"config", "nodes", "lines", "pollings", "logs", "pollingLogs",
		"syslog", "trap", "netflow", "ipfix", "arp", "ai", "report", "pollingTemplates"}
	reports := []string{"devices", "users", "flows", "servers", "allows", "dennys"}
	ds.initConf()
	return ds.db.Update(func(tx *bbolt.Tx) error {
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
func (ds *DataStore) CloseDB() {
	if ds.db == nil {
		return
	}
	ds.db.Close()
	ds.db = nil
}

func (ds *DataStore) UpdateDBStats() {
	if ds.db == nil {
		return
	}
	s := ds.db.Stats()
	d := s.Sub(&ds.prevDBStats)
	var dbSize int64
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		dbSize = tx.Size()
		return nil
	})
	ds.DBStats.Size = dbSize
	ds.DBStats.TotalWrite = int64(s.TxStats.Write)
	ds.DBStats.LastWrite = int64(d.TxStats.Write)
	if ds.DBStats.PeakWrite < ds.DBStats.LastWrite {
		ds.DBStats.PeakWrite = ds.DBStats.LastWrite
	}
	// 初回は計算しない。
	if ds.DBStats.PeakWrite > 0 && ds.DBStats.Time != "" {
		ds.DBStats.Rate = 100 * float64(d.TxStats.Write) / float64(ds.DBStats.PeakWrite)
		ds.DBStats.StartTime = ds.dbOpenTime
		dbot := time.Since(ds.dbOpenTime).Seconds()
		if dbot > 0 {
			ds.DBStats.AvgWrite = float64(s.TxStats.Write) / dbot
		}
	}
	dt := d.TxStats.WriteTime.Seconds()
	if dt != 0 {
		ds.DBStats.Speed = float64(d.TxStats.Write) / dt
		if ds.DBStats.Peak < ds.DBStats.Speed {
			ds.DBStats.Peak = ds.DBStats.Speed
		}
	} else {
		ds.DBStats.Speed = 0.0
	}
	ds.DBStats.Time = time.Now().Format("15:04:05")
	ds.prevDBStats = s

	if ds.DBStats.BackupFile != "" && ds.nextBackup != 0 && ds.nextBackup < time.Now().UnixNano() {
		ds.nextBackup += (24 * 3600 * 1000 * 1000 * 1000)
		go func() {
			log.Printf("Backup start = %s", ds.DBStats.BackupFile)
			ds.AddEventLog(&EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ開始:" + ds.DBStats.BackupFile,
			})
			if err := ds.BackupDB(); err != nil {
				log.Printf("backupDB err=%v", err)
			}
			log.Printf("Backup end = %s", ds.DBStats.BackupFile)
			ds.AddEventLog(&EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ終了:" + ds.DBStats.BackupFile,
			})
		}()
		ds.DBStats.BackupTime = ds.DBStats.Time
	}
}

// bboltに保存する場合のキーを時刻から生成する。
func makeKey() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}

func (ds *DataStore) BackupDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if ds.dstDB != nil {
		return fmt.Errorf("backup in progress")
	}
	os.Remove(ds.DBStats.BackupFile)
	var err error
	ds.dstDB, err = bbolt.Open(ds.DBStats.BackupFile, 0600, nil)
	if err != nil {
		return err
	}
	defer func() {
		ds.dstDB.Close()
		ds.dstDB = nil
	}()
	ds.dstTx, err = ds.dstDB.Begin(true)
	if err != nil {
		return err
	}
	err = ds.db.View(func(srcTx *bbolt.Tx) error {
		return srcTx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			return ds.walkBucket(b, nil, name, nil, b.Sequence())
		})
	})
	if err != nil {
		_ = ds.dstTx.Rollback()
		return err
	}
	if !ds.DBStats.BackupConfigOnly {
		mapConfTmp := ds.MapConf
		mapConfTmp.EnableNetflowd = false
		mapConfTmp.EnableSyslogd = false
		mapConfTmp.EnableTrapd = false
		mapConfTmp.LogDays = 0
		if s, err := json.Marshal(mapConfTmp); err == nil {
			if b := ds.dstTx.Bucket([]byte("config")); b != nil {
				return b.Put([]byte("mapConf"), s)
			}
		}
	}
	return ds.dstTx.Commit()
}

var configBuckets = []string{"config", "nodes", "lines", "pollings", "mibdb"}

func (ds *DataStore) walkBucket(b *bbolt.Bucket, keypath [][]byte, k, v []byte, seq uint64) error {
	if ds.stopBackup {
		return fmt.Errorf("stop backup")
	}
	if ds.DBStats.BackupConfigOnly && v == nil {
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
	if ds.dbBackupSize > 64*1024 {
		_ = ds.dstTx.Commit()
		var err error
		ds.dstTx, err = ds.dstDB.Begin(true)
		if err != nil {
			return err
		}
		ds.dbBackupSize = 0
	}
	// Execute callback.
	if err := ds.walkFunc(keypath, k, v, seq); err != nil {
		return err
	}
	ds.dbBackupSize += int64(len(k) + len(v))

	// If this is not a bucket then stop.
	if v != nil {
		return nil
	}

	// Iterate over each child key/value.
	keypath = append(keypath, k)
	return b.ForEach(func(k, v []byte) error {
		if v == nil {
			bkt := b.Bucket(k)
			return ds.walkBucket(bkt, keypath, k, nil, bkt.Sequence())
		}
		return ds.walkBucket(b, keypath, k, v, b.Sequence())
	})
}

func (ds *DataStore) walkFunc(keys [][]byte, k, v []byte, seq uint64) error {
	// Create bucket on the root transaction if this is the first level.
	nk := len(keys)
	if nk == 0 {
		bkt, err := ds.dstTx.CreateBucket(k)
		if err != nil {
			return err
		}
		if err := bkt.SetSequence(seq); err != nil {
			return err
		}
		return nil
	}
	// Create buckets on subsequent levels, if necessary.
	b := ds.dstTx.Bucket(keys[0])
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
