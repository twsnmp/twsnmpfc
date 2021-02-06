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
	"github.com/twsnmp/twsnmpfc/report"
)

type Logger struct {
	ds     *datastore.DataStore
	report *report.Report
	logCh  chan *datastore.LogEnt
}

func NewLogger(ctx context.Context, ds *datastore.DataStore, r *report.Report) *Logger {
	log.Println("Start Logger")
	l := &Logger{
		ds:     ds,
		report: r,
		logCh:  make(chan *datastore.LogEnt, 100),
	}
	go l.logger(ctx)
	return l
}

func (l *Logger) logger(ctx context.Context) {
	var syslogdRunning = false
	var trapdRunning = false
	var netflowdRunning = false
	var stopSyslogd chan bool
	var stopTrapd chan bool
	var stopNetflowd chan bool
	timer := time.NewTicker(time.Second * 10)
	logBuffer := []*datastore.LogEnt{}
	for {
		select {
		case <-ctx.Done():
			{
				timer.Stop()
				if len(logBuffer) > 0 {
					l.ds.SaveLogBuffer(logBuffer)
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
				log.Printf("Stop logger")
				return
			}
		case log := <-l.logCh:
			{
				logBuffer = append(logBuffer, log)
			}
		case <-timer.C:
			{
				if len(logBuffer) > 0 {
					l.ds.SaveLogBuffer(logBuffer)
					logBuffer = []*datastore.LogEnt{}
				}
				if l.ds.MapConf.EnableSyslogd && !syslogdRunning {
					stopSyslogd = make(chan bool)
					syslogdRunning = true
					go l.syslogd(stopSyslogd)
					log.Printf("start syslogd")
				} else if !l.ds.MapConf.EnableSyslogd && syslogdRunning {
					close(stopSyslogd)
					syslogdRunning = false
					log.Printf("stop syslogd")
				}
				if l.ds.MapConf.EnableTrapd && !trapdRunning {
					stopTrapd = make(chan bool)
					trapdRunning = true
					go l.snmptrapd(stopTrapd)
					log.Printf("start trapd")
				} else if !l.ds.MapConf.EnableTrapd && trapdRunning {
					close(stopTrapd)
					trapdRunning = false
					log.Printf("stop trapd")
				}
				if l.ds.MapConf.EnableNetflowd && !netflowdRunning {
					stopNetflowd = make(chan bool)
					netflowdRunning = true
					go l.netflowd(stopNetflowd)
					log.Printf("start netflowd")
				} else if !l.ds.MapConf.EnableNetflowd && netflowdRunning {
					close(stopNetflowd)
					netflowdRunning = false
					log.Printf("stop netflowd")
				}
			}
		}
	}
}
