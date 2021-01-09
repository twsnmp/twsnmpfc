package logger

/*
  syslogをログに記録する
*/

import (
	"encoding/json"
	"log"

	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func (l *Logger) syslogd(stopCh chan bool) {
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
		case log := <-syslogCh:
			{
				s, err := json.Marshal(log)
				if err == nil {
					l.logCh <- &datastore.LogEnt{
						Time: time.Now().UnixNano(),
						Type: "syslog",
						Log:  string(s),
					}
				}
			}
		}
	}
}
