// Package datastore : データ保存
package datastore

import (
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod

	"go.etcd.io/bbolt"
)

type DBStatsEnt struct {
	Time       int64
	Duration   float64
	Size       int64
	TotalWrite int64
	Write      int64
	PeakWrite  int64
	Speed      float64
	AvgSpeed   float64
	PeakSpeed  float64
	BackupTime int64
}

type DBStatsLogEnt struct {
	Time  int64
	Size  int64
	Speed float64
}

func (ds *DataStore) UpdateDBStats() {
	if ds.db == nil {
		return
	}
	s := ds.db.Stats()
	d := s.Sub(&ds.prevDBStats)
	var dbSize int64
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		dbSize = tx.Size()
		return nil
	})
	ds.DBStats.Size = dbSize
	ds.DBStats.TotalWrite = int64(s.TxStats.Write)
	ds.DBStats.Write = int64(d.TxStats.Write)
	if ds.DBStats.PeakWrite < ds.DBStats.Write {
		ds.DBStats.PeakWrite = ds.DBStats.Write
	}
	skipLog := true
	// 初回は計算しない。
	if ds.DBStats.PeakWrite > 0 && ds.DBStats.Time != 0 {
		ds.DBStats.Duration = time.Since(ds.dbOpenTime).Seconds()
		if ds.DBStats.Duration > 0 {
			ds.DBStats.AvgSpeed = float64(s.TxStats.Write) / ds.DBStats.Duration
		}
		skipLog = false
	}
	dt := d.TxStats.WriteTime.Seconds()
	if dt != 0 {
		ds.DBStats.Speed = float64(d.TxStats.Write) / dt
		if ds.DBStats.PeakSpeed < ds.DBStats.Speed {
			ds.DBStats.PeakSpeed = ds.DBStats.Speed
		}
	} else {
		ds.DBStats.Speed = 0.0
	}
	ds.DBStats.Time = time.Now().UnixNano()
	ds.prevDBStats = s
	if skipLog {
		return
	}
	ds.DBStatsLog = append(ds.DBStatsLog, DBStatsLogEnt{
		Time:  ds.DBStats.Time,
		Size:  ds.DBStats.Size,
		Speed: ds.DBStats.Speed,
	})
	if len(ds.DBStatsLog) > 24*60*7 {
		ds.DBStatsLog = ds.DBStatsLog[1:]
	}
}
