package logger

/*
  syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"encoding/json"
	"log"

	"fmt"
	"net"
	"time"

	gosnmp "github.com/twsnmp/gosnmp"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func snmptrapd(stopCh chan bool) {
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
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.NoAuthNoPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
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
				val = datastore.MIBDB.OIDToName(vb.Value.(string))
			case gosnmp.OctetString:
				val = vb.Value.(string)
			default:
				val = fmt.Sprintf("%d", gosnmp.ToBigInt(vb.Value).Uint64())
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
		log.Printf("close snmp trap listen")
	}()
	<-stopCh
	log.Printf("stop snmp rrap listenner")
}
