package polling

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingDNS(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("dns", pe, fmt.Errorf("node not found"))
		return
	}
	mode := pe.Mode
	if mode == "" {
		mode = "ipaddr"
	}
	target := pe.Params
	script := pe.Script
	ok := false
	var rTime int64
	var out []string
	var err error
	for i := 0; !ok && i <= pe.Retry; i++ {
		startTime := time.Now().UnixNano()
		if out, err = doLookup(mode, target); err != nil || len(out) < 1 {
			pe.Result["error"] = fmt.Sprintf("%v", err)
			continue
		}
		endTime := time.Now().UnixNano()
		rTime = endTime - startTime
		ok = true
		delete(pe.Result, "error")
	}
	if !ok {
		pe.Result["rtt"] = 0.0
		pe.Result["count"] = 0.0
		pe.Result["ip"] = ""
		setPollingState(pe, pe.Level)
		return
	}
	vm := otto.New()
	vm.Set("setResult", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() && call.Argument(1).IsNumber() {
			n := call.Argument(0).String()
			if v, err := call.Argument(1).ToFloat(); err == nil {
				pe.Result[n] = v
			}
		}
		return otto.Value{}
	})
	_ = vm.Set("rtt", rTime)
	_ = vm.Set("count", len(out))
	pe.Result["rtt"] = float64(rTime)
	pe.Result["count"] = float64(len(out))
	switch mode {
	case "ipaddr":
		oldip := ""
		if v, ok := pe.Result["ip"]; ok {
			if s, ok := v.(string); ok {
				oldip = s
			}
		}
		pe.Result["ip"] = out[0]
		if oldip != "" && oldip != pe.Result["ip"] {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	case "addr":
		vm.Set("addr", out)
		pe.Result["addr"] = strings.Join(out, ",")
	case "host":
		vm.Set("host", out)
		pe.Result["host"] = strings.Join(out, ",")
	case "mx":
		vm.Set("mx", out)
		pe.Result["mx"] = strings.Join(out, ",")
	case "ns":
		vm.Set("ns", out)
		pe.Result["ns"] = strings.Join(out, ",")
	case "txt":
		vm.Set("txt", out)
		pe.Result["txt"] = strings.Join(out, ",")
	case "cname":
		vm.Set("cname", out[0])
		pe.Result["cname"] = out[0]
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("dns", pe, fmt.Errorf("%v", err))
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}

func doLookup(mode, target string) ([]string, error) {
	ret := []string{}
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*50)
	defer cancel()
	switch mode {
	case "ipaddr":
		if addr, err := net.ResolveIPAddr("ip", target); err == nil {
			return []string{addr.String()}, nil
		} else {
			return ret, err
		}
	case "addr":
		return r.LookupAddr(ctx, target)
	case "host":
		return r.LookupHost(ctx, target)
	case "mx":
		if mxs, err := r.LookupMX(ctx, target); err == nil {
			for _, mx := range mxs {
				ret = append(ret, mx.Host)
			}
			return ret, nil
		} else {
			return ret, err
		}
	case "ns":
		if nss, err := r.LookupNS(ctx, target); err == nil {
			for _, ns := range nss {
				ret = append(ret, ns.Host)
			}
			return ret, nil
		} else {
			return ret, err
		}
	case "cname":
		if cname, err := r.LookupCNAME(ctx, target); err == nil {
			return []string{cname}, nil
		} else {
			return ret, err
		}
	case "txt":
		return r.LookupTXT(ctx, target)
	}
	return ret, nil
}
