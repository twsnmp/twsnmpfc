package datastore

import (
	"encoding/json"

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
	Info       string
}

func AddLine(l *LineEnt) error {
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
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	lines.Store(l.ID, l)
	return nil
}

func UpdateLine(l *LineEnt) error {
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
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	return nil
}

func DeleteLine(lineID string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := lines.Load(lineID); !ok {
		return ErrInvalidID
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Delete([]byte(lineID))
	})
	lines.Delete(lineID)
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
