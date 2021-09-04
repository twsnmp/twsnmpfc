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
	syslogCh := make(syslog.LogPartsChannel, 2000)
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
				syslogCount++
				s, err := json.Marshal(sl)
				if err == nil {
					tag, ok := sl["tag"].(string)
					if ok {
						switch tag {
						case "twpcap":
							report.ReportTWPCAP(sl)
						case "twwinlog":
							report.ReportTwWinLog(sl)
						case "twBlueScan":
							report.ReportTWBuleScan(sl)
						case "twWifiScan":
							report.ReportTWWifiScan(sl)
						}
					}
					logCh <- &datastore.LogEnt{
						Time: time.Now().UnixNano(),
						Type: "syslog",
						Log:  string(s),
					}
				} else {
					log.Printf("syslog Marshal err=%v", err)
				}
			}
		}
	}
}
