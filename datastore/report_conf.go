package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type ReportConfEnt struct {
	DenyCountries       []string
	DenyServices        []string
	AllowDNS            string
	AllowDHCP           string
	AllowMail           string
	AllowLDAP           string
	AllowLocalIP        string
	JapanOnly           bool
	DropFlowThTCPPacket int
	SensorTimeout       int
	IncludeNoMACIP      bool
	ExcludeIPv6         bool
	ReportDays          int
	AICleanup           bool
}

var ReportConf ReportConfEnt

// LaodReportConf : レポート設定を読み込む
func LaodReportConf() error {
	ReportConf.DropFlowThTCPPacket = 3
	ReportConf.SensorTimeout = 1
	ReportConf.AICleanup = false
	ReportConf.ReportDays = 30
	ReportConf.ExcludeIPv6 = false
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("report"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &ReportConf); err != nil {
			log.Printf("load report conf err=%v", err)
			return err
		}
		return nil
	})
}

// SaveReportConf : レポート設定を保存する
func SaveReportConf() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(ReportConf)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveReportConf dur=%v", time.Since(st))
		return b.Put([]byte("report"), s)
	})
}
