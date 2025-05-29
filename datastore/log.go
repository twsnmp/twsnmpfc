package datastore

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

const (
	LogModeNone = iota
	LogModeAlways
	LogModeOnChange
	LogModeAI
)

type EventLogEnt struct {
	Time     int64 // UnixNano()
	Type     string
	Level    string
	NodeName string
	NodeID   string
	Event    string
}

type LogEnt struct {
	Time int64 // UnixNano()
	Type string
	Log  string
}

type LogFilterEnt struct {
	StartTime string
	EndTime   string
	Filter    string
	LogType   string
}

type SFlowCounterEnt struct {
	Remote string
	Type   string
	Data   string
}

func AddEventLog(e *EventLogEnt) {
	e.Time = time.Now().UnixNano()
	if e.NodeID != "" && e.NodeName == "" {
		// Node IDのみの場合は、名前をここで解決する
		if n := GetNode(e.NodeID); n != nil {
			e.NodeName = n.Name
		}
	}
	eventLogCh <- e
}

func ForEachEventLog(st, et int64, f func(*EventLogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x", st)
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Seek([]byte(sk)); k != nil; k, v = c.Next() {
			var e EventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				continue
			}
			if e.Time < st {
				continue
			}
			if e.Time > et {
				break
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

func ForEachLastEventLog(last int64, f func(*EventLogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var e EventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				continue
			}
			if e.Time < last {
				break
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

func ForEachLastLog(t string, f func(*LogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(t))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if bytes.HasSuffix(v, []byte{0, 0, 255, 255}) {
				v = deCompressLog(v)
			}
			var e LogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				continue
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

func ForEachLog(st, et int64, t string, f func(*LogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x", st)
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(t))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Seek([]byte(sk)); k != nil; k, v = c.Next() {
			if bytes.HasSuffix(v, []byte{0, 0, 255, 255}) {
				v = deCompressLog(v)
			}
			var e LogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Printf("ForEachLog v=%s err=%v", v, err)
				continue
			}
			if e.Time < st {
				continue
			}
			if e.Time > et {
				break
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

func deleteOldLog(bucket string, days int) (bool, int) {
	s := time.Now()
	done := true
	delCount := 0
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			log.Printf("bucket %s not found", bucket)
			// bucketがないのは、エラーにしないでスキップする
			return nil
		}
		c := b.Cursor()
		delList := [][]byte{}
		for k, _ := c.First(); k != nil && st > string(k); k, _ = c.Next() {
			delList = append(delList, k)
			if len(delList) > 20000 {
				done = false
				break
			}
		}
		for _, k := range delList {
			b.Delete(k)
		}
		delCount = len(delList)
		return nil
	})
	if delCount > 0 {
		log.Printf("delete old logs bucket=%s count=%d done=%v dur=%s", bucket, delCount, done, time.Since(s))
	}
	return done, delCount
}

// deleteOldPollingLogは、古いポーリングログを削除する
func deleteOldPollingLog(days int) int {
	s := time.Now()
	delCount := 0
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			log.Println("bucket pollingLogs not found")
			// bucketがないのは、エラーにしないでスキップする
			return nil
		}
		b.ForEachBucket(func(k []byte) error {
			b2 := b.Bucket(k)
			if b2 == nil {
				return nil
			}
			c := b2.Cursor()
			delList := [][]byte{}
			for k2, _ := c.First(); k2 != nil && st > string(k2); k2, _ = c.Next() {
				delList = append(delList, k2)
			}
			for _, k2 := range delList {
				b2.Delete(k2)
			}
			delCount += len(delList)
			return nil
		})
		return nil
	})
	if delCount > 0 {
		log.Printf("delete old polling logs count=%d dur=%s", delCount, time.Since(s))
	}
	return delCount
}

func deleteOldLogs() {
	s := time.Now()
	if MapConf.LogDays < 1 {
		log.Println("mapConf.LogDays < 1 ")
		return
	}
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix", "arplog", "sflow", "sflowCounter"}
	doneMap := make(map[string]bool)
	doneCount := 0
	delCount := 0
	lt := time.Now().Unix() + 55
	for doneCount < len(buckets) && lt > time.Now().Unix() {
		for _, b := range buckets {
			if _, ok := doneMap[b]; !ok {
				if b == "pollingLogs" {
					doneMap[b] = true
					doneCount++
					delCount += deleteOldPollingLog(MapConf.LogDays)
				} else {
					done, c := deleteOldLog(b, MapConf.LogDays)
					delCount += c
					if done {
						doneMap[b] = true
						doneCount++
					}
				}
			}
		}
	}
	log.Printf("deleteOldLogs delLogs=%d done=%d dur=%s", delCount, doneCount, time.Since(s))
}

func DeleteAllLogs() {
	st := time.Now()
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix", "sflow", "sflowCounter"}
	for _, b := range buckets {
		db.Batch(func(tx *bbolt.Tx) error {
			if err := tx.DeleteBucket([]byte(b)); err != nil {
				return err
			}
			tx.CreateBucketIfNotExists([]byte(b))
			return nil
		})
	}
	log.Printf("DeleteAllLogs dur=%v", time.Since(st))
}

func DeleteLogs(b string) {
	buckets := []string{"logs", "syslog", "trap", "netflow", "ipfix", "sflow", "sflowCounter"}
	st := time.Now()
	for _, bb := range buckets {
		if bb == b {
			db.Batch(func(tx *bbolt.Tx) error {
				if err := tx.DeleteBucket([]byte(b)); err != nil {
					return err
				}
				tx.CreateBucketIfNotExists([]byte(b))
				return nil
			})
			log.Printf("DeleteLogs bucket=%s dur=%v", b, time.Since(st))
			return
		}
	}
	log.Println("DeleteLogs no bucket")
}

func DeleteArp() {
	st := time.Now()
	buckets := []string{"arp", "arplog"}
	for _, b := range buckets {
		db.Batch(func(tx *bbolt.Tx) error {
			if err := tx.DeleteBucket([]byte(b)); err != nil {
				return err
			}
			tx.CreateBucketIfNotExists([]byte(b))
			return nil
		})
	}
	log.Printf("DeleteArp dur=%v", time.Since(st))
}

func eventLogger(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start eventlog")
	timer := time.NewTicker(time.Second * 10)
	eventLogList := []*EventLogEnt{}
	pollingLogList := []*PollingLogEnt{}
	for {
		select {
		case <-ctx.Done():
			if len(eventLogList) > 0 {
				saveLogList(eventLogList)
			}
			if len(pollingLogList) > 0 {
				savePollingLogList(pollingLogList)
			}
			timer.Stop()
			log.Println("stop eventlog")
			return
		case e := <-eventLogCh:
			eventLogList = append(eventLogList, e)
		case e := <-pollingLogCh:
			pollingLogList = append(pollingLogList, e)
		case <-timer.C:
			if len(eventLogList) > 0 {
				saveLogList(eventLogList)
				eventLogList = []*EventLogEnt{}
			}
			if len(pollingLogList) > 0 {
				savePollingLogList(pollingLogList)
				pollingLogList = []*PollingLogEnt{}
			}
		}
	}
}

func oldLogChecker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start old log checker")
	timer := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop old log checker")
			return
		case <-timer.C:
			deleteOldLogs()
			chekOldOTelData()
		}
	}
}

func saveLogList(list []*EventLogEnt) {
	if db == nil {
		return
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		for i, e := range list {
			s, err := json.Marshal(e)
			if err != nil {
				return err
			}
			err = b.Put([]byte(fmt.Sprintf("%016x", e.Time+int64(i))), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
	log.Printf("save event log count=%d,dur=%v", len(list), time.Since(st))
}

func savePollingLogList(list []*PollingLogEnt) {
	if db == nil {
		return
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		for i, e := range list {
			s, err := json.Marshal(e)
			if err != nil {
				return err
			}
			bs, err := b.CreateBucketIfNotExists([]byte(e.PollingID))
			if err != nil {
				return err
			}
			err = bs.Put([]byte(fmt.Sprintf("%016x", e.Time+int64(i))), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
	log.Printf("save polling log count=%d,dur=%v", len(list), time.Since(st))
}

func SaveLogBuffer(logBuffer []*LogEnt) {
	if db == nil {
		return
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		if time.Since(st) > time.Duration(time.Second) {
			log.Printf("SaveLogBuffer batch over 1sec dur=%v", time.Since(st))
		}
		syslog := tx.Bucket([]byte("syslog"))
		netflow := tx.Bucket([]byte("netflow"))
		sflow := tx.Bucket([]byte("sflow"))
		sflowCounter := tx.Bucket([]byte("sflowCounter"))
		ipfix := tx.Bucket([]byte("ipfix"))
		trap := tx.Bucket([]byte("trap"))
		arp := tx.Bucket([]byte("arplog"))
		sc := 0
		nfc := 0
		tc := 0
		ac := 0
		oc := 0
		sf := 0
		for i, l := range logBuffer {
			k := fmt.Sprintf("%016x", l.Time+int64(i))
			s, err := json.Marshal(l)
			if err != nil {
				return err
			}
			logSize += int64(len(s))
			if len(s) > 100 {
				s = compressLog(s)
			}
			compLogSize += int64(len(s))
			switch l.Type {
			case "syslog":
				sc++
				syslog.Put([]byte(k), []byte(s))
			case "netflow":
				nfc++
				netflow.Put([]byte(k), []byte(s))
			case "ipfix":
				nfc++
				ipfix.Put([]byte(k), []byte(s))
			case "trap":
				tc++
				trap.Put([]byte(k), []byte(s))
			case "arplog":
				ac++
				arp.Put([]byte(k), []byte(s))
			case "sflow":
				sf++
				sflow.Put([]byte(k), []byte(s))
			case "sflowCounter":
				sf++
				sflowCounter.Put([]byte(k), []byte(s))
			default:
				oc++
			}
		}
		log.Printf("syslog=%d,netflow=%d,trap=%d,arplog=%d,sflow=%d,other=%d,dur=%v", sc, nfc, tc, ac, sf, oc, time.Since(st))
		return nil
	})
}

func compressLog(s []byte) []byte {
	var b bytes.Buffer
	f, _ := flate.NewWriter(&b, flate.DefaultCompression)
	if _, err := f.Write(s); err != nil {
		return s
	}
	if err := f.Flush(); err != nil {
		return s
	}
	if err := f.Close(); err != nil {
		return s
	}
	return b.Bytes()
}

func deCompressLog(s []byte) []byte {
	r := flate.NewReader(bytes.NewBuffer(s))
	d, err := io.ReadAll(r)
	if err != nil {
		return s
	}
	return d
}

// DeleteOldLogBatch : 古いログのバッチ削除
func DeleteOldLogBatch(ds, t string) {
	openDB(ds)
	dbPath := filepath.Join(ds, "twsnmpfc.db")
	_, err := os.Stat(dbPath)
	if err != nil {
		log.Fatalln("no db")
	}
	d, err := bbolt.Open(dbPath, 0444, nil)
	if err != nil {
		log.Fatalf("db open err=%v", err)
	}
	defer d.Close()
	err = d.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b != nil {
			v := b.Get([]byte("mapConf"))
			if v != nil {
				if err := json.Unmarshal(v, &MapConf); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("read map conf err=%v", err)
	}
	if MapConf.LogDays < 1 {
		log.Fatalf("log days(%d) < 1 ", MapConf.LogDays)
	}
	delCount := 0
	s := time.Now()
	switch t {
	case "all":
		delCount += deleteOldPollingLog(MapConf.LogDays)
		buckets := []string{"logs", "syslog", "trap", "netflow", "ipfix", "sflow", "sflowCounter"}
		for _, b := range buckets {
			log.Printf("start delete %s", b)
			del := 0
			for {
				done, c := deleteOldLog(b, MapConf.LogDays)
				del += c
				if done {
					break
				}
			}
			delCount += del
			log.Printf("end delete %s count=%d", b, del)
		}
	case "polling":
		delCount += deleteOldPollingLog(MapConf.LogDays)
	case "logs", "syslog", "trap", "netflow", "ipfix", "sflow", "sflowCounter":
		log.Printf("start delete %s log", t)
		for {
			done, c := deleteOldLog(t, MapConf.LogDays)
			delCount += c
			if done {
				break
			}
		}
	default:
		log.Fatalf("log type %s not found", t)
	}
	log.Printf("delete old log count=%d,dur=%v", delCount, time.Since(s))
}

// ClearAllLogOnDB : コマンドからDBをオープンしてログとレーポートをすべて削除します。
func ClearAllLogOnDB(ds string) error {
	dbPath := filepath.Join(ds, "twsnmpfc.db")
	_, err := os.Stat(dbPath)
	if err != nil {
		return err
	}
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix", "sflow", "sflowCounter", "report"}
	for _, b := range buckets {
		db.Batch(func(tx *bbolt.Tx) error {
			if err := tx.DeleteBucket([]byte(b)); err != nil {
				log.Printf("ClearAllLogOnDB Delete bucket err=%v", err)
				return nil
			}
			tx.CreateBucketIfNotExists([]byte(b))
			return nil
		})
	}
	return nil
}
