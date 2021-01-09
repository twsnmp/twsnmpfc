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

func (l *Logger) snmptrapd(stopCh chan bool) {
	tl := gosnmp.NewTrapListener()
	if l.ds.MapConf.SnmpMode != "" {
		tl.Params = &gosnmp.GoSNMP{}
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		if l.ds.MapConf.SnmpMode == "v3auth" {
			tl.Params.MsgFlags = gosnmp.AuthNoPriv
			tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 l.ds.MapConf.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: l.ds.MapConf.Password,
			}
		} else {
			tl.Params.MsgFlags = gosnmp.AuthPriv
			tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 l.ds.MapConf.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: l.ds.MapConf.Password,
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        l.ds.MapConf.Password,
			}
		}
	}
	tl.OnNewTrap = func(s *gosnmp.SnmpPacket, u *net.UDPAddr) {
		var record = make(map[string]interface{})
		record["FromAddress"] = u.String()
		record["Timestamp"] = s.Timestamp
		record["Enterprise"] = l.ds.MIBDB.OIDToName(s.Enterprise)
		record["GenericTrap"] = s.GenericTrap
		record["SpecificTrap"] = s.SpecificTrap
		record["Variables"] = ""
		vbs := ""
		for _, vb := range s.Variables {
			key := l.ds.MIBDB.OIDToName(vb.Name)
			val := ""
			switch vb.Type {
			case gosnmp.ObjectIdentifier:
				val = l.ds.MIBDB.OIDToName(vb.Value.(string))
			case gosnmp.OctetString:
				val = vb.Value.(string)
			default:
				val = fmt.Sprintf("%d", gosnmp.ToBigInt(vb.Value).Uint64())
			}
			vbs += fmt.Sprintf("%s=%s\n", key, val)
		}
		record["Variables"] = vbs
		js, err := json.Marshal(record)
		if err != nil {
			log.Println(err)
		}
		l.logCh <- &datastore.LogEnt{
			Time: time.Now().UnixNano(),
			Type: "trap",
			Log:  string(js),
		}
	}
	defer tl.Close()
	go func() {
		if err := tl.Listen("0.0.0.0:162"); err != nil {
			log.Printf("Trap Listen err=%v", err)
		}
		log.Printf("Trap Listen End")
	}()
	<-stopCh
	log.Printf("Trap Listen Done")
}
