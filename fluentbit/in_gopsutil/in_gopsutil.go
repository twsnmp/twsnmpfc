package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/input"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

var collection = ""
var diskPath = "/"

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	log.Printf("[in_gopsutil] Register called")
	return input.FLBPluginRegister(def, "gopsutil", "Input plugin for gopsutil")
}

// (fluentbit will call this)
// plugin (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	collection = input.FLBPluginConfigKey(plugin, "collection")
	diskPath = input.FLBPluginConfigKey(plugin, "disk_path")
	if diskPath == "" {
		diskPath = "/"
	}

	log.Printf("[in_gopsutil] collection= '%s'", collection)
	return input.FLB_OK
}

//export FLBPluginInputCallback
func FLBPluginInputCallback(data *unsafe.Pointer, size *C.size_t) int {
	flb_time := input.FLBTime{time.Now()}
	entry := []interface{}{flb_time}
	switch collection {
	case "cpu.percent.percpu":
		entry = append(entry, getCPUPercent(true))
	case "cpu.percent":
		entry = append(entry, getCPUPercent(false))
	case "cpu.times":
		if d, err := cpu.Times(false); err == nil && len(d) == 1 {
			entry = append(entry, d[0])
		}
	case "mem.vm":
		if d, err := mem.VirtualMemory(); err == nil {
			entry = append(entry, d)
		}
	case "mem.swap":
		if d, err := mem.SwapMemory(); err == nil {
			entry = append(entry, d)
		}
	case "host.info":
		if d, err := host.Info(); err == nil {
			entry = append(entry, d)
		}
	case "host.temp":
		entry = append(entry, getSensorTemp())
	case "load.avg":
		if d, err := load.Avg(); err == nil {
			entry = append(entry, d)
		}
	case "disk.usage":
		if d, err := disk.Usage(diskPath); err == nil {
			entry = append(entry, d)
		}
	default:
		log.Println("unknown collection")
		return input.FLB_ERROR
	}
	enc := input.NewEncoder()
	packed, err := enc.Encode(entry)
	if err != nil {
		log.Printf("[in_gopsutil] encode err:%v", err)
		return input.FLB_ERROR
	}

	length := len(packed)
	*data = C.CBytes(packed)
	*size = C.size_t(length)
	return input.FLB_OK
}

//export FLBPluginInputCleanupCallback
func FLBPluginInputCleanupCallback(data unsafe.Pointer) int {
	return input.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	log.Printf("[in_gopsutil] Exit called")
	return input.FLB_OK
}

// get CPU usage Percent
func getCPUPercent(percpu bool) map[string]interface{} {
	ret := make(map[string]interface{})
	if cpus, err := cpu.Percent(0, percpu); err == nil {
		if percpu {
			for i, v := range cpus {
				ret[fmt.Sprintf("cpu%d", i+1)] = v
			}
		} else if len(cpus) > 0 {
			ret["cpu"] = cpus[0]
		}
	}
	return ret
}

// get CPU adn hardware temperatures info
func getSensorTemp() map[string]interface{} {
	ret := make(map[string]interface{})
	if stl, err := host.SensorsTemperatures(); err == nil {
		for _, st := range stl {
			if st.Temperature != 0 {
				ret[st.SensorKey] = st.Temperature
			}
		}
	}
	return ret
}

func main() {
}
