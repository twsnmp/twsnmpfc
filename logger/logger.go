// Package logger : ログ受信処理
package logger

/*
  syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"context"
	"fmt"
	"log"
	"sync"

	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var logCh = make(chan *datastore.LogEnt, 5000)

var trapListen = "0.0.0.0:162"
var netflowListen = ":2055"
var syslogListen = "0.0.0.0:514"
var sflowListen = ":6343"

func Start(ctx context.Context, wg *sync.WaitGroup, trap, netflow, syslog, sflow int) error {
	trapListen = fmt.Sprintf("0.0.0.0:%d", trap)
	netflowListen = fmt.Sprintf(":%d", netflow)
	syslogListen = fmt.Sprintf("0.0.0.0:%d", syslog)
	sflowListen = fmt.Sprintf(":%d", sflow)
	logCh = make(chan *datastore.LogEnt, 100)
	wg.Add(1)
	go logger(ctx, wg)
	return nil
}

func logger(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	var syslogdRunning = false
	var trapdRunning = false
	var netflowdRunning = false
	var arpWatchRunning = false
	var sshdRunning = false
	var sflowdRunning = false
	var stopSyslogd chan bool
	var stopTrapd chan bool
	var stopNetflowd chan bool
	var stopArpWatch chan bool
	var stopSshd chan bool
	var stopSflowd chan bool
	log.Println("start logger")
	timer1 := time.NewTicker(time.Second * 10)
	timer2 := time.NewTicker(time.Second * 1)
	logBuffer := []*datastore.LogEnt{}
	for {
		select {
		case <-ctx.Done():
			{
				timer1.Stop()
				timer2.Stop()
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
				if sshdRunning {
					close(stopSshd)
				}
				if sflowdRunning {
					close(stopSflowd)
				}
				if len(logBuffer) > 0 {
					datastore.SaveLogBuffer(logBuffer)
				}
				log.Printf("stop logger")
				return
			}
		case l := <-logCh:
			logBuffer = append(logBuffer, l)
		case <-timer1.C:
			if len(logBuffer) > 0 {
				datastore.SaveLogBuffer(logBuffer)
				logBuffer = []*datastore.LogEnt{}
			}
		case <-timer2.C:
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
			if datastore.MapConf.EnableSshd && !sshdRunning {
				stopSshd = make(chan bool)
				sshdRunning = true
				go sshd(stopSshd)
			} else if !datastore.MapConf.EnableSshd && sshdRunning {
				close(stopSshd)
				sshdRunning = false
			}
			if datastore.MapConf.EnableSflowd && !sflowdRunning {
				stopSflowd = make(chan bool)
				sflowdRunning = true
				go sflowd(stopSflowd)
			} else if !datastore.MapConf.EnableSflowd && sflowdRunning {
				close(stopSflowd)
				sflowdRunning = false
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
