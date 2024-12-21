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
	"strconv"
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

func deleteOldLog(tx *bbolt.Tx, bucket string, days int) (bool, int) {
	s := time.Now()
	done := true
	delCount := 0
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		log.Printf("bucket %s not found", bucket)
		// bucketがないのは、エラーにしないでスキップする
		return done, 0
	}
	var lt time.Time
	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		if st < string(k) {
			break
		}
		if delCount%1000 == 0 {
			if time.Now().UnixMilli()-s.UnixMilli() > 500 {
				if n, err := strconv.ParseInt(string(k), 16, 64); err == nil {
					lt = time.Unix(0, n)
				}
				done = false
				break
			}
		}
		c.Delete()
		delCount++
	}
	if delCount > 0 {
		td := ""
		if !done {
			if n, err := strconv.ParseInt(st, 16, 64); err == nil {
				t := time.Unix(0, n)
				td = "td=" + t.Sub(lt).String()
			}
		}
		log.Printf("delete old logs bucket=%s count=%d done=%v dur=%s %s",
			bucket, delCount, done, time.Since(s), td)
	}
	return done, delCount
}

// deleteOldPollingLogは、古いポーリングログを削除する
func deleteOldPollingLog(tx *bbolt.Tx, days int) int {
	s := time.Now()
	delCount := 0
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	b := tx.Bucket([]byte("pollingLogs"))
	if b == nil {
		log.Println("bucket pollingLogs not found")
		// bucketがないのは、エラーにしないでスキップする
		return delCount
	}
	b.ForEachBucket(func(k []byte) error {
		b2 := b.Bucket(k)
		if b2 == nil {
			return nil
		}
		c := b2.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if st < string(k) {
				break
			}
			_ = c.Delete()
			delCount++
		}
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
	tx, err := db.Begin(true)
	if err != nil {
		log.Printf("deleteOldLog err=%v", err)
		return
	}
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix", "arplog", "sflow", "sflowCounter"}
	doneMap := make(map[string]bool)
	doneCount := 0
	delCount := 0
	lt := time.Now().Unix() + 50
	for doneCount < len(buckets) && lt > time.Now().Unix() {
		for _, b := range buckets {
			if _, ok := doneMap[b]; !ok {
				if b == "pollingLogs" {
					doneMap[b] = true
					doneCount++
					delCount += deleteOldPollingLog(tx, MapConf.LogDays)
				} else {
					done, c := deleteOldLog(tx, b, MapConf.LogDays)
					delCount += c
					if done {
						doneMap[b] = true
						doneCount++
					}
				}
			}
			tx.Commit()
			tx, err = db.Begin(true)
			if err != nil {
				log.Printf("deleteOldLog err=%v", err)
				return
			}
		}
	}
	tx.Commit()
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
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -MapConf.LogDays).UnixNano())
	s := time.Now()
	d.Batch(func(tx *bbolt.Tx) error {
		switch t {
		case "all":
			delCount += deleteOldPollingLog(tx, MapConf.LogDays)
			buckets := []string{"logs", "syslog", "trap", "netflow", "ipfix", "sflow", "sflowCounter"}
			for _, bn := range buckets {
				b := tx.Bucket([]byte(bn))
				if b != nil {
					log.Printf("start delete %s", bn)
					del := deleteOldLogBatchSub(st, b)
					delCount += del
					log.Printf("end delete %s count=%d", bn, del)
				}
			}
		case "polling":
			delCount += deleteOldPollingLog(tx, MapConf.LogDays)
		default:
			b := tx.Bucket([]byte(t))
			if b != nil {
				delCount += deleteOldLogBatchSub(st, b)
			}
		}
		return nil
	})
	log.Printf("delete old log count=%d,dur=%v", delCount, time.Since(s))
}

func deleteOldLogBatchSub(st string, b *bbolt.Bucket) int {
	delCount := 0
	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		if st < string(k) {
			break
		}
		c.Delete()
		delCount++
		if delCount%10000 == 0 {
			fmt.Printf("delete=%d\r", delCount)
		}
	}
	return delCount
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
