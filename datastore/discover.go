package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type DiscoverConfEnt struct {
	Active          bool
	StartIP         string `validate:"required,ipv4"`
	EndIP           string `validate:"required,ipv4"`
	AutoAddPollings []string
	Timeout         int `validate:"required,gte=1,lte=10"`
	Retry           int `validate:"required,gte=0,lte=5"`
	X               int
	Y               int
	AddNetwork      bool
}

func SaveDiscoverConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(DiscoverConf)
	if err != nil {
		return err
	}
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveDiscoverConf dur=%v", time.Since(st))
		return b.Put([]byte("discoverConf"), s)
	})
}
