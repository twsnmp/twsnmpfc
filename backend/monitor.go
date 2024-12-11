package backend

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
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
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	m.HeapAlloc = int64(ms.HeapAlloc)
	m.Sys = int64(ms.Sys)
	v, err := mem.VirtualMemory()
	if err == nil {
		m.Mem = v.UsedPercent
	}
	s, err := mem.SwapMemory()
	if err == nil {
		m.Swap = s.UsedPercent
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
	pid := os.Getpid()
	pr, err := process.NewProcess(int32(pid))
	if err == nil {
		if v, err := pr.CPUPercent(); err == nil {
			m.MyCPU = v
		}
		if v, err := pr.MemoryPercent(); err == nil {
			m.MyMem = float64(v)
		}
	}
	m.NumGoroutine = runtime.NumGoroutine()

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

func checkResourceAlert() {
	if len(datastore.MonitorDataes) < 1 {
		return
	}
	mem := []float64{}
	myMem := []float64{}
	load := []float64{}
	var disk float64
	for _, m := range datastore.MonitorDataes {
		mem = append(mem, m.Mem)
		myMem = append(myMem, m.MyMem)
		load = append(load, m.Load)
		disk = m.Disk
	}
	myMemMean, _ := stats.Mean(myMem)
	memMean, _ := stats.Mean(mem)
	loadMean, _ := stats.Mean(load)
	level := ""
	if myMemMean > 90.0 && memMean > 90.0 {
		level = "high"
	} else if myMemMean > 80.0 && memMean > 80.0 {
		level = "low"
	} else if myMemMean > 60.0 && memMean > 60.0 {
		level = "warn"
	}
	if level != "" {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: level,
			Event: fmt.Sprintf("メモリー不足 使用率:%0.2f%%", memMean),
		})
	}
	level = ""
	if disk > 95.0 {
		level = "high"
	} else if disk > 90.0 {
		level = "low"
	}
	if level != "" {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: level,
			Event: fmt.Sprintf("ストレージ容量不足 使用率:%0.2f%%", disk),
		})
	}
	level = ""
	if loadMean > float64(runtime.NumCPU()*4) {
		level = "high"
	} else if loadMean > float64(runtime.NumCPU()*2) {
		level = "low"
	} else if loadMean > float64(runtime.NumCPU()) {
		level = "warn"
	}
	if level != "" {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: level,
			Event: fmt.Sprintf("高負荷状態 CPUコア: %d,負荷:%0.2f%%", runtime.NumCPU(), loadMean),
		})
	}
}

// monitor :
func monitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start monitor")
	timer := time.NewTicker(time.Second * 60)
	updateMonData()
	defer timer.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop monitor")
			return
		case <-timer.C:
			updateMonData()
			i++
			if i%60 == 2 {
				// 1時間毎にリソースチェック
				checkResourceAlert()
			}
		}
	}
}
