package datastore

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

type OTelMetricDataPointEnt struct {
	Start      int64    `json:"Start"`
	Time       int64    `json:"Time"`
	Attributes []string `json:"Attributes"`
	// Histogram
	Count          uint64    `json:"Count"`
	BucketCounts   []uint64  `json:"BucketCounts"`
	ExplicitBounds []float64 `json:"ExplicitBounds"`
	Sum            float64   `json:"Sum"`
	Min            float64   `json:"Min"`
	Max            float64   `json:"Max"`
	// Gauge
	Gauge float64 `json:"Gauge"`
	// ExponentialHistogram
	Positive      []uint64 `json:"Positive"`
	Negative      []uint64 `json:"Negative"`
	Scale         int64    `json:"Scale"`
	ZeroCount     int64    `json:"ZeroCount"`
	ZeroThreshold float64  `json:"ZeroThreshold"`
	// Index
	Index int `json:"Index"`
}

type OTelMetricEnt struct {
	Host        string                    `json:"Host"`
	Service     string                    `json:"Service"`
	Scope       string                    `json:"Scope"`
	Name        string                    `json:"Name"`
	Type        string                    `json:"Type"`
	Description string                    `json:"Description"`
	Unit        string                    `json:"Unit"`
	DataPoints  []*OTelMetricDataPointEnt `json:"DataPoints"`
	Count       int                       `json:"Count"`
	First       int64                     `json:"First"`
	Last        int64                     `json:"Last"`
}

var metricMap sync.Map

type OTelTraceSpanEnt struct {
	SpanID       string   `json:"SpanID"`
	ParentSpanID string   `json:"ParentSpanID"`
	Host         string   `json:"Host"`
	Service      string   `json:"Service"`
	Scope        string   `json:"Scope"`
	Name         string   `json:"Name"`
	Start        int64    `json:"Start"`
	End          int64    `json:"End"`
	Dur          float64  `json:"Dur"`
	Attributes   []string `json:"Attributes"`
}

type OTelTraceEnt struct {
	Bucket  string             `json:"Bucket"`
	TraceID string             `json:"TraceID"`
	Start   int64              `json:"Start"`
	End     int64              `json:"End"`
	Dur     float64            `json:"Dur"`
	Spans   []OTelTraceSpanEnt `json:"Spans"`
	Last    int64              `json:"Last"`
}

func AddOTelMetric(m *OTelMetricEnt) {
	k := getOTelMetricKey(m.Host, m.Service, m.Scope, m.Name)
	metricMap.Store(k, m)
}

func ForEachOTelMetric(f func(id string, m *OTelMetricEnt) bool) {
	metricMap.Range(func(key any, value any) bool {
		if m, ok := value.(*OTelMetricEnt); ok {
			if k, ok := key.(string); ok {
				return f(k, m)
			}
		}
		return true
	})
}

func FindOTelMetric(host, service, scope, name string) *OTelMetricEnt {
	id := getOTelMetricKey(host, service, scope, name)
	return FindOTelMetricByID(id)
}

func FindOTelMetricByID(id string) *OTelMetricEnt {
	if v, ok := metricMap.Load(id); ok {
		if m, ok := v.(*OTelMetricEnt); ok {
			return m
		}
	}
	return nil
}

func DeleteOTelMetric(m *OTelMetricEnt) {
	k := getOTelMetricKey(m.Host, m.Service, m.Scope, m.Name)
	metricMap.Delete(k)
}

func LoadOTelMetric() {
	if db == nil {
		return
	}
	st := time.Now()
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("otelMetric"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k []byte, v []byte) error {
			var m OTelMetricEnt
			if err := json.Unmarshal(v, &m); err == nil {
				metricMap.Store(string(k), &m)
			}
			return nil
		})
	})
	log.Printf("load otel metric dur=%v", time.Since(st))
}

// SaveOTelMetricはメトリックをDBに保存します。
func SaveOTelMetric() {
	if db == nil {
		return
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		tx.DeleteBucket([]byte("otelMetric"))
		b, err := tx.CreateBucket([]byte("otelMetric"))
		if b == nil || err != nil {
			return nil
		}
		metricMap.Range(func(key any, value any) bool {
			if k, ok := key.(string); ok {
				if m, ok := value.(*OTelMetricEnt); ok {
					if j, err := json.Marshal(m); err == nil {
						b.Put([]byte(k), j)
					}
				}
			}
			return true
		})
		return nil
	})
	log.Printf("save otel metric dur=%v", time.Since(st))
}

func getOTelMetricKey(host, service, scope, name string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%s\t%s\t%s\t%s", host, service, scope, name))))
}

func UpdateOTelTrace(list []*OTelTraceEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	st := time.Now()
	err := db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("otelTrace"))
		if b == nil {
			return nil
		}
		for _, t := range list {
			j, err := json.Marshal(t)
			if err != nil {
				continue
			}

			bt, err := b.CreateBucketIfNotExists([]byte(t.Bucket))
			if err != nil {
				continue
			}
			bt.Put([]byte(t.TraceID), j)
		}
		return nil
	})
	log.Printf("update otel trace len=%d dur=%v", len(list), time.Since(st))
	return err
}

func GetOTelTrace(bucket, tid string) *OTelTraceEnt {
	if db == nil {
		return nil
	}
	var ret *OTelTraceEnt
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("otelTrace"))
		if b == nil {
			return nil
		}
		bt := b.Bucket([]byte(bucket))
		if bt == nil {
			return nil
		}
		if v := bt.Get([]byte(tid)); v != nil {
			var t OTelTraceEnt
			json.Unmarshal(v, &t)
			ret = &t
		}
		return nil
	})
	return ret
}

func ForEachOTelTrace(tbk string, f func(t *OTelTraceEnt) bool) {
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("otelTrace"))
		if b == nil {
			return nil
		}
		bt := b.Bucket([]byte(tbk))
		if bt == nil {
			return fmt.Errorf("bucket not fond")
		}
		bt.ForEach(func(k []byte, v []byte) error {
			var t OTelTraceEnt
			if err := json.Unmarshal(v, &t); err == nil {
				if !f(&t) {
					return fmt.Errorf("stop earch")
				}
			}
			return nil
		})
		return nil
	})
}

func GetOTelTraceBucketList() []string {
	ret := []string{}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("otelTrace"))
		if b == nil {
			return nil
		}
		b.ForEachBucket(func(k []byte) error {
			ret = append(ret, string(k))
			return nil
		})
		return nil
	})
	sort.Strings(ret)
	return ret
}

func chekOldOTelData() {
	delMetrics := []string{}
	t := time.Now().Add(time.Hour * time.Duration(MapConf.OTelRetention) * -1)
	tn := t.UnixNano()
	tbk := t.Format("2006-01-02T15:04")
	ForEachOTelMetric(func(id string, m *OTelMetricEnt) bool {
		if m.Last < tn {
			delMetrics = append(delMetrics, id)
		}
		return true
	})
	if len(delMetrics) > 0 {
		for _, id := range delMetrics {
			metricMap.Delete(id)
		}
		log.Printf("delete old otel metrics len=%d", len(delMetrics))
	}
	delTraceBucket := [][]byte{}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("otelTrace"))
		if b == nil {
			return nil
		}
		b.ForEachBucket(func(k []byte) error {
			if string(k) < tbk {
				delTraceBucket = append(delTraceBucket, k)
			}
			return nil
		})
		return nil
	})
	if len(delTraceBucket) > 0 {
		st := time.Now()
		db.Batch(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("otelTrace"))
			if b == nil {
				return nil
			}
			for _, k := range delTraceBucket {
				b.DeleteBucket(k)
			}
			return nil
		})
		log.Printf("delete old otel trace len=%d dur=%v", len(delTraceBucket), time.Since(st))
	}
}

func DeleteAllOTelData() error {
	metricMap.Clear()
	if db == nil {
		return nil
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		tx.DeleteBucket([]byte("otelTrace"))
		tx.CreateBucketIfNotExists([]byte("otelTrace"))
		return nil
	})
}
