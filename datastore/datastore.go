// Package datastore : データ保存
package datastore

import (
	"context"
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
	dspath      string
	prevDBStats bbolt.Stats
	dbOpenTime  time.Time
	// Conf Data on Memory
	MapConf      MapConfEnt
	NotifyConf   NotifyConfEnt
	InfluxdbConf InfluxdbConfEnt
	DiscoverConf DiscoverConfEnt
	Backup       DBBackupEnt
	DBStats      DBStatsEnt
	DBStatsLog   []DBStatsEnt
	// Map Data on Memory not export
	nodes    sync.Map
	lines    sync.Map
	pollings sync.Map
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

func NewDataStore(ctx context.Context, dspath string, fs http.FileSystem) *DataStore {
	ds := &DataStore{
		dspath:     dspath,
		devices:    make(map[string]*DeviceEnt),
		users:      make(map[string]*UserEnt),
		flows:      make(map[string]*FlowEnt),
		servers:    make(map[string]*ServerEnt),
		dennyRules: make(map[string]bool),
		allowRules: make(map[string]*AllowRuleEnt),
		eventLogCh: make(chan *EventLogEnt, 100),
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
	ds.InitDataStore(fs)
	go ds.eventLogger(ctx)
	return ds
}

func (ds *DataStore) InitDataStore(fs http.FileSystem) {
	if ds.dspath == "" {
		log.Println("No DataStore Path Skip Init")
		return
	}
	// BBoltをオープン
	if err := ds.OpenDB(filepath.Join(ds.dspath, "twsnmpfc.db")); err != nil {
		log.Fatalf("InitDataStore OpenDB err=%v", err)
	}
	// MIBDB
	if r, err := os.Open(filepath.Join(ds.dspath, "mib.txt")); err == nil {
		ds.loadMIBDB(r)
	} else {
		if r, err := fs.Open("/conf/mib.txt"); err == nil {
			ds.loadMIBDB(r)
		} else {
			log.Fatalf("InitDataStore MIBDB err=%v", err)
		}
	}
	// 拡張MIBの読み込み
	ds.loadExtMIBs(filepath.Join(ds.dspath, "extmibs"))
	// サービスの定義ファイル、ユーザー指定があれば利用、なければ内蔵
	if r, err := os.Open(filepath.Join(ds.dspath, "services.txt")); err == nil {
		ds.loadServiceMap(r)
	} else {
		if r, err := fs.Open("/conf/services.txt"); err == nil {
			ds.loadServiceMap(r)
		} else {
			log.Fatalf("InitDataStore services.txt err=%v", err)
		}
	}
	// OUIの定義
	if r, err := os.Open(filepath.Join(ds.dspath, "oui.txt")); err == nil {
		ds.loadOUIMap(r)
	} else {
		if r, err := fs.Open("/conf/oui.txt"); err == nil {
			ds.loadOUIMap(r)
		} else {
			log.Fatalf("InitDataStore oui.txt err=%v", err)
		}
	}
	// TLS暗号名の定義
	if r, err := os.Open(filepath.Join(ds.dspath, "tlsparams.csv")); err == nil {
		ds.loadTLSCihperNameMap(r)
	} else {
		if r, err := fs.Open("/conf/tlsparams.csv"); err == nil {
			ds.loadTLSCihperNameMap(r)
		} else {
			log.Fatalf("InitDataStore tlsparams.csv err=%v", err)
		}
	}
	p := filepath.Join(ds.dspath, "geoip.mmdb")
	if _, err := os.Stat(p); err == nil {
		ds.openGeoIP(p)
	}
	p = filepath.Join(ds.dspath, "grok.txt")
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
	return nil
}

func (ds *DataStore) initDB() error {
	buckets := []string{"config", "nodes", "lines", "pollings", "logs", "pollingLogs",
		"syslog", "trap", "netflow", "ipfix", "arp", "ai", "report"}
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

// bboltに保存する場合のキーを時刻から生成する。
func makeKey() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}
