package datastore

import (
	"fmt"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type ArpEnt struct {
	IP  string
	MAC string
}

func UpdateArpEnt(ip, mac string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		log.Printf("UpdateArpEnt dur=%v", time.Since(st))
		return b.Put([]byte(ip), []byte(mac))
	})
}

func ForEachArp(f func(*ArpEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var e = ArpEnt{
				IP:  string(k),
				MAC: string(v),
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

func ResetArpTable() error {
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return fmt.Errorf("bucket arp not found")
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			_ = c.Delete()
		}
		log.Printf("ResetArpTable  dur=%v", time.Since(st))
		return nil
	})
}
