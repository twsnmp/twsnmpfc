package datastore

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type DrawItemType int

const (
	DrawItemTypeRect = iota
	DrawItemTypeEllipse
	DrawItemTypeText
	DrawItemTypeImage
	DrawItemTypePollingText
	DrawItemTypePollingGauge
	DrawItemTypePollingNewGauge
	DrawItemTypePollingBar
	DrawItemTypePollingLine
)

type DrawItemEnt struct {
	ID        string
	Type      DrawItemType
	X         int
	Y         int
	W         int // Width
	H         int // Higeht
	Color     string
	Path      string
	Text      string
	Size      int       // Font Size | GaugeSize
	PollingID string    // Polling ID
	VarName   string    // Pollingから取得する項目
	Format    string    // 表示フォーマット
	Value     float64   // Gauge,Barの値
	Values    []float64 // Lineの値
	Scale     float64   // 値の補正倍率
}

func AddDrawItem(di *DrawItemEnt) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	for {
		di.ID = makeKey()
		if _, ok := items.Load(di.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(di)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		return b.Put([]byte(di.ID), s)
	})
	items.Store(di.ID, di)
	log.Printf("AddItem  dur=%v", time.Since(st))
	return nil
}

func DeleteDrawItem(id string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := items.Load(id); !ok {
		return ErrInvalidID
	} else {
		AddEventLog(&EventLogEnt{
			Type:  "user",
			Level: "info",
			Event: "描画アイテムを削除しました",
		})
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		return b.Delete([]byte(id))
	})
	items.Delete(id)
	log.Printf("DeleteDrawItem dur=%v", time.Since(st))
	return nil
}

func GetDrawItem(id string) *DrawItemEnt {
	if db == nil {
		return nil
	}
	if di, ok := items.Load(id); ok {
		return di.(*DrawItemEnt)
	}
	return nil
}

func ForEachItems(f func(*DrawItemEnt) bool) {
	items.Range(func(_, p interface{}) bool {
		if d, ok := p.(*DrawItemEnt); ok {
			return f(d)
		}
		return true
	})
}
