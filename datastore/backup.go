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
	Mode       string
	ConfigOnly bool
	Generation int
}

func SaveBackupToDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(Backup)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("backup"), s)
	})
}

func CheckDBBackup() {
	if db == nil || Backup.Mode == "" {
		return
	}
	if Backup.Mode == "daily" && nextBackup == 0 {
		now := time.Now()
		d := 0
		if now.Hour() > 2 {
			d = 1
		}
		nextBackup = time.Date(now.Year(), now.Month(), now.Day()+d, 3, 0, 0, 0, time.Local).UnixNano()
	}
	if err := os.MkdirAll(filepath.Join(dspath, "backup"), 0666); err != nil {
		return
	}
	file := filepath.Join(dspath, "backup", "twsnmpfs.db."+time.Now().Format("20060102150405"))
	if nextBackup != 0 && nextBackup < time.Now().UnixNano() {
		if Backup.Mode == "daily" {
			nextBackup += (24 * 3600 * 1000 * 1000 * 1000)
		} else {
			Backup.Mode = ""
			nextBackup = 0
			SaveBackupToDB()
		}
		go func() {
			log.Printf("Backup start = %s", file)
			AddEventLog(&EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ開始:" + file,
			})
			if err := BackupDB(file); err != nil {
				log.Printf("backupDB err=%v", err)
				AddEventLog(&EventLogEnt{
					Type:  "system",
					Level: "error",
					Event: "バックアップ失敗:" + file,
				})
				return
			}
			log.Printf("Backup end = %s", file)
			AddEventLog(&EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ終了:" + file,
			})
			DBStats.BackupTime = time.Now().UnixNano()
		}()
	}
}

func BackupDB(file string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if dstDB != nil {
		return fmt.Errorf("backup in progress")
	}
	os.Remove(file)
	var err error
	dstDB, err = bbolt.Open(file, 0600, nil)
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
	if !Backup.ConfigOnly {
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
	if Backup.ConfigOnly && v == nil {
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
