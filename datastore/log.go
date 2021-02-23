package datastore

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"regexp"
	"strings"
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

type LogFilterEnt struct {
	StartTime string
	EndTime   string
	Filter    string
	LogType   string
}

func (ds *DataStore) AddEventLog(e *EventLogEnt) {
	e.Time = time.Now().UnixNano()
	log.Printf("log=%v", e)
	ds.eventLogCh <- e
}

func (ds *DataStore) ForEachEventLog(st, et int64, f func(*EventLogEnt) bool) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x", st)
	return ds.db.View(func(tx *bbolt.Tx) error {
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

func (ds *DataStore) ForEachLastEventLog(skey string, f func(*EventLogEnt) bool) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	return ds.db.View(func(tx *bbolt.Tx) error {
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

func (ds *DataStore) ForEachLog(st, et int64, t string, f func(*LogEnt) bool) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x", st)
	return ds.db.View(func(tx *bbolt.Tx) error {
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

// --

type logFilterParamEnt struct {
	StartKey    string
	StartTime   int64
	EndTime     int64
	RegexFilter *regexp.Regexp
}

func ParseFilter(f string) string {
	f = strings.TrimSpace(f)
	if f == "``" {
		return ""
	}
	if strings.HasPrefix(f, "`") && strings.HasSuffix(f, "`") {
		return f[1 : len(f)-1]
	}
	f = regexp.QuoteMeta(f)
	f = strings.ReplaceAll(f, "\\*", ".+")
	return f
}

func getFilterParams(filter *LogFilterEnt) *logFilterParamEnt {
	var err error
	var t time.Time
	ret := &logFilterParamEnt{}
	t, err = time.Parse("2006-01-02T15:04 MST", filter.StartTime+" JST")
	if err == nil {
		ret.StartTime = t.UnixNano()
	} else {
		log.Printf("getFilterParams err=%v", err)
		ret.StartTime = time.Now().Add(-time.Hour * 24).UnixNano()
	}
	t, err = time.Parse("2006-01-02T15:04 MST", filter.EndTime+" JST")
	if err == nil {
		ret.EndTime = t.UnixNano()
	} else {
		log.Printf("getFilterParams err=%v", err)
		ret.EndTime = time.Now().UnixNano()
	}
	ret.StartKey = fmt.Sprintf("%016x", ret.StartTime)
	filter.Filter = strings.TrimSpace(filter.Filter)
	if filter.Filter == "" {
		return ret
	}
	fs := ParseFilter(filter.Filter)
	ret.RegexFilter, err = regexp.Compile(fs)
	if err != nil {
		log.Printf("getFilterParams err=%v", err)
		ret.RegexFilter = nil
	}
	return ret
}

func (ds *DataStore) GetEventLogs(filter *LogFilterEnt) []EventLogEnt {
	ret := []EventLogEnt{}
	if ds.db == nil {
		return ret
	}
	f := getFilterParams(filter)
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek([]byte(f.StartKey)); k != nil && i < MaxDispLog; k, v = c.Next() {
			var e EventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Printf("getEventLogs err=%v", err)
				continue
			}
			if e.Time < f.StartTime {
				continue
			}
			if e.Time > f.EndTime {
				break
			}
			if f.RegexFilter != nil && !f.RegexFilter.Match(v) {
				continue
			}
			ret = append(ret, e)
			i++
		}
		return nil
	})
	return ret
}

func (ds *DataStore) GetLogs(filter *LogFilterEnt) []LogEnt {
	ret := []LogEnt{}
	if ds.db == nil {
		return ret
	}
	f := getFilterParams(filter)
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(filter.LogType))
		if b == nil {
			log.Printf("getLogs no Bucket=%s", filter.LogType)
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek([]byte(f.StartKey)); k != nil && i < MaxDispLog; k, v = c.Next() {
			if bytes.HasSuffix(v, []byte{0, 0, 255, 255}) {
				v = deCompressLog(v)
			}
			var l LogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				log.Printf("getLogs err=%v", err)
				continue
			}
			if l.Time < f.StartTime {
				continue
			}
			if l.Time > f.EndTime {
				break
			}
			if f.RegexFilter != nil && !f.RegexFilter.Match(v) {
				continue
			}
			ret = append(ret, l)
			i++
		}
		return nil
	})
	return ret
}

func (ds *DataStore) AddPollingLog(p *PollingEnt) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	pl := PollingLogEnt{
		Time:      time.Now().UnixNano(),
		PollingID: p.ID,
		State:     p.State,
		NumVal:    p.LastVal,
		StrVal:    p.LastResult,
	}
	s, err := json.Marshal(pl)
	if err != nil {
		return err
	}
	_ = ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		return b.Put([]byte(makeKey()), s)
	})
	return nil
}

func (ds *DataStore) ForEachPollingLog(st, et int64, pollingID string, f func(*PollingLogEnt) bool) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x", st)
	return ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Seek([]byte(sk)); k != nil; k, v = c.Next() {
			var e PollingLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Printf("ForEachPollingLog v=%s err=%v", v, err)
				continue
			}
			if e.PollingID != pollingID {
				continue
			}
			if math.IsNaN(e.NumVal) {
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

func (ds *DataStore) GetPollingLog(startTime, endTime, pollingID string) []PollingLogEnt {
	ret := []PollingLogEnt{}
	if ds.db == nil {
		return ret
	}
	var st int64
	var et int64
	if t, err := time.Parse("2006-01-02T15:04 MST", startTime+" JST"); err == nil {
		st = t.UnixNano()
	} else {
		log.Printf("getPollingLog err=%v", err)
		st = time.Now().Add(-time.Hour * 24).UnixNano()
	}
	if t, err := time.Parse("2006-01-02T15:04 MST", endTime+" JST"); err == nil {
		et = t.UnixNano()
	} else {
		log.Printf("getFilterParams err=%v", err)
		et = time.Now().UnixNano()
	}
	startKey := fmt.Sprintf("%016x", st)
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			log.Printf("getPollingLog no Bucket getPollingLog")
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek([]byte(startKey)); k != nil && i < MaxDispLog; k, v = c.Next() {
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			var l PollingLogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				log.Printf("getPollingLog err=%v", err)
				continue
			}
			if l.Time < st {
				continue
			}
			if l.Time > et {
				break
			}
			if l.PollingID != pollingID {
				continue
			}
			if math.IsNaN(l.NumVal) {
				continue
			}
			ret = append(ret, l)
			i++
		}
		return nil
	})
	return ret
}

func (ds *DataStore) GetAllPollingLog(pollingID string) []PollingLogEnt {
	ret := []PollingLogEnt{}
	if ds.db == nil {
		return ret
	}
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			log.Printf("getPollingLog no Bucket getPollingLog")
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.First(); k != nil && i < MaxDispLog*100; k, v = c.Next() {
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			var l PollingLogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				log.Printf("getPollingLog err=%v", err)
				continue
			}
			if l.PollingID != pollingID {
				continue
			}
			ret = append(ret, l)
			i++
		}
		return nil
	})
	return ret
}

func (ds *DataStore) ClearPollingLog(pollingID string) error {
	return ds.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			return fmt.Errorf("bucket pollingLogs not found")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			_ = c.Delete()
		}
		b = tx.Bucket([]byte("ai"))
		if b != nil {
			_ = b.Delete([]byte(pollingID))
		}
		return nil
	})
}

func (ds *DataStore) deleteOldLog(bucket string, days int) error {
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	return ds.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if st < string(k) || ds.delCount > MaxDelLog {
				break
			}
			_ = c.Delete()
			ds.delCount++
		}
		return nil
	})
}

func (ds *DataStore) deleteOldLogs() {
	ds.delCount = 0
	if ds.MapConf.LogDays < 1 {
		log.Println("mapConf.LogDays < 1 ")
		return
	}
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix"}
	for _, b := range buckets {
		if err := ds.deleteOldLog(b, ds.MapConf.LogDays); err != nil {
			log.Printf("deleteOldLog err=%v", err)
		}
	}
	if ds.delCount > 0 {
		log.Printf("DeleteLogs=%d", ds.delCount)
	}
}

func (ds *DataStore) DeleteAllLogs() {
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix"}
	for _, b := range buckets {
		ds.db.Batch(func(tx *bbolt.Tx) error {
			if err := tx.DeleteBucket([]byte(b)); err != nil {
				return err
			}
			tx.CreateBucketIfNotExists([]byte(b))
			return nil
		})
	}
}

func (ds *DataStore) eventLogger(ctx context.Context) {
	log.Println("Start EventLogger")
	timer1 := time.NewTicker(time.Minute * 2)
	timer2 := time.NewTicker(time.Second * 5)
	list := []*EventLogEnt{}
	for {
		select {
		case <-ctx.Done():
			if len(list) > 0 {
				ds.saveLogList(list)
			}
			timer1.Stop()
			timer2.Stop()
			return
		case e := <-ds.eventLogCh:
			list = append(list, e)
			if len(list) > 100 {
				ds.saveLogList(list)
				list = []*EventLogEnt{}
			}
		case <-timer1.C:
			ds.deleteOldLogs()
		case <-timer2.C:
			if len(list) > 0 {
				ds.saveLogList(list)
				list = []*EventLogEnt{}
			}
		}
	}
}

func (ds *DataStore) saveLogList(list []*EventLogEnt) {
	if ds.db == nil {
		return
	}
	log.Printf("saveLogList len=%d", len(list))
	ds.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		for _, e := range list {
			s, err := json.Marshal(e)
			if err != nil {
				return err
			}
			err = b.Put([]byte(fmt.Sprintf("%016x", e.Time)), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ds *DataStore) SaveLogBuffer(logBuffer []*LogEnt) {
	if ds.db == nil {
		log.Printf("saveLogBuffer DB Not open")
		return
	}
	_ = ds.db.Batch(func(tx *bbolt.Tx) error {
		syslog := tx.Bucket([]byte("syslog"))
		netflow := tx.Bucket([]byte("netflow"))
		ipfix := tx.Bucket([]byte("ipfix"))
		trap := tx.Bucket([]byte("trap"))
		for _, l := range logBuffer {
			k := fmt.Sprintf("%016x", l.Time)
			s, err := json.Marshal(l)
			if err != nil {
				return err
			}
			ds.logSize += int64(len(s))
			if len(s) > 100 {
				s = compressLog(s)
			}
			ds.compLogSize += int64(len(s))
			switch l.Type {
			case "syslog":
				_ = syslog.Put([]byte(k), []byte(s))
			case "netflow":
				_ = netflow.Put([]byte(k), []byte(s))
			case "ipfix":
				_ = ipfix.Put([]byte(k), []byte(s))
			case "trap":
				_ = trap.Put([]byte(k), []byte(s))
			}
		}
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
