// Package datastore : データ保存
package datastore

import (
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod

	"go.etcd.io/bbolt"
)

type DBStatsEnt struct {
	Time       int64
	Size       int64
	TotalWrite int64
	LastWrite  int64
	PeakWrite  int64
	AvgWrite   float64
	StartTime  int64
	Speed      float64
	Peak       float64
	Rate       float64
	BackupTime int64
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
	ds.DBStats.LastWrite = int64(d.TxStats.Write)
	if ds.DBStats.PeakWrite < ds.DBStats.LastWrite {
		ds.DBStats.PeakWrite = ds.DBStats.LastWrite
	}
	skipLog := true
	// 初回は計算しない。
	if ds.DBStats.PeakWrite > 0 && ds.DBStats.Time != 0 {
		ds.DBStats.Rate = 100 * float64(d.TxStats.Write) / float64(ds.DBStats.PeakWrite)
		ds.DBStats.StartTime = ds.dbOpenTime.UnixNano()
		dbot := time.Since(ds.dbOpenTime).Seconds()
		if dbot > 0 {
			ds.DBStats.AvgWrite = float64(s.TxStats.Write) / dbot
		}
		skipLog = false
	}
	dt := d.TxStats.WriteTime.Seconds()
	if dt != 0 {
		ds.DBStats.Speed = float64(d.TxStats.Write) / dt
		if ds.DBStats.Peak < ds.DBStats.Speed {
			ds.DBStats.Peak = ds.DBStats.Speed
		}
	} else {
		ds.DBStats.Speed = 0.0
	}
	ds.DBStats.Time = time.Now().UnixNano()
	ds.prevDBStats = s
	if skipLog {
		return
	}
	ds.DBStatsLog = append(ds.DBStatsLog, ds.DBStats)
	if len(ds.DBStatsLog) > 24*60*7 {
		ds.DBStatsLog = ds.DBStatsLog[1:]
	}
}
