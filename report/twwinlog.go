package report

import (
	"fmt"
	"log"
	"strings"

	"github.com/twsnmp/twsnmpfc/datastore"
)

var (
	winEventIDCount   = 0
	winLogonCount     = 0
	winAccountCount   = 0
	winKerberosCount  = 0
	winPrivilegeCount = 0
	winProcessCount   = 0
	winTaskCount      = 0
	winOtherCount     = 0
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
	case "EventID":
		checkWinEventID(h, m, l)
	case "Logon", "Logoff", "LogonFailed":
		checkWinLogon(m)
	case "Account":
		checkWinAccount(h, m)
	case "Kerberos":
		checkWinKerberos(h, m)
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

// type=EventID,computer=%s,channel=%s,provider=%s,eventID=%d,total=%d,count=%d,ft=%s,lt=%s
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

// type=Logon,subject=@,target=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,ip=192.168.1.250,logonType=Network,time=2021-08-19T05:17:43+09:00
// type=LogonFailed,subject=@,target=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,ip=192.168.1.9,logonType=Network,failedCode=BadPassword,time=2021-08-19T04:46:28+09:00
// type=Logoff,subject=@,target=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,ip=,logonType=Network,time=2021-08-19T04:46:10+09:00

func checkWinLogon(m map[string]string) {
	winLogonCount++
	tp, ok := m["type"]
	if !ok {
		return
	}
	target, ok := m["target"]
	if !ok {
		return
	}
	computer := m["computer"]
	logonType := m["logonType"]
	failedCode := m["failedCode"]
	ts := getTimeFromTWLog(m["time"])
	if tp == "Logoff" {
		datastore.ForEachWinLogon(func(e *datastore.WinLogonEnt) bool {
			if e.Computer == computer && e.Target == target && e.LogonType[logonType] > 0 {
				e.Logoff++
				if e.LastTime < ts {
					e.LastTime = ts
				}
				return false
			}
			return true
		})
		return
	}

	id := fmt.Sprintf("%s:%s:%s", target, m["computer"], m["ip"])
	e := datastore.GetWinLogon(id)
	if e != nil {
		if e.LastTime < ts {
			e.LastTime = ts
		}
		e.Count++
		if tp == "LogonFailed" {
			e.Failed++
			if failedCode != "" {
				e.FailedCode[failedCode]++
			}
		} else {
			e.Logon++
			if logonType != "" {
				e.LogonType[logonType]++
			}
		}
		return
	}
	e = &datastore.WinLogonEnt{
		ID:         id,
		Target:     target,
		Computer:   computer,
		IP:         m["ip"],
		Count:      1,
		FirstTime:  ts,
		LastTime:   ts,
		LogonType:  make(map[string]int),
		FailedCode: make(map[string]int),
	}
	if tp == "LogonFailed" {
		e.Failed++
		if failedCode != "" {
			e.FailedCode[failedCode] = 1
		}
	} else {
		if logonType != "" {
			e.LogonType[logonType] = 1
		}
		e.Logon++
	}
	datastore.AddWinLogon(e)
}

/*
2021-08-19T04:47:58.322 logon id=MYAMAI@DESKTOP-T6L1D1U,e=type=Logon,target=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,ip=,count=7,logon=3,failed=1,logoff=3,logonType[Network]=3,logonType[Unlock]=4,failedCode[BadPassword]=1,ft=2021-08-19T04:46:10+09:00,lt=2021-08-19T04:46:33+09:00
2021-08-19T04:47:58.323 logon id=MYAMAI@LOCALHOST,e=type=Logon,target=myamai@localhost,computer=DESKTOP-T6L1D1U,ip=192.168.1.9,count=1,logon=1,failed=0,logoff=0,logonType[Explicit]=1,ft=2021-08-19T04:46:33+09:00,lt=2021-08-19T04:46:33+09:00
2021-08-19T04:48:12.053 privilege id=myamai@DESKTOP-T6L1D1U,e=type=Privilege,subject=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,count=6,ft=2021-08-19T04:48:02+09:00,lt=2021-08-19T04:48:03+09:00
2021-08-19T04:47:58.323 privilege id=myamai@DESKTOP-T6L1D1U,e=type=Privilege,subject=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,count=13,ft=2021-08-19T04:46:11+09:00,lt=2021-08-19T04:46:34+09:00
2021-08-19T04:47:53.337 privilege id=myamai@DESKTOP-T6L1D1U,e=type=Privilege,subject=myamai@DESKTOP-T6L1D1U,computer=DESKTOP-T6L1D1U,count=462,ft=2021-08-18T16:51:00+09:00,lt=2021-08-19T04:41:17+09:00
2021-08-19T04:47:58.323 privilege id=DESKTOP-T6L1D1U$@WORKGROUP,e=type=Privilege,subject=DESKTOP-T6L1D1U$@WORKGROUP,computer=DESKTOP-T6L1D1U,count=11,ft=2021-08-19T04:46:11+09:00,lt=2021-08-19T04:46:33+09:00
2021-08-19T04:47:53.339 privilege id=DESKTOP-T6L1D1U$@WORKGROUP,e=type=Privilege,subject=DESKTOP-T6L1D1U$@WORKGROUP,computer=DESKTOP-T6L1D1U,count=295,ft=2021-08-18T16:55:52+09:00,lt=2021-08-19T04:40:31+09:00
*/

// type=Account,subject=%s,target=%s,computer=%s,count=%d,edit=%d,password=%d,other=%d,ft=%s,lt=%s",
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
	id := fmt.Sprintf("%s:%s:%s", target, m["computer"], m["subject"])
	e := datastore.GetWinAccount(id)
	if e != nil {
		e.LastTime = lt
		e.Count += count
		e.Edit += edit
		e.Password += password
		e.Other += other
		return
	}
	datastore.AddWinAccount(&datastore.WinAccountEnt{
		ID:        id,
		Subject:   m["subject"],
		Target:    target,
		Computer:  m["computer"],
		Count:     count,
		Edit:      edit,
		Password:  password,
		Other:     other,
		FirstTime: getTimeFromTWLog(m["ft"]),
		LastTime:  lt,
	})
}

// type=Kerberos,target=%s,computer=%s,ip=%s,service=%s,ticketType=%s,count=%d,failed=%d,status=%s,cert=%s,ft=%s,lt=%s
func checkWinKerberos(h string, m map[string]string) {
	winKerberosCount++
	target, ok := m["target"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := fmt.Sprintf("%s:%s:%s:%s:%s", target, m["ricketType"], m["computer"], m["service"], m["ip"])
	failed := getNumberFromTWLog(m["failed"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinKerberos(id)
	if e != nil {
		e.LastTime = lt
		e.Count += count
		e.Failed += failed
		e.IP = m["ip"]
		return
	}
	datastore.AddWinKerberos(&datastore.WinKerberosEnt{
		ID:        id,
		Target:    target,
		Count:     count,
		Failed:    failed,
		IP:        m["ip"],
		FirstTime: getTimeFromTWLog(m["ft"]),
		LastTime:  lt,
	})
}

// type=Privilege,subject=%s,computer=%s,count=%d,ft=%s,lt=%s
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
		e.Count += count
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

// type=Process,computer=%s,process=%s,count=%d,start=%d,exit=%d,subject=%s,status=%s,parent=%s,ft=%s,lt=%s",

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
		e.Count += count
		e.Start += start
		e.Exit += exit
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

// type=Task,subject=%s,taskname=%s,computer=%s,count=%d,ft=%s,lt=%s
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
		e.Count += count
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
