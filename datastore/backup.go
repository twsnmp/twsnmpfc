// Package datastore : データ保存
package datastore

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod

	"go.etcd.io/bbolt"
)

type DBBackupEnt struct {
	Mode       string
	ConfigOnly bool
	Generation int
}

func SaveBackup() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(Backup)
	if err != nil {
		return err
	}
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveBackup dur=%v", time.Since(st))
		return b.Put([]byte("backup"), s)
	})
}

func CheckDBBackup() {
	log.Printf("check backup mode='%s' nextBackup=%v", Backup.Mode, time.Unix(0, nextBackup))
	if db == nil || Backup.Mode == "" {
		return
	}
	if DBStats.BackupStart > 0 {
		log.Printf("skip backup mode=%s ", Backup.Mode)
		return
	}
	if Backup.Mode == "daily" && nextBackup == 0 {
		now := time.Now()
		d := 0
		if now.Hour() > 2 {
			d = 1
		}
		nextBackup = time.Date(now.Year(), now.Month(), now.Day()+d, 3, 0, 0, 0, time.Local).UnixNano()
	} else if Backup.Mode == "onece" {
		// 1回だけはすぐ実行するように修正
		nextBackup = 0
	}
	if BackupPath == "" {
		if err := os.MkdirAll(filepath.Join(dspath, "backup"), 0777); err != nil {
			log.Printf("backup err=%v", err)
			return
		}
	}
	if nextBackup < time.Now().UnixNano() {
		if Backup.Mode == "daily" {
			nextBackup += (24 * 3600 * 1000 * 1000 * 1000)
		} else {
			Backup.Mode = ""
			nextBackup = 0
			SaveBackup()
		}
		DBStats.BackupStart = time.Now().UnixNano()
		go func() {
			log.Printf("start backup")
			st := time.Now()
			file := filepath.Join(dspath, "backup", "twsnmpfc.db."+time.Now().Format("20060102150405"))
			if BackupPath != "" {
				file = filepath.Join(BackupPath, "twsnmpfc.db."+time.Now().Format("20060102150405"))
			}
			stopBackup = false
			defer func() {
				DBStats.BackupStart = 0
			}()
			AddEventLog(&EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ開始:" + file,
			})
			if err := backupDB(file); err != nil {
				log.Printf("backup err=%v", err)
				os.RemoveAll(file)
				AddEventLog(&EventLogEnt{
					Type:  "system",
					Level: "warn",
					Event: "バックアップ失敗:" + file,
				})
				return
			}
			log.Printf("backup file=%s dur=%v", file, time.Since(st))
			AddEventLog(&EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ終了:" + file,
			})
			DBStats.BackupTime = time.Now().UnixNano()
			rotateBackup()
		}()
	}
}

func StopBackup() {
	stopBackup = true
}

func backupDB(file string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if dstDB != nil {
		return fmt.Errorf("backup in progress")
	}
	os.Remove(file)
	if CopyBackup && !Backup.ConfigOnly {
		log.Println("backup copyfile mode")
		return db.View(func(tx *bbolt.Tx) error {
			return tx.CopyFile(file, 0600)
		})
	}
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
				if err := b.Put([]byte("mapConf"), s); err != nil {
					_ = dstTx.Rollback()
					return err
				}
			}
		}
	}
	return dstTx.Commit()
}

var configBuckets = map[string]bool{
	"config":   true,
	"nodes":    true,
	"lines":    true,
	"items":    true,
	"pollings": true,
	"grok":     true,
	"images":   true,
}

func walkBucket(b *bbolt.Bucket, keypath [][]byte, k, v []byte, seq uint64) error {
	if stopBackup {
		return fmt.Errorf("stop backup")
	}
	if Backup.ConfigOnly && v == nil {
		if _, ok := configBuckets[string(k)]; !ok {
			log.Printf("skip backup %s", string(k))
			return nil
		}
		log.Printf("do backup %s", string(k))
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

func rotateBackup() {
	if Backup.Mode != "daily" {
		return
	}
	dirList, err := os.ReadDir(filepath.Join(dspath, "backup"))
	if err != nil {
		log.Printf("rotate backup err=%v", err)
		return
	}
	backupList := []fs.FileInfo{}
	for _, d := range dirList {
		if d.IsDir() || !strings.HasPrefix(d.Name(), "twsnmpfc.db.") {
			continue
		}
		if f, err := d.Info(); err == nil {
			backupList = append(backupList, f)
		}
	}
	if Backup.Generation+1 >= len(backupList) {
		return
	}
	sort.Slice(backupList, func(i, j int) bool {
		return backupList[i].ModTime().Before(backupList[j].ModTime())
	})
	for i := 0; i < len(backupList)-(Backup.Generation+1); i++ {
		backup := filepath.Join(dspath, "backup", backupList[i].Name())
		if err := os.Remove(backup); err != nil {
			log.Printf("delete backup file=%s err=%v", backup, err)
		} else {
			log.Printf("delete backup file=%s", backup)
		}
	}
}

func RestoreDB(ds, backup string) error {
	if BackupPath != "" {
		return restoreDBWithPath(ds, backup)
	}
	srcBackup := filepath.Join(ds, "backup", backup)
	if _, err := os.Stat(srcBackup); err != nil {
		return err
	}
	dbPath := filepath.Join(ds, "twsnmpfc.db")
	newBackup := filepath.Join(ds, "backup", "twsnmpfc.db."+time.Now().Format("20060102150405"))
	if err := os.Rename(dbPath, newBackup); err != nil {
		return err
	}
	os.RemoveAll(dbPath)
	if err := os.Rename(srcBackup, dbPath); err != nil {
		return err
	}
	return nil
}

func restoreDBWithPath(ds, backup string) error {
	srcBackup := filepath.Join(BackupPath, backup)
	if _, err := os.Stat(srcBackup); err != nil {
		return err
	}
	dbPath := filepath.Join(ds, "twsnmpfc.db")
	newBackup := filepath.Join(BackupPath, "twsnmpfc.db."+time.Now().Format("20060102150405"))
	if err := copyFile(dbPath, newBackup); err != nil {
		return err
	}
	os.RemoveAll(dbPath)
	if err := copyFile(srcBackup, dbPath); err != nil {
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	d, err := os.Create(dst)
	if err != nil {
		s.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	s.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	return nil
}

func CompactDB(ds, compact string) error {
	dbPath := filepath.Join(ds, "twsnmpfc.db")
	fi, err := os.Stat(dbPath)
	if err != nil {
		return err
	}
	src, err := bbolt.Open(dbPath, 0444, &bbolt.Options{ReadOnly: true})
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := bbolt.Open(compact, fi.Mode(), nil)
	if err != nil {
		return err
	}
	defer dst.Close()
	return bbolt.Compact(dst, src, 1024*1024*64)
}

// 再起動後にも最終バックアップ時刻を表示するため
func setLastBackupTime() {
	path := BackupPath
	if path == "" {
		path = filepath.Join(dspath, "backup")
	}
	dirList, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, d := range dirList {
		if d.IsDir() || !strings.HasPrefix(d.Name(), "twsnmpfc.db.") {
			continue
		}
		if f, err := d.Info(); err == nil {
			if DBStats.BackupTime < f.ModTime().UnixNano() {
				DBStats.BackupTime = f.ModTime().UnixNano()
			}
		}
	}
}
