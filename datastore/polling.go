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
	st := time.Now()
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
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
	p.Result = make(map[string]interface{})
	pollings.Store(p.ID, p)
	SetNodeStateChanged(p.NodeID)
	log.Printf("AddPolling dur=%v", time.Since(st))
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
	pollings.Store(p.ID, p)
	return nil
}

func DeletePollings(ids []string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	for _, id := range ids {
		if e, ok := pollings.Load(id); ok {
			p := e.(*PollingEnt)
			SetNodeStateChanged(p.NodeID)
			pollings.Delete(id)
		}
	}
	// Delete lines
	lines.Range(func(_, p interface{}) bool {
		l := p.(*LineEnt)
		for _, id := range ids {
			if l.PollingID1 == id || l.PollingID2 == id {
				_ = DeleteLine(l.ID)
				return true
			}
		}
		return true
	})
	db.Batch(func(tx *bbolt.Tx) error {
		pb := tx.Bucket([]byte("pollings"))
		aib := tx.Bucket([]byte("ai"))
		if pb != nil && aib != nil {
			for _, id := range ids {
				pb.Delete([]byte(id))
				aib.Delete([]byte(id))
			}
		}
		return nil
	})
	go clearDeletedPollingLogs(ids)
	log.Printf("DeletePollings dur=%v", time.Since(st))
	return nil
}

// GetPolling : ポーリングを取得する
func GetPolling(id string) *PollingEnt {
	if p, ok := pollings.Load(id); ok {
		return p.(*PollingEnt)
	}
	return nil
}

// ForEachPollings : ポーリング毎の処理
func ForEachPollings(f func(*PollingEnt) bool) {
	pollings.Range(func(_, p interface{}) bool {
		return f(p.(*PollingEnt))
	})
}

func saveAllPollings() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		pollings.Range(func(_, p interface{}) bool {
			pe := p.(*PollingEnt)
			s, err := json.Marshal(pe)
			if err == nil {
				b.Put([]byte(pe.ID), s)
			}
			return true
		})
		return nil
	})
	log.Printf("saveAllPollings dur=%v", time.Since(st))
	return nil
}

func AddPollingLog(p *PollingEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	pollingLogCh <- &PollingLogEnt{
		Time:      time.Now().UnixNano(),
		PollingID: p.ID,
		State:     p.State,
		Result:    p.Result,
	}
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
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			var e PollingLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Printf("load polling log err=%v", err)
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

// ClearPollingLog : ポーリングログを削除する
func ClearPollingLog(pollingID string) error {
	st := time.Now()
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
			var e PollingLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Printf("ClearPollingLog log err=%v", err)
				continue
			}
			if e.PollingID != pollingID {
				continue
			}
			c.Delete()
		}
		log.Printf("ClearPollingLog id=%s,dur=%v", pollingID, time.Since(st))
		return nil
	})
}

// clearDeletedPollingLogs : ポーリングログの削除をまとめて行う
func clearDeletedPollingLogs(ids []string) error {
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			return fmt.Errorf("bucket pollingLogs not found")
		}
		c := b.Cursor()
		del := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			for _, id := range ids {
				if bytes.Contains(v, []byte(id)) {
					var e PollingLogEnt
					err := json.Unmarshal(v, &e)
					if err != nil {
						log.Printf("ClearDeletedPollingLogs err=%v", err)
					} else {
						if e.PollingID == id {
							_ = c.Delete()
							del++
							break
						}
					}
				}
			}
		}
		log.Printf("clearDeletedPollingLogs del=%d,dur=%v", del, time.Since(st))
		return nil
	})
}

// GetAllPollingLog :全てのポーリングログを取得する
func GetAllPollingLog(pollingID string) []PollingLogEnt {
	ret := []PollingLogEnt{}
	if db == nil {
		return ret
	}
	_ = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			log.Printf("no polling log bucket")
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			var l PollingLogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				log.Printf("get polling log err=%v", err)
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
