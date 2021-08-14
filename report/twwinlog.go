package report

import (
	"fmt"
	"log"
	"strings"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	winEventIDCount     = 0
	winLogonCount       = 0
	winAccountCount     = 0
	winKerberosTGTCount = 0
	winKerberosSTCount  = 0
	winPrivilegeCount   = 0
	winProcessCount     = 0
	winTaskCount        = 0
	winOtherCount       = 0
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
	case "Logon":
		checkWinLogon(h, m)
	case "Account":
		checkWinAccount(h, m)
	case "KerberosTGT":
		checkWinKerberosTGT(h, m)
	case "KerberosST":
		checkWinKerberosST(h, m)
	case "Privilege":
		checkWinPrivilege(h, m)
	case "Process":
		checkWinProcess(h, m)
	case "Task":
		checkWinTask(h, m)
	default:
		log.Printf("twwinlog unkown type %v", m)
		winOtherCount++
	}
}

// type=Summary,computer=%s,channel=%s,provider=%s,eventID=%d,total=%d,count=%d,ft=%s,lt=%s
func checkWinEventID(h string, m map[string]string, l map[string]interface{}) {
	winEventIDCount++
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
func checkWinLogon(h string, m map[string]string) {
	winLogonCount++
	id, ok := m["target"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	logon := getNumberFromTWLog(m["logon"])
	failed := getNumberFromTWLog(m["failed"])
	logoff := getNumberFromTWLog(m["logoff"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinLogon(id)
	if e != nil {
		e.LastTime = lt
		e.Count = count
		e.Logon = logon
		e.Logoff = logoff
		e.Failed = failed
		e.LastIP = m["ip"]
		return
	}
	datastore.AddWinLogon(&datastore.WinLogonEnt{
		ID:          id,
		Count:       count,
		Logon:       logon,
		Logoff:      logoff,
		Failed:      failed,
		LastIP:      m["ip"],
		LastSubject: m["subject"],
		FirstTime:   getTimeFromTWLog(m["ft"]),
		LastTime:    lt,
	})
}

// type=Account,target=%s,sid=%s,computer=%s,count=%d,edit=%d,password=%d,other=%d,changesubject=%d,subject=%s,sbjectsid=%s,ft=%s,lt=%s
func checkWinAccount(h string, m map[string]string) {
	winAccountCount++
	target, ok := m["target"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	edit := getNumberFromTWLog(m["edit"])
	password := getNumberFromTWLog(m["password"])
	other := getNumberFromTWLog(m["other"])
	lt := getTimeFromTWLog(m["lt"])
	id := fmt.Sprintf("%s:%s", target, m["computer"])
	e := datastore.GetWinAccount(id)
	if e != nil {
		e.LastTime = lt
		e.Count = count
		e.Edit = edit
		e.Password = password
		e.Other = other
		e.LastSubject = m["subject"]
		return
	}
	datastore.AddWinAccount(&datastore.WinAccountEnt{
		ID:          id,
		Count:       count,
		Edit:        edit,
		Password:    password,
		Other:       other,
		LastSubject: m["subject"],
		FirstTime:   getTimeFromTWLog(m["ft"]),
		LastTime:    lt,
	})
}

// type=KerberosTGT,target=%s,sid=%s,ip=%s,computer=%s,count=%d,failed=%d,changeStatus=%d,changeCert=%d,status=%s,cert=%s,ft=%s,lt=%s
func checkWinKerberosTGT(h string, m map[string]string) {
	winKerberosTGTCount++
	target, ok := m["target"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := fmt.Sprintf("%s<%s", target, m["ip"])
	failed := getNumberFromTWLog(m["failed"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinKerberosTGT(id)
	if e != nil {
		e.LastTime = lt
		e.Count = count
		e.Failed = failed
		e.IP = m["ip"]
		return
	}
	datastore.AddWinKerberosTGT(&datastore.WinKerberosTGTEnt{
		ID:        id,
		Target:    target,
		Count:     count,
		Failed:    failed,
		IP:        m["ip"],
		FirstTime: getTimeFromTWLog(m["ft"]),
		LastTime:  lt,
	})
}

// type=KerberosST,target=%s,servcie=%s,sid=%s,ip=%s,computer=%s,count=%d,failed=%d,changeStatus=%d,status=%s,ft=%s,lt=%s
func checkWinKerberosST(h string, m map[string]string) {
	winKerberosSTCount++
	target, ok := m["target"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := fmt.Sprintf("%s<%s", target, m["ip"])
	failed := getNumberFromTWLog(m["failed"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinKerberosST(id)
	if e != nil {
		e.LastTime = lt
		e.Count = count
		e.Failed = failed
		e.IP = m["ip"]
		return
	}
	datastore.AddWinKerberosST(&datastore.WinKerberosSTEnt{
		ID:        id,
		Target:    target,
		Count:     count,
		Failed:    failed,
		IP:        m["ip"],
		FirstTime: getTimeFromTWLog(m["ft"]),
		LastTime:  lt,
	})
}

// type=Privilege,subject=%s,sid=%s,computer=%s,count=%d,ft=%s,lt=%s
func checkWinPrivilege(h string, m map[string]string) {
	winPrivilegeCount++
	subject, ok := m["subject"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := fmt.Sprintf("%s:%s", subject, m["computer"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinPrivilege(id)
	if e != nil {
		e.LastTime = lt
		e.Count = count
		return
	}
	datastore.AddWinPrivilege(&datastore.WinPrivilegeEnt{
		ID:        id,
		Subject:   subject,
		Count:     count,
		Computer:  m["computer"],
		FirstTime: getTimeFromTWLog(m["ft"]),
		LastTime:  lt,
	})
}

// type=Process,computer=%s,process=%s,count=%d,start=%d,exit=%d,changeSubject=%d,changeStatus=%d,changeParent=%d,subject=%s,status=%s,parent=%s,ft=%s,lt=%s
func checkWinProcess(h string, m map[string]string) {
	winProcessCount++
	process, ok := m["process"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := fmt.Sprintf("%s:%s", process, m["computer"])
	start := getNumberFromTWLog(m["start"])
	exit := getNumberFromTWLog(m["exit"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinProcess(id)
	if e != nil {
		e.LastTime = lt
		e.Count = count
		e.Start = start
		e.Exit = exit
		return
	}
	datastore.AddWinProcess(&datastore.WinProcessEnt{
		ID:        id,
		Process:   process,
		Computer:  m["computer"],
		Count:     count,
		Start:     start,
		Exit:      exit,
		FirstTime: getTimeFromTWLog(m["ft"]),
		LastTime:  lt,
	})
}

// type=Task,taskname=%s,computer=%s,subject=%s,sid=%s,count=%d,ft=%s,lt=%s
func checkWinTask(h string, m map[string]string) {
	winTaskCount++
	taskname, ok := m["taskname"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := fmt.Sprintf("%s:%s", taskname, m["computer"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinTask(id)
	if e != nil {
		e.LastTime = lt
		e.Count = count
		return
	}
	datastore.AddWinTask(&datastore.WinTaskEnt{
		ID:        id,
		TaskName:  taskname,
		Computer:  m["computer"],
		Subject:   m["subject"],
		Count:     count,
		FirstTime: getTimeFromTWLog(m["ft"]),
		LastTime:  lt,
	})
}

// 通知系
// type=ClearLog,subject=%s@%s,sid=%s
// type=LogonFailed,subject=%s@%s,target=%s@%s,targetsid=%s,logonType=%s,ip=%s,code=%s,time=%s
// type=KerberosTGTFailed,target=%s,sid=%s,ip=%s,status=%s,time=%s
// type=KerberosSTFailed,target=%s,servcie=%s,sid=%s,ip=%s,status=%s,time=%s",
