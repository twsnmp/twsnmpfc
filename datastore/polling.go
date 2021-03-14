package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type PollingEnt struct {
	ID        string
	Name      string
	NodeID    string
	Type      string
	Mode      string
	Params    string
	Filter    string
	Extractor string
	Script    string
	Level     string
	PollInt   int
	Timeout   int
	Retry     int
	LogMode   int
	NextTime  int64
	LastTime  int64
	Result    map[string]interface{}
	State     string
}

type PollingLogEnt struct {
	Time      int64 // UnixNano()
	PollingID string
	State     string
	Result    map[string]interface{}
}

// AddPolling : ポーリングを追加する
func AddPolling(p *PollingEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	for {
		p.ID = makeKey()
		if _, ok := pollings.Load(p.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
	p.Result = make(map[string]interface{})
	pollings.Store(p.ID, p)
	return nil
}

func UpdatePolling(p *PollingEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := pollings.Load(p.ID); !ok {
		return ErrInvalidID
	}
	p.LastTime = time.Now().UnixNano()
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
	pollings.Store(p.ID, p)
	return nil
}

func DeletePolling(pollingID string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := pollings.Load(pollingID); !ok {
		return ErrInvalidID
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Delete([]byte(pollingID))
	})
	pollings.Delete(pollingID)
	// Delete lines
	lines.Range(func(_, p interface{}) bool {
		l := p.(*LineEnt)
		if l.PollingID1 == pollingID || l.PollingID2 == pollingID {
			_ = DeleteLine(l.ID)
		}
		return true
	})
	ClearPollingLog(pollingID)
	DeleteAIResult(pollingID)
	return nil
}

// GetPolling : ポーリングを取得する
func GetPolling(id string) *PollingEnt {
	p, _ := pollings.Load(id)
	return p.(*PollingEnt)
}

// ForEachPollings : ポーリング毎の処理
func ForEachPollings(f func(*PollingEnt) bool) {
	pollings.Range(func(_, p interface{}) bool {
		return f(p.(*PollingEnt))
	})
}

func AddPollingLog(p *PollingEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	pl := PollingLogEnt{
		Time:      time.Now().UnixNano(),
		PollingID: p.ID,
		State:     p.State,
		Result:    p.Result,
	}
	s, err := json.Marshal(pl)
	if err != nil {
		return err
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		return b.Put([]byte(makeKey()), s)
	})
	return nil
}

func ForEachPollingLog(st, et int64, pollingID string, f func(*PollingLogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x", st)
	return db.View(func(tx *bbolt.Tx) error {
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

func ClearPollingLog(pollingID string) error {
	return db.Batch(func(tx *bbolt.Tx) error {
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

// GetAllPollingLog この関数も不要なので削除予定
func GetAllPollingLog(pollingID string) []PollingLogEnt {
	ret := []PollingLogEnt{}
	if db == nil {
		return ret
	}
	_ = db.View(func(tx *bbolt.Tx) error {
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
