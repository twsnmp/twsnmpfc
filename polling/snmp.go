package polling

// SNMPのポーリング処理

import (
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

func doPollingSnmp(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
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
	mode := pe.Mode
	if mode == "" {
		setPollingError("snmp", pe, fmt.Errorf("invalid snmp polling"))
		return
	}
	if mode == "sysUpTime" {
		doPollingSnmpSysUpTime(pe, agent)
	} else if mode == "ifOperStatus" {
		doPollingSnmpIF(pe, agent)
	} else if mode == "count" {
		doPollingSnmpCount(pe, agent)
	} else if mode == "process" {
		doPollingSnmpProcess(pe, agent)
	} else if mode == "stats" {
		doPollingSnmpStats(pe, agent)
	} else {
		doPollingSnmpGet(pe, agent)
	}
}

func doPollingSnmpSysUpTime(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	oids := []string{datastore.MIBDB.NameToOID("sysUpTime.0")}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingError("snmpUpTime", pe, err)
		return
	}
	var uptime int64
	for _, variable := range result.Variables {
		if variable.Name == datastore.MIBDB.NameToOID("sysUpTime.0") {
			uptime = int64(gosnmp.ToBigInt(variable.Value).Uint64())
			break
		}
	}
	if uptime == 0 {
		setPollingError("snmpUpTime", pe, fmt.Errorf("uptime==0"))
		return
	}
	if v, ok := pe.Result["sysUpTime"]; ok {
		lastUptime, ok := v.(float64)
		if !ok {
			delete(pe.Result, "sysUpTime")
			setPollingError("snmp", pe, fmt.Errorf("sysUpTime not floate64"))
			return

		}
		diff := float64(uptime) - lastUptime
		pe.Result["sysUpTime"] = float64(uptime)
		pe.Result["deltaSysUpTime"] = diff
		if lastUptime < float64(uptime) {
			setPollingState(pe, "normal")
			return
		}
		setPollingState(pe, pe.Level)
		return
	}
	pe.Result["sysUpTime"] = float64(uptime)
	pe.Result["deltaSysUpTime"] = 0.0
	setPollingState(pe, "unknown")
}

func doPollingSnmpIF(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	if pe.Params == "" {
		setPollingError("snmpif", pe, fmt.Errorf("invalid format"))
		return
	}
	oids := []string{datastore.MIBDB.NameToOID("ifOperStatus." + pe.Params), datastore.MIBDB.NameToOID("ifAdminStatus." + pe.Params)}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingError("snmpif", pe, err)
		return
	}
	var oper int64
	var admin int64
	for _, variable := range result.Variables {
		if strings.HasPrefix(datastore.MIBDB.OIDToName(variable.Name), "ifOperStatus") {
			oper = gosnmp.ToBigInt(variable.Value).Int64()
		} else if strings.HasPrefix(datastore.MIBDB.OIDToName(variable.Name), "ifAdminStatus") {
			admin = gosnmp.ToBigInt(variable.Value).Int64()
		}
	}
	pe.Result["ifOperStatus"] = float64(oper)
	pe.Result["ifAdminStatus"] = float64(admin)
	if oper == 1 {
		setPollingState(pe, "normal")
		return
	} else if admin == 2 {
		setPollingState(pe, "normal")
		return
	} else if oper == 2 && admin == 1 {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "unknown")
}

func doPollingSnmpGet(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	names := strings.Split(pe.Params, ",")
	script := pe.Script
	mode := pe.Mode
	oids := []string{}
	for _, n := range names {
		if n == "" {
			continue
		}
		if oid := datastore.MIBDB.NameToOID(n); oid != "" {
			oids = append(oids, strings.TrimSpace(oid))
		}
	}
	if len(oids) < 1 {
		setPollingError("snmp", pe, fmt.Errorf("invalid format"))
		return
	}
	if mode == "ps" {
		oids = append(oids, datastore.MIBDB.NameToOID("sysUpTime.0"))
	}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingError("snmp", pe, err)
		return
	}
	vm := otto.New()
	lr := make(map[string]interface{})
	for _, variable := range result.Variables {
		if variable.Name == datastore.MIBDB.NameToOID("sysUpTime.0") {
			sut := gosnmp.ToBigInt(variable.Value).Uint64()
			vm.Set("sysUpTime", sut)
			lr["sysUpTime"] = float64(sut)
			continue
		}
		n := datastore.MIBDB.OIDToName(variable.Name)
		vn := getValueName(n)
		if variable.Type == gosnmp.OctetString {
			v := variable.Value.(string)
			vm.Set(vn, v)
			lr[n] = v
		} else if variable.Type == gosnmp.ObjectIdentifier {
			v := datastore.MIBDB.OIDToName(variable.Value.(string))
			vm.Set(vn, v)
			lr[n] = v
		} else {
			v := gosnmp.ToBigInt(variable.Value).Uint64()
			vm.Set(vn, v)
			lr[n] = float64(v)
		}
	}
	if mode == "ps" || mode == "delta" {
		if _, ok := pe.Result["lastTime"]; !ok {
			lr["lastTime"] = float64(time.Now().UnixNano())
			pe.Result = lr
			setPollingState(pe, "unknown")
			return
		}
		for k, v := range lr {
			if vf, ok := v.(float64); ok {
				if vo, ok := pe.Result[k]; ok {
					if vof, ok := vo.(float64); ok {
						d := vf - vof
						lr[k+"_Delta"] = d
						vm.Set(k+"_Delta", d)
					}
				}
			}
		}
		if mode == "ps" {
			var diff float64
			if v, ok := lr["sysUpTime_Delta"]; ok {
				if vf, ok := v.(float64); ok {
					diff = vf
				}
				if diff < 1.0 {
					setPollingError("snmp", pe, fmt.Errorf("no sysUptime"))
					return
				}
				for k, v := range lr {
					if strings.HasPrefix(k, "_Delta") {
						continue
					}
					if _, ok := v.(float64); ok {
						if vd, ok := lr[k+"_Delta"]; ok {
							if vdf, ok := vd.(float64); ok {
								lr[k+"_PS"] = float64((vdf * 100.0) / diff)
								vm.Set(k+"_PS", float64((vdf*100.0)/diff))
							}
						}
					}
				}
			}
		}
		value, err := vm.Run(script)
		if err == nil {
			if ok, _ := value.ToBoolean(); !ok {
				setPollingState(pe, pe.Level)
				return
			}
			setPollingState(pe, "normal")
			return
		}
		setPollingError("snmp", pe, err)
	}
}

func getValueName(n string) string {
	a := strings.SplitN(n, ".", 2)
	return (a[0])
}

func doPollingSnmpCount(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	oid := datastore.MIBDB.NameToOID(pe.Params)
	filter := pe.Filter
	script := pe.Script
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
			if strings.Contains(datastore.MIBDB.OIDToName(variable.Name), "ifPhysAd") {
				a := variable.Value.(string)
				if len(a) > 5 {
					s = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", a[0], a[1], a[2], a[3], a[4], a[5])
				}
			} else {
				s = variable.Value.(string)
			}
		} else if variable.Type == gosnmp.ObjectIdentifier {
			s = datastore.MIBDB.OIDToName(variable.Value.(string))
		} else {
			s = fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Uint64())
		}
		if regexFilter != nil && !regexFilter.Match([]byte(s)) {
			return nil
		}
		count++
		return nil
	}); err != nil {
		setPollingError("snmp", pe, err)
		return
	}
	vm := otto.New()
	vm.Set("count", count)
	pe.Result["count"] = float64(count)
	value, err := vm.Run(script)
	if err == nil {
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	}
	setPollingError("snmp", pe, err)
}

func doPollingSnmpProcess(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	oid := datastore.MIBDB.NameToOID("hrSWRunName")
	filter := pe.Filter
	script := pe.Script
	var regexFilter *regexp.Regexp
	var err error
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			log.Printf("doPollingSnmpProcess err=%v", err)
			regexFilter = nil
		}
	}
	lastPidSum := 0.0
	if v, ok := pe.Result["pidSum"]; ok {
		if vf, ok := v.(float64); ok {
			lastPidSum = vf
		}
	}
	pidSum := 0.0
	count := 0
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		if variable.Type != gosnmp.OctetString {
			return nil
		}
		n := datastore.MIBDB.OIDToName(variable.Name)
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
		pidSum += float64(pid)
		count++
		return nil
	}); err != nil {
		setPollingError("snmp", pe, err)
		return
	}
	changed := 0
	if lastPidSum != 0 && pidSum != lastPidSum {
		changed = 1
	}
	vm := otto.New()
	vm.Set("count", count)
	vm.Set("changed", changed)
	pe.Result["count"] = float64(count)
	pe.Result["pidSum"] = float64(pidSum)
	pe.Result["changed"] = float64(changed)
	value, err := vm.Run(script)
	if err == nil {
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	}
	setPollingError("snmp", pe, err)
}

func doPollingSnmpStats(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	oid := datastore.MIBDB.NameToOID(pe.Params)
	script := pe.Script
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
		setPollingError("snmp", pe, err)
		return
	}
	if count < 1 {
		setPollingError("snmp", pe, fmt.Errorf("no data"))
		return
	}
	avg := float64(sum) / float64(count)
	vm := otto.New()
	vm.Set("count", count)
	vm.Set("sum", sum)
	vm.Set("avg", avg)
	pe.Result["count"] = float64(count)
	pe.Result["sum"] = float64(sum)
	pe.Result["avg"] = float64(avg)
	value, err := vm.Run(script)
	if err == nil {
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	}
	setPollingError("snmp", pe, err)
}
