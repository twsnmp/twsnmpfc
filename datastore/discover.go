package datastore

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
)

type DiscoverConfEnt struct {
	StartIP string
	EndIP   string
	Timeout int
	Retry   int
	X       int
	Y       int
}

func SaveDiscoverConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(DiscoverConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("discoverConf"), s)
	})
}
