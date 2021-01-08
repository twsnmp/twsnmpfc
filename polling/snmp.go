package polling

// SNMPのポーリング処理

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	gosnmp "github.com/twsnmp/gosnmp"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func (p *Polling) doPollingSnmp(pe *datastore.PollingEnt) {
	n := p.ds.GetNode(pe.NodeID)
	if n == nil {
		log.Printf("node not found nodeID=%s", pe.NodeID)
		return
	}
	agent := &gosnmp.GoSNMP{
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(pe.Timeout) * time.Second,
		Retries:            pe.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	if n.SnmpMode != "" {
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		if n.SnmpMode == "v3auth" {
			agent.MsgFlags = gosnmp.AuthNoPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 n.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: n.Password,
			}
		} else {
			agent.MsgFlags = gosnmp.AuthPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 n.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: n.Password,
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        n.Password,
			}
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("SNMP agent.Connect err=%v", err)
		return
	}
	defer agent.Conn.Close()
	mode, params := parseSnmpPolling(pe.Polling)
	if mode == "" {
		p.setPollingError("snmp", pe, fmt.Errorf("invalid snmp polling"))
		return
	}
	if mode == "sysUpTime" {
		p.doPollingSnmpSysUpTime(pe, agent)
	} else if strings.HasPrefix(mode, "ifOperStatus.") {
		p.doPollingSnmpIF(pe, mode, agent)
	} else if mode == "count" {
		p.doPollingSnmpCount(pe, params, agent)
	} else if mode == "process" {
		p.doPollingSnmpProcess(pe, params, agent)
	} else if mode == "stats" {
		p.doPollingSnmpStats(pe, params, agent)
	} else {
		p.doPollingSnmpGet(pe, mode, params, agent)
	}
}

func parseSnmpPolling(s string) (string, string) {
	a := strings.SplitN(s, "|", 2)
	if len(a) < 1 {
		return "", ""
	}
	if len(a) < 2 {
		return strings.TrimSpace(a[0]), ""
	}
	return strings.TrimSpace(a[0]), strings.TrimSpace(a[1])
}

func (p *Polling) doPollingSnmpSysUpTime(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	oids := []string{p.ds.MIBDB.NameToOID("sysUpTime.0")}
	result, err := agent.Get(oids)
	if err != nil {
		p.setPollingError("snmpUpTime", pe, err)
		return
	}
	var uptime int64
	for _, variable := range result.Variables {
		if variable.Name == p.ds.MIBDB.NameToOID("sysUpTime.0") {
			uptime = int64(gosnmp.ToBigInt(variable.Value).Uint64())
			break
		}
	}
	if uptime == 0 {
		p.setPollingError("snmpUpTime", pe, fmt.Errorf("uptime==0"))
		return
	}
	lr := make(map[string]string)
	json.Unmarshal([]byte(pe.LastResult), &lr)
	if lut, ok := lr["sysUpTime"]; ok {
		lastUptime, err := strconv.ParseInt(lut, 10, 64)
		if err != nil {
			delete(lr, "sysUpTime")
			pe.LastResult = makeLastResult(lr)
			p.setPollingError("snmp", pe, err)
			return
		}
		pe.LastVal = float64(uptime - lastUptime)
		lr["sysUpTime"] = fmt.Sprintf("%d", uptime)
		pe.LastResult = makeLastResult(lr)
		if lastUptime < uptime {
			p.setPollingState(pe, "normal")
			return
		}
		p.setPollingState(pe, pe.Level)
		return
	}
	pe.LastVal = 0.0
	lr["sysUpTime"] = fmt.Sprintf("%d", uptime)
	pe.LastResult = makeLastResult(lr)
	p.setPollingState(pe, "unknown")
}

func (p *Polling) doPollingSnmpIF(pe *datastore.PollingEnt, ps string, agent *gosnmp.GoSNMP) {
	a := strings.Split(ps, ".")
	if len(a) < 2 {
		p.setPollingError("snmpif", pe, fmt.Errorf("invalid format"))
		return
	}
	oids := []string{p.ds.MIBDB.NameToOID("ifOperStatus." + a[1]), p.ds.MIBDB.NameToOID("ifAdminStatus." + a[1])}
	result, err := agent.Get(oids)
	if err != nil {
		p.setPollingError("snmpif", pe, err)
		return
	}
	var oper int64
	var admin int64
	for _, variable := range result.Variables {
		if strings.HasPrefix(p.ds.MIBDB.OIDToName(variable.Name), "ifOperStatus") {
			oper = gosnmp.ToBigInt(variable.Value).Int64()
		} else if strings.HasPrefix(p.ds.MIBDB.OIDToName(variable.Name), "ifAdminStatus") {
			admin = gosnmp.ToBigInt(variable.Value).Int64()
		}
	}
	lr := make(map[string]string)
	pe.LastVal = float64(oper)
	lr["oper"] = fmt.Sprintf("%d", oper)
	lr["admin"] = fmt.Sprintf("%d", admin)
	pe.LastResult = makeLastResult(lr)
	if oper == 1 {
		p.setPollingState(pe, "normal")
		return
	} else if admin == 2 {
		p.setPollingState(pe, "normal")
		return
	} else if oper == 2 && admin == 1 {
		p.setPollingState(pe, pe.Level)
		return
	}
	p.setPollingState(pe, "unknown")
}

func (p *Polling) doPollingSnmpGet(pe *datastore.PollingEnt, mode, params string, agent *gosnmp.GoSNMP) {
	a := strings.Split(params, "|")
	if len(a) < 2 {
		p.setPollingError("snmp", pe, fmt.Errorf("invalid format"))
		return
	}
	names := strings.Split(a[0], ",")
	script := a[1]
	oids := []string{}
	for _, n := range names {
		if n == "" {
			continue
		}
		if oid := p.ds.MIBDB.NameToOID(n); oid != "" {
			oids = append(oids, strings.TrimSpace(oid))
		}
	}
	if len(oids) < 1 {
		p.setPollingError("snmp", pe, fmt.Errorf("invalid format"))
		return
	}
	if mode == "ps" {
		oids = append(oids, p.ds.MIBDB.NameToOID("sysUpTime.0"))
	}
	result, err := agent.Get(oids)
	if err != nil {
		p.setPollingError("snmp", pe, err)
		return
	}
	vm := otto.New()
	lr := make(map[string]string)
	for _, variable := range result.Variables {
		if variable.Name == p.ds.MIBDB.NameToOID("sysUpTime.0") {
			sut := gosnmp.ToBigInt(variable.Value).Uint64()
			_ = vm.Set("sysUpTime", sut)
			lr["sysUpTime.0"] = fmt.Sprintf("%d", sut)
			if mode == "ps" || mode == "delta" {
				lr["sysUpTime.0_Last"] = fmt.Sprintf("%d", sut)
			}
			continue
		}
		n := p.ds.MIBDB.OIDToName(variable.Name)
		vn := getValueName(n)
		if variable.Type == gosnmp.OctetString {
			v := variable.Value.(string)
			_ = vm.Set(vn, v)
			lr[n] = v
		} else if variable.Type == gosnmp.ObjectIdentifier {
			v := p.ds.MIBDB.OIDToName(variable.Value.(string))
			_ = vm.Set(vn, v)
			lr[n] = v
		} else {
			v := gosnmp.ToBigInt(variable.Value).Uint64()
			_ = vm.Set(vn, v)
			lr[n] = fmt.Sprintf("%d", v)
			if mode == "ps" || mode == "delta" {
				lr[n+"_Last"] = lr[n]
			}
		}
	}
	if mode == "ps" || mode == "delta" {
		oldlr := make(map[string]string)
		if err := json.Unmarshal([]byte(pe.LastResult), &oldlr); err != nil || oldlr["error"] != "" {
			pe.LastResult = makeLastResult(lr)
			p.setPollingState(pe, "unknown")
			return
		}
		nvmap := make(map[string]int64)
		for k, v := range lr {
			if strings.HasPrefix(k, "_Last") {
				continue
			}
			if vo, ok := oldlr[k+"_Last"]; ok {
				if nv, err := strconv.ParseInt(v, 10, 64); err == nil {
					if nvo, err := strconv.ParseInt(vo, 10, 64); err == nil {
						nvmap[k] = nv - nvo
					}
				}
			}
		}
		sut := float64(1.0)
		if mode == "ps" {
			v, ok := nvmap["sysUpTime.0"]
			if !ok || v == 0 {
				p.setPollingError("snmp", pe, fmt.Errorf("invalid format %v", nvmap))
				return
			}
			sut = float64(v)
		}
		for k, v := range nvmap {
			lr[k] = fmt.Sprintf("%f", float64(v*100.0)/sut)
			vn := getValueName(k)
			_ = vm.Set(vn, float64(v*100.0)/sut)
		}
	}
	value, err := vm.Run(script)
	if err == nil {
		if v, err := vm.Get("numVal"); err == nil {
			if v.IsNumber() {
				if vf, err := v.ToFloat(); err == nil {
					pe.LastVal = vf
				}
			}
		}
		pe.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); !ok {
			p.setPollingState(pe, pe.Level)
			return
		}
		p.setPollingState(pe, "normal")
		return
	}
	p.setPollingError("snmp", pe, err)
}

func getValueName(n string) string {
	a := strings.SplitN(n, ".", 2)
	return (a[0])
}

func (p *Polling) doPollingSnmpCount(pe *datastore.PollingEnt, params string, agent *gosnmp.GoSNMP) {
	cmds := splitCmd(params)
	if len(cmds) < 3 {
		p.setPollingError("snmp", pe, fmt.Errorf("invalid format"))
		return
	}
	oid := p.ds.MIBDB.NameToOID(cmds[0])
	filter := datastore.ParseFilter(cmds[1])
	script := cmds[2]
	count := 0
	var regexFilter *regexp.Regexp
	var err error
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			log.Printf("doPollingSnmpCount err=%v", err)
			regexFilter = nil
		}
	}
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		s := ""
		if variable.Type == gosnmp.OctetString {
			if strings.Contains(p.ds.MIBDB.OIDToName(variable.Name), "ifPhysAd") {
				a := variable.Value.(string)
				if len(a) > 5 {
					s = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", a[0], a[1], a[2], a[3], a[4], a[5])
				}
			} else {
				s = variable.Value.(string)
			}
		} else if variable.Type == gosnmp.ObjectIdentifier {
			s = p.ds.MIBDB.OIDToName(variable.Value.(string))
		} else {
			s = fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Uint64())
		}
		if regexFilter != nil && !regexFilter.Match([]byte(s)) {
			return nil
		}
		count++
		return nil
	}); err != nil {
		p.setPollingError("snmp", pe, err)
		return
	}
	vm := otto.New()
	lr := make(map[string]string)
	_ = vm.Set("count", count)
	lr["count"] = fmt.Sprintf("%d", count)
	value, err := vm.Run(script)
	if err == nil {
		pe.LastVal = float64(count)
		if v, err := vm.Get("numVal"); err == nil {
			if v.IsNumber() {
				if vf, err := v.ToFloat(); err == nil {
					pe.LastVal = vf
				}
			}
		}
		pe.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); !ok {
			p.setPollingState(pe, pe.Level)
			return
		}
		p.setPollingState(pe, "normal")
		return
	}
	p.setPollingError("snmp", pe, err)
}

func (p *Polling) doPollingSnmpProcess(pe *datastore.PollingEnt, params string, agent *gosnmp.GoSNMP) {
	cmds := splitCmd(params)
	if len(cmds) < 2 {
		p.setPollingError("snmp", pe, fmt.Errorf("doPollingSnmpProcess Invalid format"))
		return
	}
	oid := p.ds.MIBDB.NameToOID("hrSWRunName")
	filter := datastore.ParseFilter(cmds[0])
	script := cmds[1]
	var regexFilter *regexp.Regexp
	var err error
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			log.Printf("doPollingSnmpProcess err=%v", err)
			regexFilter = nil
		}
	}
	lastPidSum := 0
	lr := make(map[string]string)
	if err := json.Unmarshal([]byte(pe.LastResult), &lr); err == nil {
		if s, ok := lr["pidSum"]; ok {
			if n, err := strconv.Atoi(s); err == nil {
				lastPidSum = n
			}
		}
	}
	pidSum := 0
	count := 0
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		if variable.Type != gosnmp.OctetString {
			return nil
		}
		n := p.ds.MIBDB.OIDToName(variable.Name)
		a := strings.SplitN(n, ".", 2)
		s := variable.Value.(string)
		if len(a) != 2 || a[0] != "hrSWRunName" {
			return nil
		}
		pid, err := strconv.Atoi(a[1])
		if err != nil {
			return nil
		}
		if regexFilter != nil && !regexFilter.Match([]byte(s)) {
			return nil
		}
		pidSum += pid
		count++
		return nil
	}); err != nil {
		p.setPollingError("snmp", pe, err)
		return
	}
	changed := 0
	if lastPidSum != 0 && pidSum != lastPidSum {
		changed = 1
	}
	vm := otto.New()
	_ = vm.Set("count", count)
	_ = vm.Set("changed", changed)
	lr["count"] = fmt.Sprintf("%d", count)
	lr["pidSum"] = fmt.Sprintf("%d", pidSum)
	lr["changed"] = fmt.Sprintf("%d", changed)
	value, err := vm.Run(script)
	if err == nil {
		pe.LastVal = float64(count)
		if v, err := vm.Get("numVal"); err == nil {
			if v.IsNumber() {
				if vf, err := v.ToFloat(); err == nil {
					pe.LastVal = vf
				}
			}
		}
		pe.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); !ok {
			p.setPollingState(pe, pe.Level)
			return
		}
		p.setPollingState(pe, "normal")
		return
	}
	p.setPollingError("snmp", pe, err)
}

func (p *Polling) doPollingSnmpStats(pe *datastore.PollingEnt, params string, agent *gosnmp.GoSNMP) {
	cmds := splitCmd(params)
	if len(cmds) < 2 {
		p.setPollingError("snmp", pe, fmt.Errorf("invalid format"))
		return
	}
	oid := p.ds.MIBDB.NameToOID(cmds[0])
	script := cmds[1]
	count := uint64(0)
	sum := uint64(0)
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		if variable.Type != gosnmp.Counter32 &&
			variable.Type != gosnmp.Counter64 &&
			variable.Type != gosnmp.Integer &&
			variable.Type != gosnmp.Uinteger32 &&
			variable.Type != gosnmp.Gauge32 {
			return fmt.Errorf("mib is not number %#v", variable)
		}
		sum += gosnmp.ToBigInt(variable.Value).Uint64()
		count++
		return nil
	}); err != nil {
		p.setPollingError("snmp", pe, err)
		return
	}
	if count < 1 {
		p.setPollingError("snmp", pe, fmt.Errorf("no data"))
		return
	}
	avg := float64(sum) / float64(count)
	vm := otto.New()
	lr := make(map[string]string)
	_ = vm.Set("count", count)
	_ = vm.Set("sum", sum)
	_ = vm.Set("avg", avg)
	lr["count"] = fmt.Sprintf("%d", count)
	lr["sum"] = fmt.Sprintf("%d", sum)
	lr["avg"] = fmt.Sprintf("%f", avg)
	value, err := vm.Run(script)
	if err == nil {
		pe.LastVal = float64(avg)
		if v, err := vm.Get("numVal"); err == nil {
			if v.IsNumber() {
				if vf, err := v.ToFloat(); err == nil {
					pe.LastVal = vf
				}
			}
		}
		pe.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); !ok {
			p.setPollingState(pe, pe.Level)
			return
		}
		p.setPollingState(pe, "normal")
		return
	}
	p.setPollingError("snmp", pe, err)
}
