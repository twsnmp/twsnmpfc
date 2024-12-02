package datastore

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type LineEnt struct {
	ID         string
	NodeID1    string
	PollingID1 string
	State1     string
	NodeID2    string
	PollingID2 string
	State2     string
	PollingID  string
	Width      int
	State      string
	Info       string
	Port       string
}

func AddLine(l *LineEnt) error {
	st := time.Now()
	for {
		l.ID = makeKey()
		if _, ok := lines.Load(l.ID); !ok {
			break
		}
	}
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	lines.Store(l.ID, l)
	log.Printf("AddLine dur=%v", time.Since(st))
	return nil
}

func UpdateLine(l *LineEnt) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := lines.Load(l.ID); !ok {
		return ErrInvalidID
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	log.Printf("UpdateLine dur=%v", time.Since(st))
	return nil
}

func DeleteLine(lineID string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := lines.Load(lineID); !ok {
		return ErrInvalidID
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Delete([]byte(lineID))
	})
	lines.Delete(lineID)
	log.Printf("delete line dur=%v", time.Since(st))
	return nil
}

func GetLine(lineID string) *LineEnt {
	if db == nil {
		return nil
	}
	if n, ok := lines.Load(lineID); ok {
		return n.(*LineEnt)
	}
	return nil
}

// ForEachLines : Line毎の処理
func ForEachLines(f func(*LineEnt) bool) {
	lines.Range(func(_, v interface{}) bool {
		return f(v.(*LineEnt))
	})
}

// HasLine : 同じラインを探す
func HasLine(l1 *LineEnt, isNet bool) bool {
	ret := false
	lines.Range(func(_, v interface{}) bool {
		if l2, ok := v.(*LineEnt); ok {
			if isNet {
				if l1.NodeID1 == l2.NodeID1 && l1.PollingID1 == l2.PollingID1 &&
					l1.NodeID2 == l2.NodeID2 && l1.PollingID2 == l2.PollingID2 {
					ret = true
					return false
				}
				if l1.NodeID1 == l2.NodeID2 && l1.PollingID1 == l2.PollingID2 &&
					l1.NodeID2 == l2.NodeID1 && l1.PollingID2 == l2.PollingID1 {
					ret = true
					return false
				}
			} else {
				if l1.NodeID1 == l2.NodeID1 && l1.NodeID2 == l2.NodeID2 {
					ret = true
					return false
				}
				if l1.NodeID1 == l2.NodeID2 && l1.NodeID2 == l2.NodeID1 {
					ret = true
					return false
				}
			}
		}
		return true
	})
	return ret
}
