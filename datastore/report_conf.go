package datastore

import (
	"encoding/json"
	"fmt"
	"log"

	"go.etcd.io/bbolt"
)

type ReportConfEnt struct {
	DenyCountries        []string
	DenyServices         []string
	AllowDNS             string
	AllowDHCP            string
	AllowMail            string
	AllowLDAP            string
	AllowLocalIP         string
	JapanOnly            bool
	DropFlowThTCPPacket  int
	RetentionTimeForSafe int
	IncludeNoMACIP       bool
}

var ReportConf ReportConfEnt

// LaodReportConf : レポート設定を読み込む
func LaodReportConf() error {
	ReportConf.RetentionTimeForSafe = 24
	ReportConf.DropFlowThTCPPacket = 3
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("report"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &ReportConf); err != nil {
			log.Printf("Unmarshal mapConf from DB error=%v", err)
			return err
		}
		return nil
	})
}

// SaveReportConf : レポート設定を保存する
func SaveReportConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ReportConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("report"), s)
	})
}
