// Package datastore : データ保存
package datastore

import (
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod

	"go.etcd.io/bbolt"
)

type DBStatsEnt struct {
	Time        int64
	Duration    float64
	Size        int64
	TotalWrite  int64
	Write       int64
	PeakWrite   int64
	Speed       float64
	AvgSpeed    float64
	PeakSpeed   float64
	BackupTime  int64
	BackupStart int64
}

type DBStatsLogEnt struct {
	Time  int64
	Size  int64
	Speed float64
}

func UpdateDBStats() {
	if db == nil {
		return
	}
	s := db.Stats()
	d := s.Sub(&prevDBStats)
	var dbSize int64
	_ = db.View(func(tx *bbolt.Tx) error {
		dbSize = tx.Size()
		return nil
	})
	DBStats.Size = dbSize
	DBStats.TotalWrite = int64(s.TxStats.Write)
	DBStats.Write = int64(d.TxStats.Write)
	if DBStats.PeakWrite < DBStats.Write {
		DBStats.PeakWrite = DBStats.Write
	}
	skipLog := true
	// 初回は計算しない。
	if DBStats.PeakWrite > 0 && DBStats.Time != 0 {
		DBStats.Duration = time.Since(dbOpenTime).Seconds()
		if DBStats.Duration > 0 {
			DBStats.AvgSpeed = float64(s.TxStats.Write) / DBStats.Duration
		}
		skipLog = false
	}
	dt := d.TxStats.WriteTime.Seconds()
	if dt != 0 {
		DBStats.Speed = float64(d.TxStats.Write) / dt
		if DBStats.PeakSpeed < DBStats.Speed {
			DBStats.PeakSpeed = DBStats.Speed
		}
	} else {
		DBStats.Speed = 0.0
	}
	DBStats.Time = time.Now().UnixNano()
	prevDBStats = s
	if skipLog {
		return
	}
	DBStatsLog = append(DBStatsLog, DBStatsLogEnt{
		Time:  DBStats.Time,
		Size:  DBStats.Size,
		Speed: DBStats.Speed,
	})
	if len(DBStatsLog) > 24*60*7 {
		DBStatsLog = DBStatsLog[1:]
	}
}

// MonitorDataEnt :
type MonitorDataEnt struct {
	CPU   float64
	Mem   float64
	Disk  float64
	Load  float64
	Bytes float64
	Net   float64
	Proc  int
	Conn  int
	At    int64
}

// MonitorDataes : モニターデータ
var MonitorDataes []*MonitorDataEnt
