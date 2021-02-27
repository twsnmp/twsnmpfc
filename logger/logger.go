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

var logCh chan *datastore.LogEnt

func StartLogger(ctx context.Context) error {
	log.Println("Start Logger")
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
				log.Printf("Stop logger")
				return
			}
		case log := <-logCh:
			{
				logBuffer = append(logBuffer, log)
			}
		case <-timer.C:
			{
				if len(logBuffer) > 0 {
					datastore.SaveLogBuffer(logBuffer)
					logBuffer = []*datastore.LogEnt{}
				}
				if datastore.MapConf.EnableSyslogd && !syslogdRunning {
					stopSyslogd = make(chan bool)
					syslogdRunning = true
					go syslogd(stopSyslogd)
					log.Printf("start syslogd")
				} else if !datastore.MapConf.EnableSyslogd && syslogdRunning {
					close(stopSyslogd)
					syslogdRunning = false
					log.Printf("stop syslogd")
				}
				if datastore.MapConf.EnableTrapd && !trapdRunning {
					stopTrapd = make(chan bool)
					trapdRunning = true
					go snmptrapd(stopTrapd)
					log.Printf("start trapd")
				} else if !datastore.MapConf.EnableTrapd && trapdRunning {
					close(stopTrapd)
					trapdRunning = false
					log.Printf("stop trapd")
				}
				if datastore.MapConf.EnableNetflowd && !netflowdRunning {
					stopNetflowd = make(chan bool)
					netflowdRunning = true
					go netflowd(stopNetflowd)
					log.Printf("start netflowd")
				} else if !datastore.MapConf.EnableNetflowd && netflowdRunning {
					close(stopNetflowd)
					netflowdRunning = false
					log.Printf("stop netflowd")
				}
				if datastore.MapConf.EnableArpWatch && !arpWatchRunning {
					stopArpWatch = make(chan bool)
					arpWatchRunning = true
					go arpWatch(stopArpWatch)
					log.Printf("start arpWatch")
				} else if !datastore.MapConf.EnableArpWatch && arpWatchRunning {
					close(stopArpWatch)
					arpWatchRunning = false
					log.Printf("stop arpWatch")
				}
			}
		}
	}
}
