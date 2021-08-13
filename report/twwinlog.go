package report

import (
	"fmt"
	"log"
	"strings"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	winLogSummaryCount = 0
	winLogOtherCount   = 0
)

func ReportTwWinLog(l map[string]interface{}) {
	twWinLogCh <- l
}

func checkTWWinLogReport(l map[string]interface{}) {
	h, ok := l["hostname"].(string)
	if !ok {
		log.Printf("twwinlog no hostname %v", l)
		return
	}
	msg, ok := l["content"].(string)
	if !ok {
		log.Printf("twwinlog no content %v", l)
		return
	}
	kvs := strings.Split(msg, ",")
	var m = make(map[string]string)
	for _, kv := range kvs {
		a := strings.SplitN(kv, "=", 2)
		if len(a) == 2 {
			m[a[0]] = a[1]
		}
	}
	t, ok := m["type"]
	if !ok {
		log.Printf("twwinlog no type %v", m)
		return
	}
	switch t {
	case "Stats":
		checkStats(h, "twwinlog", m)
	case "Monitor":
		checkMonitor(h, "twwinlog", m)
	case "Summary":
		checkWinEventID(h, m, l)
	default:
		log.Printf("twwinlog unkown type %v", m)
		winLogOtherCount++
	}
}

// type=Summary,computer=%s,channel=%s,provider=%s,eventID=%d,total=%d,count=%d,ft=%s,lt=%s

func checkWinEventID(h string, m map[string]string, l map[string]interface{}) {
	winLogSummaryCount++
	eventID := getNumberFromTWLog(m["eventID"])
	if eventID < 1 {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	level := "normal"
	if si, ok := l["severity"]; ok {
		if sn, ok := si.(int); ok {
			if sn < 4 {
				level = "error"
			} else if sn == 4 {
				level = "warn"
			}
		}
	}
	computer := m["computer"]
	channel := m["channel"]
	provider := m["provider"]
	lt := getTimeFromTWLog(m["lt"])
	id := fmt.Sprintf("%s:%s:%d", computer, provider, eventID)
	e := datastore.GetWinEventID(id)
	if e != nil {
		if e.LastTime < lt {
			e.LastTime = lt
		}
		if level == "error" || (e.Level != "error" && level == "warn") {
			e.Level = level
		}
		e.Count += count
		return
	}
	total := getNumberFromTWLog(m["total"])
	if total > count {
		count = total
	}
	datastore.AddWinEventID(&datastore.WinEventIDEnt{
		ID:        id,
		Level:     level,
		Provider:  provider,
		EventID:   int(eventID),
		Computer:  computer,
		Channel:   channel,
		Count:     count,
		FirstTime: getTimeFromTWLog(m["ft"]),
		LastTime:  lt,
	})
}

// type=Logon,target=%s,sid=%s,count=%d,logon=%d,failed=%d,logoff=%d,changeSubject=%d,changeLogonType=%d,changeIP=%d,subject=%s,subsid=%s,logonType=%s,ip=%s,failCode=%s,ft=%s,lt=%s
// type=Account,target=%s,sid=%s,computer=%s,count=%d,edit=%d,password=%d,other=%d,changesubject=%d,subject=%s,sbjectsid=%s,ft=%s,lt=%s
// type=KerberosTGT,target=%s,sid=%s,ip=%s,computer=%s,count=%d,failed=%d,changeStatus=%d,changeCert=%d,status=%s,cert=%s,ft=%s,lt=%s
// type=KerberosST,target=%s,servcie=%s,sid=%s,ip=%s,computer=%s,count=%d,failed=%d,changeStatus=%d,status=%s,ft=%s,lt=%s

// type=Privilege,subject=%s,sid=%s,computer=%s,count=%d,ft=%s,lt=%s
// type=Process,computer=%s,process=%s,count=%d,start=%d,exit=%d,changeSubject=%d,changeStatus=%d,changeParent=%d,subject=%s,status=%s,parent=%s,ft=%s,lt=%s
// type=Task,taskname=%s,computer=%s,subject=%s,sid=%s,count=%d,ft=%s,lt=%s

// 通知系
// type=ClearLog,subject=%s@%s,sid=%s
// type=LogonFailed,subject=%s@%s,target=%s@%s,targetsid=%s,logonType=%s,ip=%s,code=%s,time=%s
// type=KerberosTGTFailed,target=%s,sid=%s,ip=%s,status=%s,time=%s
// type=KerberosSTFailed,target=%s,servcie=%s,sid=%s,ip=%s,status=%s,time=%s",
