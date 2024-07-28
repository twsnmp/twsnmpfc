package backend

import (
	"context"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	gopsnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/twsnmp/twsnmpfc/datastore"
)

const (
	maxMonitorData = 4 * 60 * 24
)

func updateMonData() {
	m := &datastore.MonitorDataEnt{}
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
	n, err := gopsnet.IOCounters(true)
	if err == nil {
		for _, nif := range n {
			if isMonitorIF(nif.Name) {
				m.Bytes += float64(nif.BytesRecv)
				m.Bytes += float64(nif.BytesSent)
			}
		}
		if len(datastore.MonitorDataes) >= 1 {
			o := datastore.MonitorDataes[len(datastore.MonitorDataes)-1]
			if m.Bytes >= o.Bytes && m.At > o.At {
				m.Net = float64(8.0 * (m.Bytes - o.Bytes) / float64(m.At-o.At))
			} else {
				log.Println("skip network monitor")
			}
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
	for len(datastore.MonitorDataes) > maxMonitorData {
		datastore.MonitorDataes = append(datastore.MonitorDataes[:0], datastore.MonitorDataes[1:]...)
	}
	datastore.MonitorDataes = append(datastore.MonitorDataes, m)
}

// isMonitorIF はモニターするLANポートを判断する
func isMonitorIF(n string) bool {
	switch runtime.GOOS {
	case "darwin":
		if strings.HasPrefix(n, "utun") {
			return false
		}
		if strings.HasPrefix(n, "lo") {
			return false
		}
	case "linux":
		if strings.HasPrefix(n, "lo") {
			return false
		}
	}
	return true
}

// monitor :
func monitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start monitor")
	timer := time.NewTicker(time.Second * 60)
	updateMonData()
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop monitor")
			return
		case <-timer.C:
			updateMonData()
		}
	}
}
