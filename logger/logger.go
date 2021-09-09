// Package logger : ログ受信処理
package logger

/*
  syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"context"
	"log"

	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var logCh = make(chan *datastore.LogEnt, 5000)
var syslogCount = 0
var netflowCount = 0

func Start(ctx context.Context) error {
	logCh = make(chan *datastore.LogEnt, 100)
	go logger(ctx)
	return nil
}

func logger(ctx context.Context) {
	var syslogdRunning = false
	var trapdRunning = false
	var netflowdRunning = false
	var arpWatchRunning = false
	var stopSyslogd chan bool
	var stopTrapd chan bool
	var stopNetflowd chan bool
	var stopArpWatch chan bool
	log.Println("start logger")
	timer := time.NewTicker(time.Second * 10)
	logBuffer := []*datastore.LogEnt{}
	for {
		select {
		case <-ctx.Done():
			{
				timer.Stop()
				if len(logBuffer) > 0 {
					datastore.SaveLogBuffer(logBuffer)
				}
				if syslogdRunning {
					close(stopSyslogd)
				}
				if netflowdRunning {
					close(stopNetflowd)
				}
				if trapdRunning {
					close(stopTrapd)
				}
				if arpWatchRunning {
					close(stopArpWatch)
				}
				log.Printf("stop logger")
				return
			}
		case l := <-logCh:
			logBuffer = append(logBuffer, l)
		case <-timer.C:
			{
				if len(logBuffer) > 0 {
					st := time.Now()
					datastore.SaveLogBuffer(logBuffer)
					log.Printf("save log len=%d dur=%v syslog=%d netflow=%d",
						len(logBuffer), time.Since(st), syslogCount, netflowCount)
					logBuffer = []*datastore.LogEnt{}
				}
				if datastore.MapConf.EnableSyslogd && !syslogdRunning {
					stopSyslogd = make(chan bool)
					syslogdRunning = true
					go syslogd(stopSyslogd)
				} else if !datastore.MapConf.EnableSyslogd && syslogdRunning {
					close(stopSyslogd)
					syslogdRunning = false
				}
				if datastore.MapConf.EnableTrapd && !trapdRunning {
					stopTrapd = make(chan bool)
					trapdRunning = true
					go snmptrapd(stopTrapd)
				} else if !datastore.MapConf.EnableTrapd && trapdRunning {
					close(stopTrapd)
					trapdRunning = false
				}
				if datastore.MapConf.EnableNetflowd && !netflowdRunning {
					stopNetflowd = make(chan bool)
					netflowdRunning = true
					go netflowd(stopNetflowd)
				} else if !datastore.MapConf.EnableNetflowd && netflowdRunning {
					close(stopNetflowd)
					netflowdRunning = false
				}
				if datastore.MapConf.EnableArpWatch && !arpWatchRunning {
					stopArpWatch = make(chan bool)
					arpWatchRunning = true
					go arpWatch(stopArpWatch)
				} else if !datastore.MapConf.EnableArpWatch && arpWatchRunning {
					close(stopArpWatch)
					arpWatchRunning = false
				}
				if datastore.RestartSnmpTrapd && trapdRunning {
					close(stopTrapd)
					datastore.RestartSnmpTrapd = false
					trapdRunning = false
					log.Printf("resatrt trapd")
				}
			}
		}
	}
}
