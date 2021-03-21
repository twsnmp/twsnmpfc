package polling

// VMWare ESXiのポーリング

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/robertkrimen/otto"

	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
)

func doPollingVMWare(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found nodeID=%s", pe.NodeID)
		return
	}
	mode := pe.Mode
	target := pe.Params
	script := pe.Script
	us := n.URL
	if us == "" {
		us = fmt.Sprintf("https://%s:%s@%s/sdk", n.User, n.Password, n.IP)
	}
	if !strings.Contains(us, "/sdk") {
		us += "/sdk"
	}
	u, err := soap.ParseURL(us)
	if err != nil {
		setPollingError("vmware", pe, fmt.Errorf("invalid url"))
		return
	}
	if u.User == nil || u.User.String() == ":" {
		u.User = url.UserPassword(n.User, n.Password)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pe.Timeout)*time.Second)
	defer cancel()
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		setPollingError("vmware", pe, err)
		return
	}
	var rMap = make(map[string]float64)
	switch mode {
	case "HostSystem":
		rMap, err = vmwareHostSystem(ctx, client.Client, target)
	case "Datastore":
		rMap, err = vmwareDatastore(ctx, client.Client, target)
	case "VirtualMachine":
		rMap, err = vmwareVirtualMachine(ctx, client.Client, target)
	}
	if err != nil {
		setPollingError("vmware", pe, err)
		return
	}
	vm := otto.New()
	for k, v := range rMap {
		vm.Set(k, v)
		pe.Result[k] = v
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("vmware", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); !ok {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "normal")
}

func vmwareHostSystem(ctx context.Context, c *vim25.Client, target string) (map[string]float64, error) {
	r := make(map[string]float64)
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return r, err
	}
	defer func() {
		_ = v.Destroy(ctx)
	}()
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		return r, err
	}
	r["totalCPU"] = 0.0
	r["totalMEM"] = 0.0
	r["usedCPU"] = 0.0
	r["usedMEM"] = 0.0
	r["totalHost"] = 0.0
	for _, hs := range hss {
		if target != "" && target != hs.Summary.Config.Name {
			continue
		}
		totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
		r["totalCPU"] += float64(totalCPU)
		r["usedCPU"] += float64(hs.Summary.QuickStats.OverallCpuUsage)
		r["totalMEM"] += float64(hs.Summary.Hardware.MemorySize)
		r["usedMEM"] += float64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024
		r["totalHost"] += 1.0
	}
	if r["totalCPU"] > 0.0 {
		r["usageCPU"] = 100.0 * r["usedCPU"] / r["totalCPU"]
	} else {
		r["usageCPU"] = 0.0
	}
	if r["totalMEM"] > 0.0 {
		r["usageMEM"] = 100.0 * r["usedMEM"] / r["totalMEM"]
	} else {
		r["usageMEM"] = 0.0
	}
	return r, nil
}

func vmwareDatastore(ctx context.Context, c *vim25.Client, target string) (map[string]float64, error) {
	r := make(map[string]float64)
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		return r, err
	}
	defer func() {
		_ = v.Destroy(ctx)
	}()
	var dss []mo.Datastore
	err = v.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	if err != nil {
		return r, err
	}
	r["capacity"] = 0.0
	r["freeSpace"] = 0.0
	r["total"] = 0.0
	for _, ds := range dss {
		if target != "" && target != ds.Summary.Name {
			continue
		}
		r["capacity"] += float64(ds.Summary.Capacity)
		r["freeSpace"] += float64(ds.Summary.FreeSpace)
		r["total"] += 1.0
	}
	if r["capacity"] > 0.0 {
		r["usage"] = 100.0 * (r["capacity"] - r["freeSpace"]) / r["capacity"]
	} else {
		r["usage"] = 0.0
	}
	return r, nil
}

func vmwareVirtualMachine(ctx context.Context, c *vim25.Client, target string) (map[string]float64, error) {
	r := make(map[string]float64)
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return r, err
	}
	defer func() {
		_ = v.Destroy(ctx)
	}()
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return r, err
	}
	r["up"] = 0.0
	r["total"] = 0.0
	r["rate"] = 0.0
	for _, vm := range vms {
		if target != "" && target != vm.Summary.Config.Name {
			continue
		}
		if vm.Summary.Runtime.PowerState == "poweredOn" {
			r["up"] += 1.0
		}
		r["total"] += 1.0
	}
	if r["total"] > 0.0 {
		r["rate"] = 100.0 * r["up"] / r["total"]
	}
	return r, nil
}
