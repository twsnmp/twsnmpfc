// Package datastore : データ保存
package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod

	"go.etcd.io/bbolt"
)

type DBBackupEnt struct {
	ConfigOnly bool
	Daily      bool
	Generation int
}

func (ds *DataStore) SaveBackupParamToDB() error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ds.Backup)
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

func (ds *DataStore) CheckDBBackup() {
	if ds.db == nil {
		return
	}
	if ds.Backup.Daily && ds.nextBackup == 0 {
		now := time.Now()
		d := 0
		if now.Hour() > 2 {
			d = 1
		}
		ds.nextBackup = time.Date(now.Year(), now.Month(), now.Day()+d, 3, 0, 0, 0, time.Local).UnixNano()
	}
	if err := os.MkdirAll(filepath.Join(ds.dspath, "backup"), 0666); err != nil {
		return
	}
	file := filepath.Join(ds.dspath, "backup", "twsnmpfs.db."+time.Now().Format("20060102150405"))
	if ds.nextBackup != 0 && ds.nextBackup < time.Now().UnixNano() {
		if ds.Backup.Daily {
			ds.nextBackup += (24 * 3600 * 1000 * 1000 * 1000)
		} else {
			ds.nextBackup = 0
		}
		go func() {
			log.Printf("Backup start = %s", file)
			ds.AddEventLog(&EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ開始:" + file,
			})
			if err := ds.BackupDB(file); err != nil {
				log.Printf("backupDB err=%v", err)
				ds.AddEventLog(&EventLogEnt{
					Type:  "system",
					Level: "error",
					Event: "バックアップ失敗:" + file,
				})
				return
			}
			log.Printf("Backup end = %s", file)
			ds.AddEventLog(&EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ終了:" + file,
			})
			ds.DBStats.BackupTime = time.Now().UnixNano()
		}()
	}
}

func (ds *DataStore) BackupDB(file string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	if ds.dstDB != nil {
		return fmt.Errorf("backup in progress")
	}
	os.Remove(file)
	var err error
	ds.dstDB, err = bbolt.Open(file, 0600, nil)
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
	if !ds.Backup.ConfigOnly {
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
	if ds.Backup.ConfigOnly && v == nil {
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
