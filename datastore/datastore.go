// Package datastore : データ保存
package datastore

import (
	"context"
	"fmt"
	"io/ioutil"
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

var (
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
	DBStatsLog   []DBStatsLogEnt
	Yasumi       string
	// Restrt snmptrapd
	RestartSnmpTrapd bool
	// Map Data on Memory
	nodes    sync.Map
	lines    sync.Map
	pollings sync.Map
	// Report Data on Memory
	devices sync.Map
	users   sync.Map
	flows   sync.Map
	servers sync.Map
	ips     sync.Map
	// TWPCAP
	etherType   sync.Map
	radiusFlows sync.Map
	tlsFlows    sync.Map
	dnsq        sync.Map
	certs       sync.Map
	sensors     sync.Map
	// TWWINLOG
	winEventID   sync.Map
	winLogon     sync.Map
	winAccount   sync.Map
	winKerberos  sync.Map
	winPrivilege sync.Map
	winProcess   sync.Map
	winTask      sync.Map
	// twBlueScan
	blueDevice sync.Map
	envMonitor sync.Map
	// twWifiScan
	wifiAP sync.Map
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

func Init(ctx context.Context, path string, fs http.FileSystem, wg *sync.WaitGroup) error {
	dspath = path
	eventLogCh = make(chan *EventLogEnt, 100)
	protMap = map[int]string{
		1:   "icmp",
		2:   "igmp",
		6:   "tcp",
		8:   "egp",
		17:  "udp",
		112: "vrrp",
	}
	serviceMap = make(map[string]string)
	geoipMap = make(map[string]string)
	ouiMap = make(map[string]string)
	tlsCSMap = make(map[string]string)
	if err := loadDataFromFS(fs); err != nil {
		return err
	}
	wg.Add(1)
	go eventLogger(ctx, wg)
	setLastBackupTime()
	return nil
}

func loadDataFromFS(fs http.FileSystem) error {
	if dspath == "" {
		return fmt.Errorf("no data base path")
	}
	// BBoltをオープン
	if err := openDB(filepath.Join(dspath, "twsnmpfc.db")); err != nil {
		return err
	}
	// MIBDB
	if r, err := os.Open(filepath.Join(dspath, "mib.txt")); err == nil {
		loadMIBDB(r)
	} else {
		if r, err := fs.Open("/conf/mib.txt"); err == nil {
			loadMIBDB(r)
		} else {
			return err
		}
	}
	// 拡張MIBの読み込み
	loadExtMIBs(filepath.Join(dspath, "extmibs"))
	// サービスの定義ファイル、ユーザー指定があれば利用、なければ内蔵
	if r, err := os.Open(filepath.Join(dspath, "services.txt")); err == nil {
		loadServiceMap(r)
	} else {
		if r, err := fs.Open("/conf/services.txt"); err == nil {
			loadServiceMap(r)
		} else {
			return err
		}
	}
	makeMibTreeList()
	// OUIの定義
	if r, err := os.Open(filepath.Join(dspath, "mac-vendors-export.csv")); err == nil {
		loadOUIMap(r)
	} else {
		if r, err := fs.Open("/conf/mac-vendors-export.csv"); err == nil {
			loadOUIMap(r)
		} else {
			return err
		}
	}
	// 休みの定義
	if r, err := os.Open(filepath.Join(dspath, "yasumi.txt")); err == nil {
		if b, err := ioutil.ReadAll(r); err == nil && len(b) > 0 {
			Yasumi = string(b)
		}
		r.Close()
	}
	if Yasumi == "" {
		if r, err := fs.Open("/conf/yasumi.txt"); err == nil {
			if b, err := ioutil.ReadAll(r); err == nil && len(b) > 0 {
				Yasumi = string(b)
			}
			r.Close()
		} else {
			log.Printf("open yasumi.txt err=%v", err)
		}
	}
	// TLS暗号名の定義
	if r, err := os.Open(filepath.Join(dspath, "tlsparams.csv")); err == nil {
		loadTLSCihperNameMap(r)
	} else {
		if r, err := fs.Open("/conf/tlsparams.csv"); err == nil {
			loadTLSCihperNameMap(r)
		} else {
			return err
		}
	}
	p := filepath.Join(dspath, "geoip.mmdb")
	if _, err := os.Stat(p); err == nil {
		openGeoIP(p)
	}
	loadGrokMap()
	if r, err := fs.Open("/conf/polling.json"); err == nil {
		if b, err := ioutil.ReadAll(r); err == nil && len(b) > 0 {
			if err := loadPollingTemplate(b); err != nil {
				log.Printf("load polling template err=%v", err)
			}
		}
		r.Close()
	} else {
		log.Printf("open polling template err=%v", err)
	}
	if r, err := os.Open(filepath.Join(dspath, "polling.json")); err == nil {
		if b, err := ioutil.ReadAll(r); err == nil && len(b) > 0 {
			if err := loadPollingTemplate(b); err != nil {
				log.Printf("load polling template err=%v", err)
			}
		}
		r.Close()
	}
	return nil
}

func openDB(path string) error {
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
	err = loadConf()
	if err != nil {
		db.Close()
		return err
	}
	err = loadMapData()
	if err != nil {
		db.Close()
		return err
	}
	err = setupInfluxdb()
	if err != nil {
		log.Printf("setup influxdb err=%v", err)
	}
	return nil
}

func initDB() error {
	buckets := []string{"config", "nodes", "lines", "pollings", "logs", "pollingLogs",
		"syslog", "trap", "netflow", "ipfix", "arplog", "arp", "ai", "report", "grok"}
	reports := []string{"devices", "users", "flows", "servers", "ips",
		"ether", "dns", "radius", "tls", "cert",
		"sensor",
		"winEventID", "winLogon", "winAccount", "winKerberos",
		"winPrivilege", "winProcess", "winTask",
		"wifiAP", "blueDevice", "envMonitor",
	}
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

// CloseDataStore : DBをクローズする
func CloseDataStore() {
	if db == nil {
		return
	}
	db.Close()
	db = nil
}

// bboltに保存する場合のキーを時刻から生成する。
func makeKey() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}

// Data Storeのパスを返す、何かと必要なので
func GetDataStorePath() string {
	return dspath
}
