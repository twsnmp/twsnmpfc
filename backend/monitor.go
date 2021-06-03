package backend

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	gopsnet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

const (
	maxMonitorData = 4 * 60 * 24
)

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

func updateMonData() {
	m := &MonitorDataEnt{}
	cpus, err := cpu.Percent(0, false)
	if err == nil {
		m.CPU = cpus[0]
	}
	l, err := load.Avg()
	if err == nil {
		m.Load = l.Load1
	}
	v, err := mem.VirtualMemory()
	if err == nil {
		m.Mem = v.UsedPercent
	}
	m.At = time.Now().Unix()
	d, err := disk.Usage(dspath)
	if err == nil {
		m.Disk = d.UsedPercent
	}
	n, err := gopsnet.IOCounters(false)
	if err == nil {
		m.Bytes = float64(n[0].BytesRecv)
		m.Bytes += float64(n[0].BytesSent)
		if len(MonitorDataes) > 1 {
			o := MonitorDataes[len(MonitorDataes)-1]
			m.Net = float64(8.0 * (m.Bytes - o.Bytes) / float64(m.At-o.At))
		}
	}
	conn, err := gopsnet.Connections("tcp")
	if err == nil {
		m.Conn = len(conn)
	}
	pids, err := process.Pids()
	if err == nil {
		m.Proc = len(pids)
	}
	for len(MonitorDataes) > maxMonitorData {
		MonitorDataes = append(MonitorDataes[:0], MonitorDataes[1:]...)
	}
	MonitorDataes = append(MonitorDataes, m)
}

// monitor :
func monitor(ctx context.Context) {
	log.Println("start monitor")
	// 60秒に同期する
	for (time.Now().Second() % 60) != 0 {
		time.Sleep(time.Millisecond * 100)
	}
	timer := time.NewTicker(time.Second * 60)
	updateMonData()
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			updateMonData()
		}
	}
}
