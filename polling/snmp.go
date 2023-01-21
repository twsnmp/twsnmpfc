package polling

// SNMPのポーリング処理

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	gosnmp "github.com/gosnmp/gosnmp"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func doPollingSnmp(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
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
	switch n.SnmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        n.Password,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        n.Password,
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("polling snnmp err=%v", err)
		return
	}
	defer agent.Conn.Close()
	mode := pe.Mode
	if mode == "" {
		setPollingError("snmp", pe, fmt.Errorf("invalid snmp polling"))
		return
	}
	switch mode {
	case "sysUpTime":
		doPollingSnmpSysUpTime(pe, agent)
	case "ifOperStatus":
		doPollingSnmpIF(pe, agent)
	case "hrSystemDate":
		doPollingSnmpSystemDate(pe, agent)
	case "count":
		doPollingSnmpCount(pe, agent)
	case "process":
		doPollingSnmpProcess(pe, agent)
	case "stats":
		doPollingSnmpStats(pe, agent)
	case "traffic":
		doPollingSnmpTraffic(pe, agent)
	default:
		doPollingSnmpGet(pe, agent)
	}
}

func doPollingSnmpSysUpTime(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	oids := []string{datastore.MIBDB.NameToOID("sysUpTime.0")}
	result, err := agent.Get(oids)
	if err != nil {
		pe.Result["error"] = fmt.Sprintf("%v", err)
		setPollingState(pe, pe.Level)
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
		pe.Result["error"] = fmt.Sprintf("%v", fmt.Errorf("uptime==0"))
		setPollingState(pe, pe.Level)
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
		pe.Result["error"] = ""
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
		switch variable.Type {

		case gosnmp.OctetString:
			v := getMIBStringVal(variable.Value)
			mi := datastore.FindMIBInfo(n)
			if mi != nil {
				switch mi.Type {
				case "PhysAddress", "OctetString":
					a, ok := variable.Value.([]uint8)
					if !ok {
						a = []uint8(getMIBStringVal(variable.Value))
					}
					mac := []string{}
					for _, m := range a {
						mac = append(mac, fmt.Sprintf("%02X", m&0x00ff))
					}
					v = strings.Join(mac, ":")
				case "DisplayString":
				default:
				}
			}
			vm.Set(vn, v)
			lr[n] = v
		case gosnmp.ObjectIdentifier:
			v := datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
			vm.Set(vn, v)
			lr[n] = v
		default:
			v := gosnmp.ToBigInt(variable.Value).Uint64()
			vm.Set(vn, v)
			lr[n] = float64(v)
		}
	}
	keys := []string{}
	for k := range lr {
		keys = append(keys, k)
	}
	if mode == "ps" || mode == "delta" {
		if _, ok := pe.Result["lastTime"]; !ok {
			lr["lastTime"] = float64(time.Now().UnixNano())
			pe.Result = lr
			setPollingState(pe, "unknown")
			return
		}
		lr["lastTime"] = float64(time.Now().UnixNano())
		for _, k := range keys {
			v := lr[k]
			if vf, ok := v.(float64); ok {
				if vo, ok := pe.Result[k]; ok {
					if vof, ok := vo.(float64); ok {
						d := vf - vof
						lr[k+"_Delta"] = d
						vn := getValueName(k) + "_Delta"
						vm.Set(vn, d)
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
				for _, k := range keys {
					v := lr[k]
					if _, ok := v.(float64); ok {
						if vd, ok := lr[k+"_Delta"]; ok {
							if vdf, ok := vd.(float64); ok {
								lr[k+"_PS"] = float64((vdf * 100.0) / diff)
								vn := getValueName(k) + "_PS"
								vm.Set(vn, float64((vdf*100.0)/diff))
							}
						}
					}
				}
			}
		}
	}
	pe.Result = lr
	value, err := vm.Run(script)
	if err == nil {
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	}
	log.Printf("snmp polling err=%v", err)
	setPollingError("snmp", pe, err)
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
			regexFilter = nil
		}
	}
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		s := ""
		switch variable.Type {
		case gosnmp.OctetString:
			s = getMIBStringVal(variable.Value)
			mi := datastore.FindMIBInfo(variable.Name)
			if mi != nil {
				switch mi.Type {
				case "PhysAddress", "OctetString":
					a, ok := variable.Value.([]uint8)
					if !ok {
						a = []uint8(getMIBStringVal(variable.Value))
					}
					mac := []string{}
					for _, m := range a {
						mac = append(mac, fmt.Sprintf("%02X", m&0x00ff))
					}
					s = strings.Join(mac, ":")
				case "DisplayString":
				default:
				}
			}
		case gosnmp.ObjectIdentifier:
			s = datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
		default:
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
		s := getMIBStringVal(variable.Value)
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

func autoAddSnmpPolling(n *datastore.NodeEnt, pt *datastore.PollingTemplateEnt) {
	indexMIB := ""
	if strings.HasPrefix(pt.AutoMode, "index:") {
		a := strings.SplitAfterN(pt.AutoMode, ":", 2)
		if len(a) != 2 {
			return
		}
		indexMIB = a[1]
	} else {
		return
	}
	indexes := getSnmpIndex(n, indexMIB)
	for _, index := range indexes {
		p := new(datastore.PollingEnt)
		p.Name = pt.Name + " : " + index
		if hasSameNamePolling(n.ID, p.Name) {
			continue
		}
		p.NodeID = n.ID
		p.Type = pt.Type
		p.Params = strings.ReplaceAll(pt.Params, "$i", index)
		p.Mode = pt.Mode
		p.Script = pt.Script
		p.Extractor = pt.Extractor
		p.Filter = pt.Filter
		p.Level = pt.Level
		p.PollInt = datastore.MapConf.PollInt
		p.Timeout = datastore.MapConf.Timeout
		p.Retry = datastore.MapConf.Timeout
		p.LogMode = 0
		p.NextTime = 0
		p.State = "unknown"
		if err := datastore.AddPolling(p); err != nil {
			return
		}
	}
}

func getSnmpIndex(n *datastore.NodeEnt, name string) []string {
	ret := []string{}
	agent := &gosnmp.GoSNMP{
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(datastore.MapConf.Timeout) * time.Second,
		Retries:            datastore.MapConf.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	switch n.SnmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        n.Password,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        n.Password,
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("polling snmp err=%v", err)
		return ret
	}
	defer agent.Conn.Close()
	oid := datastore.MIBDB.NameToOID(name)
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		n := datastore.MIBDB.OIDToName(variable.Name)
		a := strings.SplitN(n, ".", 2)
		if len(a) == 2 {
			ret = append(ret, a[1])
		}
		return nil
	}); err != nil {
		return ret
	}
	return ret
}

func doPollingSnmpTraffic(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	oids := []string{}
	oids = append(oids, datastore.MIBDB.NameToOID("ifInOctets."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifInUcastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifInNUcastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifInDiscards."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifInErrors."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifInUnknownProtos."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifOutOctets."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifOutUcastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifOutNUcastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("sysUpTime.0"))
	result, err := agent.Get(oids)
	if err != nil {
		setPollingError("snmp", pe, err)
		return
	}
	lr := make(map[string]interface{})
	for _, variable := range result.Variables {
		if variable.Name == datastore.MIBDB.NameToOID("sysUpTime.0") {
			sut := gosnmp.ToBigInt(variable.Value).Uint64()
			lr["sysUpTime"] = float64(sut)
			continue
		}
		n := datastore.MIBDB.OIDToName(variable.Name)
		switch variable.Type {
		case gosnmp.OctetString:
			v := getMIBStringVal(variable.Value)
			lr[n] = v
		case gosnmp.ObjectIdentifier:
			v := datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
			lr[n] = v
		default:
			v := gosnmp.ToBigInt(variable.Value).Uint64()
			lr[n] = float64(v)
		}
	}
	// ifXTableからも取得すると
	oids = []string{}
	oids = append(oids, datastore.MIBDB.NameToOID("ifHCInOctets."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifHCInUcastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifHCInMulticastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifHCInBroadcastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifHCOutOctets."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifHCOutUcastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifHCOutMulticastPkts."+pe.Params))
	oids = append(oids, datastore.MIBDB.NameToOID("ifHCOutBroadcastPkts."+pe.Params))
	result, err = agent.Get(oids)
	if err == nil && len(result.Variables) > 0 {
		ifInNUcastPkts := float64(0.0)
		ifOutNUcastPkts := float64(0.0)
		for _, variable := range result.Variables {
			n := datastore.MIBDB.OIDToName(variable.Name)
			if strings.HasPrefix(n, "ifHCInOctets") {
				lr["ifInOctets."+pe.Params] = float64(gosnmp.ToBigInt(variable.Value).Uint64())
			} else if strings.HasPrefix(n, "ifHCInUcastPkts") {
				lr["ifInUcastPkts."+pe.Params] = float64(gosnmp.ToBigInt(variable.Value).Uint64())
			} else if strings.HasPrefix(n, "ifHCInMulticastPkts") {
				ifInNUcastPkts += float64(gosnmp.ToBigInt(variable.Value).Uint64())
			} else if strings.HasPrefix(n, "ifHCInBroadcastPkts") {
				ifInNUcastPkts += float64(gosnmp.ToBigInt(variable.Value).Uint64())
			} else if strings.HasPrefix(n, "ifHCOutOctets") {
				lr["ifOutOctets."+pe.Params] = float64(gosnmp.ToBigInt(variable.Value).Uint64())
			} else if strings.HasPrefix(n, "ifHCOutUcastPkts") {
				lr["ifOutUcastPkts."+pe.Params] = float64(gosnmp.ToBigInt(variable.Value).Uint64())
			} else if strings.HasPrefix(n, "ifHCOutMulticastPkts") {
				ifOutNUcastPkts += float64(gosnmp.ToBigInt(variable.Value).Uint64())
			} else if strings.HasPrefix(n, "ifHCOutBroadcastPkts") {
				ifOutNUcastPkts += float64(gosnmp.ToBigInt(variable.Value).Uint64())
			}
		}
		lr["ifOutNUcastPkts."+pe.Params] = ifInNUcastPkts
		lr["ifInNUcastPkts."+pe.Params] = ifOutNUcastPkts
	} else {
		log.Printf("ifXTable not found err=%v", err)
	}
	keys := []string{}
	for k := range lr {
		keys = append(keys, k)
	}
	if _, ok := pe.Result["lastTime"]; !ok {
		lr["lastTime"] = float64(time.Now().UnixNano())
		pe.Result = lr
		setPollingState(pe, "unknown")
		return
	}
	lr["lastTime"] = float64(time.Now().UnixNano())
	for _, k := range keys {
		v := lr[k]
		if vf, ok := v.(float64); ok {
			if vo, ok := pe.Result[k]; ok {
				if vof, ok := vo.(float64); ok {
					d := vf - vof
					lr[k+"_Delta"] = d
				}
			}
		}
	}
	var diff float64
	if v, ok := lr["sysUpTime_Delta"]; ok {
		if vf, ok := v.(float64); ok {
			diff = vf
		}
		if diff < 1.0 {
			setPollingError("snmp", pe, fmt.Errorf("no sysUptime"))
			return
		}
		for _, k := range keys {
			v := lr[k]
			if _, ok := v.(float64); ok {
				if vd, ok := lr[k+"_Delta"]; ok {
					if vdf, ok := vd.(float64); ok {
						lr[k+"_PS"] = float64((vdf * 100.0) / diff)
					}
				}
			}
		}
	}
	var bytes float64
	var packets float64
	var outBytes float64
	var outPackets float64
	var errors float64
	var bps float64
	var pps float64
	var obps float64
	var opps float64
	var eps float64
	for k, v := range lr {
		if strings.HasSuffix(k, "_Delta") {
			if vf, ok := v.(float64); ok {
				if strings.HasPrefix(k, "ifInOctets") {
					bytes += vf
				} else if strings.HasPrefix(k, "ifOutOctets") {
					outBytes += vf
				} else if strings.HasPrefix(k, "ifOut") {
					outPackets += vf
				} else if strings.HasPrefix(k, "ifIn") {
					packets += vf
					if strings.HasPrefix(k, "ifInErrors") {
						errors += vf
					}
				}
			}
			continue
		}
		if strings.HasSuffix(k, "_PS") {
			if vf, ok := v.(float64); ok {
				if strings.HasPrefix(k, "ifInOctets") {
					bps += vf
				} else if strings.HasPrefix(k, "ifOutOctets") {
					obps += vf
				} else if strings.HasPrefix(k, "ifOut") {
					opps += vf
				} else if strings.HasPrefix(k, "ifIn") {
					pps += vf
					if strings.HasPrefix(k, "ifInErrors") {
						eps += vf
					}
				}
			}
			continue
		}
	}
	pe.Result = lr
	pe.Result["bytes"] = bytes
	pe.Result["packets"] = packets
	pe.Result["outBytes"] = outBytes
	pe.Result["outPackets"] = outPackets
	pe.Result["erros"] = errors
	pe.Result["bps"] = bps
	pe.Result["pps"] = pps
	pe.Result["obps"] = obps
	pe.Result["opps"] = opps
	pe.Result["eps"] = eps
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm := otto.New()
	vm.Set("bps", bps)
	vm.Set("pps", pps)
	vm.Set("obps", obps)
	vm.Set("opps", opps)
	vm.Set("eps", eps)
	vm.Set("bytes", bytes)
	vm.Set("packets", packets)
	vm.Set("outBytes", outBytes)
	vm.Set("outPackets", outPackets)
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("snmp", pe, err)
	}
	if ok, _ := value.ToBoolean(); !ok {
		setPollingState(pe, pe.Level)
		return
	}
	setPollingState(pe, "normal")
}

func getMIBStringVal(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case []uint8:
		return string(v)
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return ""
}

func doPollingSnmpSystemDate(pe *datastore.PollingEnt, agent *gosnmp.GoSNMP) {
	script := pe.Script
	oids := []string{datastore.MIBDB.NameToOID("hrSystemDate.0")}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingError("snmp", pe, err)
		return
	}
	vm := otto.New()
	lr := make(map[string]interface{})
	for _, variable := range result.Variables {
		if variable.Name == datastore.MIBDB.NameToOID("hrSystemDate.0") {
			if v, ok := variable.Value.([]uint8); ok {
				ts := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d%c%02d:%02d",
					(int(v[0])*256 + int(v[1])), v[2], v[3], v[4], v[5], v[6], v[8], v[9], v[10])
				t, err := time.Parse(time.RFC3339, ts)
				if err != nil {
					setPollingError("snmp", pe, err)
					return
				}
				diff := t.Unix() - time.Now().Unix()
				if diff < 0 {
					diff *= -1
				}
				lr["hrSystemDate"] = ts
				lr["diff"] = float64(diff)
				vm.Set("diff", diff)
			}
			break
		}
	}
	pe.Result = lr
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
