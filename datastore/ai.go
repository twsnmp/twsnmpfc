package datastore

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
)

type AIResult struct {
	PollingID string
	LastTime  int64
	LossData  [][]float64
	ScoreData [][]float64
}

func SaveAIResultToDB(res *AIResult) error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		if b == nil {
			return fmt.Errorf("bucket ai is nil")
		}
		return b.Put([]byte(res.PollingID), s)
	})
}

func LoadAIReesult(id string) (*AIResult, error) {
	var ret AIResult
	r := ""
	if db == nil {
		return &ret, ErrDBNotOpen
	}
	_ = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		if b == nil {
			return nil
		}
		tmp := b.Get([]byte(id))
		if tmp != nil {
			r = string(tmp)
		}
		return nil
	})
	if r == "" {
		return &ret, fmt.Errorf("ai result not found id=%v", id)
	}
	if err := json.Unmarshal([]byte(r), &ret); err != nil {
		return &ret, err
	}
	return &ret, nil
}

func DeleteAIResult(id string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		_ = b.Delete([]byte(id))
		return nil
	})
}
