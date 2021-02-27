package polling

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingDNS(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("dns", pe, fmt.Errorf("node not found"))
		return
	}
	cmds := splitCmd(pe.Polling)
	mode := "ipaddr"
	target := pe.Polling
	script := ""
	if len(cmds) == 3 {
		mode = cmds[0]
		target = cmds[1]
		script = cmds[2]
	}
	ok := false
	var rTime int64
	var out []string
	var err error
	lr := make(map[string]string)
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		if out, err = doLookup(mode, target); err != nil || len(out) < 1 {
			lr["error"] = fmt.Sprintf("%v", err)
			log.Printf("doLookup err=%v %v", err, cmds)
			continue
		}
		endTime := time.Now().UnixNano()
		rTime = endTime - startTime
		ok = true
		delete(lr, "error")
	}
	oldlr := make(map[string]string)
	_ = json.Unmarshal([]byte(pe.LastResult), &oldlr)
	if !ok {
		for k, v := range oldlr {
			if k != "error" {
				lr[k] = v
			}
		}
		pe.LastResult = makeLastResult(lr)
		pe.LastVal = 0.0
		setPollingState(pe, pe.Level)
		return
	}
	pe.LastVal = float64(rTime)
	vm := otto.New()
	_ = vm.Set("rtt", pe.LastVal)
	_ = vm.Set("count", len(out))
	lr["rtt"] = fmt.Sprintf("%f", pe.LastVal)
	lr["count"] = fmt.Sprintf("%d", len(out))
	switch mode {
	case "ipaddr":
		lr["ip"] = out[0]
		pe.LastResult = makeLastResult(lr)
		if oldlr["ip"] != "" && oldlr["ip"] != lr["ip"] {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	case "addr":
		_ = vm.Set("addr", out)
		lr["addr"] = strings.Join(out, ",")
	case "host":
		_ = vm.Set("host", out)
		lr["host"] = strings.Join(out, ",")
	case "mx":
		_ = vm.Set("mx", out)
		lr["mx"] = strings.Join(out, ",")
	case "ns":
		_ = vm.Set("ns", out)
		lr["ns"] = strings.Join(out, ",")
	case "txt":
		_ = vm.Set("txt", out)
		lr["txt"] = strings.Join(out, ",")
	case "cname":
		_ = vm.Set("cname", out[0])
		lr["cname"] = out[0]
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("dns", pe, fmt.Errorf("%v", err))
		return
	}
	pe.LastResult = makeLastResult(lr)
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}

func doLookup(mode, target string) ([]string, error) {
	ret := []string{}
	switch mode {
	case "ipaddr":
		if addr, err := net.ResolveIPAddr("ip", target); err == nil {
			return []string{addr.String()}, nil
		} else {
			return ret, err
		}
	case "addr":
		return net.LookupAddr(target)
	case "host":
		return net.LookupHost(target)
	case "mx":
		if mxs, err := net.LookupMX(target); err == nil {
			for _, mx := range mxs {
				ret = append(ret, mx.Host)
			}
			return ret, nil
		} else {
			return ret, err
		}
	case "ns":
		if nss, err := net.LookupNS(target); err == nil {
			for _, ns := range nss {
				ret = append(ret, ns.Host)
			}
			return ret, nil
		} else {
			return ret, err
		}
	case "cname":
		if cname, err := net.LookupCNAME(target); err == nil {
			return []string{cname}, nil
		} else {
			return ret, err
		}
	case "txt":
		return net.LookupTXT(target)
	}
	return ret, nil
}
