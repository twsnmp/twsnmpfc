package polling

import (
	"fmt"
	"log"
	"strconv"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/lxi"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingLxi(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		return
	}
	if pe.Script == "" {
		setPollingError("lxi", pe, fmt.Errorf("lxi no script"))
		return
	}
	addr := pe.Params
	if addr == "" {
		addr = fmt.Sprintf("TCPIP0::%s::5025::SOCKET", n.IP)
	} else if p, err := strconv.Atoi(addr); err == nil && p > 0 && p < 65535 {
		addr = fmt.Sprintf("TCPIP0::%s::%d::SOCKET", n.IP, p)
	}
	d, err := lxi.NewDevice(addr, pe.Timeout*1000)
	if err != nil {
		setPollingError("lxi", pe, err)
		return
	}
	defer d.Close()
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	pe.Result = make(map[string]interface{})
	vm.Set("lxiCommand", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() {
			f := call.Argument(0).String()
			args := []interface{}{}
			for i := 1; i < len(call.ArgumentList); i++ {
				if call.Argument(i).IsNumber() {
					if v, err := call.Argument(1).ToFloat(); err == nil {
						args = append(args, v)
					}
				} else if call.Argument(i).IsString() {
					args = append(args, call.Argument(i).String())
				}
			}
			if err := d.Command(f, args...); err != nil {
				log.Printf("lxi command err=%v", err)
				if ov, err := otto.ToValue(false); err == nil {
					return ov
				}
			} else {
				if ov, err := otto.ToValue(true); err == nil {
					return ov
				}
			}
		}
		return otto.Value{}
	})
	vm.Set("lxiQuery", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() {
			q := call.Argument(0).String()
			to := pe.Timeout * 1000
			if call.Argument(1).IsNumber() {
				if i, err := call.Argument(1).ToInteger(); err == nil && i > 0 && i < 10*1000 {
					to = int(i)
				}
			}
			d.SetTimeout(to)
			if v, err := d.Query(q); err == nil {
				if ov, err := otto.ToValue(v); err == nil {
					return ov
				}
			} else {
				log.Printf("lxi query err=%v", err)
			}
		}
		return otto.UndefinedValue()
	})
	value, err := vm.Run(pe.Script)
	if err == nil {
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	}
	log.Printf("lxi polling err=%v", err)
	setPollingError("lxi", pe, err)
}
