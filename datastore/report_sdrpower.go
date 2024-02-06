package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

type SdrPowerEnt struct {
	Host string
	Time int64
	Freq int64
	Dbm  float64
}

func AddSdrPower(list []*SdrPowerEnt) {
	st := time.Now()
	if db == nil || len(list) < 1 {
		return
	}
	err := db.Batch(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		b := r.Bucket([]byte("sdrPower"))
		if b == nil {
			return fmt.Errorf("no bucket sdrPower")
		}
		for _, e := range list {
			id := fmt.Sprintf("%016x:%s:%016x", e.Time, e.Host, e.Freq)
			s, err := json.Marshal(e)
			if err != nil {
				return err
			}
			err = b.Put([]byte(id), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("AddSdrPower err=%v", err)
	}
	log.Printf("AddSdrPower dur=%v", time.Since(st))
}

func ForEachSdrPower(st int64, h string, f func(*SdrPowerEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x:%s:", st, h)
	return db.View(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		if r == nil {
			return nil
		}
		b := r.Bucket([]byte("sdrPower"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Seek([]byte(sk)); bytes.HasPrefix(k, []byte(sk)); k, v = c.Next() {
			var e SdrPowerEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Printf("ForEachSdrPower v=%s err=%v", v, err)
				continue
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

type SdrPowerKey struct {
	Host string
	Time int64
}

func GetSdrPowerKeys() []SdrPowerKey {
	m := make(map[SdrPowerKey]bool)
	db.View(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		if r == nil {
			return nil
		}
		b := r.Bucket([]byte("sdrPower"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var e SdrPowerEnt
				if err := json.Unmarshal(v, &e); err == nil {
					m[SdrPowerKey{
						Host: e.Host,
						Time: e.Time,
					}] = true
				}
				return nil
			})
		}
		return nil
	})
	keys := []SdrPowerKey{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func DeleteSdrPower(st int64, h string) error {
	s := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x:%s", st, h)
	return db.Batch(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		b := r.Bucket([]byte("sdrPower"))
		if b == nil {
			return fmt.Errorf("no bucket sdrPower")
		}
		c := b.Cursor()
		for k, _ := c.Seek([]byte(sk)); bytes.HasPrefix(k, []byte(sk)); k, _ = c.Next() {
			c.Delete()
		}
		log.Printf("DeleteSdrPower dur=%v", time.Since(s))
		return nil
	})
}

// DeleteOldSdrPower は古い電波強度レポートを削除します。
func DeleteOldSdrPower(delOld int64) error {
	s := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	dk := fmt.Sprintf("%016x", delOld)
	return db.Batch(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		b := r.Bucket([]byte("sdrPower"))
		if b == nil {
			return fmt.Errorf("no bucket sdrPower")
		}
		c := b.Cursor()
		count := 0
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			a := strings.Split(string(k), ":")
			if len(a) > 2 && a[0] > dk {
				log.Printf("DeleteOldSdrPower end %v %s", a, dk)
				break
			}
			c.Delete()
			count++
		}
		log.Printf("DeleteOldSdrPower count=%d dur=%v", count, time.Since(s))
		return nil
	})
}
