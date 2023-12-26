package logger

/*
  syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"encoding/json"
	"log"
	"strings"

	"fmt"
	"net"
	"time"

	gosnmp "github.com/gosnmp/gosnmp"

	"github.com/twsnmp/twsnmpfc/datastore"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func snmptrapd(stopCh chan bool) {
	log.Printf("start snmp trapd")
	tl := gosnmp.NewTrapListener()
	tl.Params = &gosnmp.GoSNMP{}
	switch datastore.MapConf.SnmpMode {
	case "v3auth":
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthNoPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
		}
	case "v3authpriv":
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	case "v3authprivex":
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	default:
		// SNMPv2c
		tl.Params.Version = gosnmp.Version2c
		tl.Params.Community = datastore.MapConf.Community
	}
	tl.OnNewTrap = func(s *gosnmp.SnmpPacket, u *net.UDPAddr) {
		var record = make(map[string]interface{})
		record["FromAddress"] = u.String()
		record["Timestamp"] = s.Timestamp
		record["Enterprise"] = datastore.MIBDB.OIDToName(s.Enterprise)
		record["GenericTrap"] = s.GenericTrap
		record["SpecificTrap"] = s.SpecificTrap
		record["Variables"] = ""
		vbs := ""
		for _, vb := range s.Variables {
			key := datastore.MIBDB.OIDToName(vb.Name)
			val := ""
			switch vb.Type {
			case gosnmp.ObjectIdentifier:
				val = datastore.MIBDB.OIDToName(datastore.PrintMIBStringVal(vb.Value))
			case gosnmp.OctetString:
				mi := datastore.FindMIBInfo(key)
				if mi != nil {
					switch mi.Type {
					case "PhysAddress", "OctetString":
						a, ok := vb.Value.([]uint8)
						if !ok {
							a = []uint8(datastore.PrintMIBStringVal(vb.Value))
						}
						mac := []string{}
						for _, m := range a {
							mac = append(mac, fmt.Sprintf("%02X", m&0x00ff))
						}
						val = strings.Join(mac, ":")
					case "BITS":
						a, ok := vb.Value.([]uint8)
						if !ok {
							a = []uint8(datastore.PrintMIBStringVal(vb.Value))
						}
						hex := []string{}
						ap := []string{}
						bit := 0
						for _, m := range a {
							hex = append(hex, fmt.Sprintf("%02X", m&0x00ff))
							if mi.Enum != "" {
								for i := 0; i < 8; i++ {
									if (m & 0x80) == 0x80 {
										if n, ok := mi.EnumMap[bit]; ok {
											ap = append(ap, fmt.Sprintf("%s(%d)", n, bit))
										}
									}
									m <<= 1
									bit++
								}
							}
						}
						val = strings.Join(hex, " ")
						if len(ap) > 0 {
							val += " " + strings.Join(ap, " ")
						}
					case "DisplayString":
						val = datastore.PrintMIBStringVal(vb.Value)
						if datastore.MapConf.AutoCharCode {
							val = CheckCharCode(val)
						}
					case "DateAndTime":
						val = datastore.PrintDateAndTime(vb.Value)
					default:
						val = datastore.PrintMIBStringVal(vb.Value)
					}
				} else {
					val = datastore.PrintMIBStringVal(vb.Value)
					if datastore.MapConf.AutoCharCode {
						val = CheckCharCode(val)
					}
				}
			case gosnmp.TimeTicks:
				val = getTimeTickStr(gosnmp.ToBigInt(vb.Value).Int64())
			case gosnmp.IPAddress:
				val = datastore.PrintIPAddress(vb.Value)
			default:
				v := int(gosnmp.ToBigInt(vb.Value).Uint64())
				val = fmt.Sprintf("%d", v)
				mi := datastore.FindMIBInfo(key)
				if mi != nil {
					if mi.Enum != "" {
						if vn, ok := mi.EnumMap[v]; ok {
							val += "(" + vn + ")"
						}
					} else {
						if mi.Hint != "" {
							val = datastore.PrintHintedMIBIntVal(int32(v), mi.Hint, vb.Type != gosnmp.Integer)
						}
						if mi.Units != "" {
							val += " " + mi.Units
						}
					}
				}
			}
			vbs += fmt.Sprintf("%s=%s\n", key, val)
		}
		record["Variables"] = vbs
		js, err := json.Marshal(record)
		if err == nil {
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "trap",
				Log:  string(js),
			}
		}
	}
	defer tl.Close()
	go func() {
		if err := tl.Listen("0.0.0.0:162"); err != nil {
			log.Printf("snmp trap listen err=%v", err)
		}
		log.Printf("close snmp trapd")
	}()
	<-stopCh
	log.Printf("stop snmp trapd")
}

func getTimeTickStr(t int64) string {
	ft := float64(t) / 100
	if ft > 3600*24 {
		return fmt.Sprintf("%.2f日(%d)", ft/(3600*24), t)
	} else if ft > 3600 {
		return fmt.Sprintf("%.2f時間(%d)", ft/(3600), t)
	}
	return fmt.Sprintf("%.2f秒(%d)", ft, t)
}

func CheckCharCode(s string) string {
	if isSjis([]byte(s)) {
		dec := japanese.ShiftJIS.NewDecoder()
		if b, _, err := transform.Bytes(dec, []byte(s)); err == nil {
			return string(b)
		}
	}
	return s
}

func isSjis(p []byte) bool {
	f := false
	for _, c := range p {
		if f {
			if c < 0x0040 || c > 0x00fc {
				return false
			}
			f = false
			continue
		}
		if c < 0x007f {
			continue
		}
		if (c >= 0x0081 && c <= 0x9f) ||
			(c >= 0x00e0 && c <= 0x00ef) {
			f = true
		} else {
			return false
		}
	}
	return true
}
