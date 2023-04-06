package datastore

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

func ForEachLastEventLog(skey string, f func(*EventLogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil && string(k) != skey; k, v = c.Prev() {
			var e EventLogEnt
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

func deleteOldLog(tx *bbolt.Tx, bucket string, days int) error {
	s := time.Now()
	lt := s.Unix() + 10 // 10秒間削除
	delCount := 0
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s not found", bucket)
	}
	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		if st < string(k) || lt < time.Now().Unix() {
			break
		}
		_ = c.Delete()
		delCount++
	}
	if delCount > 0 {
		log.Printf("delete old logs bucket=%s count=%d dur=%s", bucket, delCount, time.Since(s))
	}
	return nil
}

func deleteOldLogs() {
	if MapConf.LogDays < 1 {
		log.Println("mapConf.LogDays < 1 ")
		return
	}
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix", "arplog"}
	db.Batch(func(tx *bbolt.Tx) error {
		for _, b := range buckets {
			if err := deleteOldLog(tx, b, MapConf.LogDays); err != nil {
				log.Printf("deleteOldLog bucket=%s err=%v", b, err)
			}
		}
		return nil
	})
}

func DeleteAllLogs() {
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix"}
	for _, b := range buckets {
		db.Batch(func(tx *bbolt.Tx) error {
			if err := tx.DeleteBucket([]byte(b)); err != nil {
				return err
			}
			tx.CreateBucketIfNotExists([]byte(b))
			return nil
		})
	}
}

func DeleteArp() {
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
}

func eventLogger(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start eventlog")
	timer1 := time.NewTicker(time.Minute)
	timer2 := time.NewTicker(time.Second * 5)
	list := []*EventLogEnt{}
	for {
		select {
		case <-ctx.Done():
			if len(list) > 0 {
				saveLogList(list)
			}
			timer1.Stop()
			timer2.Stop()
			log.Println("stop eventlog")
			return
		case e := <-eventLogCh:
			list = append(list, e)
			if len(list) > 100 {
				saveLogList(list)
				list = []*EventLogEnt{}
			}
		case <-timer1.C:
			deleteOldLogs()
		case <-timer2.C:
			if len(list) > 0 {
				saveLogList(list)
				list = []*EventLogEnt{}
			}
		}
	}
}

func saveLogList(list []*EventLogEnt) {
	if db == nil {
		return
	}
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
		ipfix := tx.Bucket([]byte("ipfix"))
		trap := tx.Bucket([]byte("trap"))
		arp := tx.Bucket([]byte("arplog"))
		sc := 0
		nfc := 0
		tc := 0
		ac := 0
		oc := 0
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
			default:
				oc++
			}
		}
		log.Printf("syslog=%d,netflow=%d,trap=%d,arplog=%d,other=%d,dur=%v", sc, nfc, tc, ac, oc, time.Since(st))
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
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return s
	}
	return d
}
