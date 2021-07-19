package logger

/*
  syslogをログに記録する
*/

import (
	"encoding/json"
	"log"

	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func syslogd(stopCh chan bool) {
	syslogCh := make(syslog.LogPartsChannel)
	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(syslog.NewChannelHandler(syslogCh))
	_ = server.ListenUDP("0.0.0.0:514")
	_ = server.ListenTCP("0.0.0.0:514")
	_ = server.Boot()
	log.Printf("syslogd start")
	for {
		select {
		case <-stopCh:
			{
				log.Printf("syslogd stop")
				_ = server.Kill()
				return
			}
		case sl := <-syslogCh:
			{
				s, err := json.Marshal(sl)
				if err == nil {
					if tag, ok := sl["tag"].(string); ok && tag == "twpcap" {
						report.ReportTWPCAP(sl)
					}
					logCh <- &datastore.LogEnt{
						Time: time.Now().UnixNano(),
						Type: "syslog",
						Log:  string(s),
					}
				}
			}
		}
	}
}
