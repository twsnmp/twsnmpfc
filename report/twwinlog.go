package report

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

func ReportTwWinLog(l map[string]interface{}) {
	twWinLogCh <- l
}

func ResetWinLogonScore() {
	datastore.ForEachWinLogon(func(e *datastore.WinLogonEnt) bool {
		setWinLogonPenalty(e)
		return true
	})
	calcWinLogonScore()
}

func ResetWinKerberosScore() {
	datastore.ForEachWinKerberos(func(e *datastore.WinKerberosEnt) bool {
		setWinKerberosPenalty(e)
		return true
	})
	calcWinKerberosScore()
}

func checkTWWinLogReport(l map[string]interface{}) {
	h, ok := l["hostname"].(string)
	if !ok {
		return
	}
	msg, ok := l["content"].(string)
	if !ok {
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
		return
	}
	switch t {
	case "Stats":
		checkStats(h, "twwinlog", m)
	case "Monitor":
		checkMonitor(h, "twwinlog", m)
	case "EventID":
		checkWinEventID(m, l)
	case "Logon", "Logoff", "LogonFailed":
		checkWinLogon(m)
	case "Account":
		checkWinAccount(m)
	case "Kerberos":
		checkWinKerberos(m)
	case "Privilege":
		checkWinPrivilege(m)
	case "Process":
		checkWinProcess(m)
	case "Task":
		checkWinTask(m)
	default:
		log.Printf("twwinlog unknown type=%s", t)
	}
}

// type=EventID,computer=%s,channel=%s,provider=%s,eventID=%d,total=%d,count=%d,ft=%s,lt=%s
func checkWinEventID(m map[string]string, l map[string]interface{}) {
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
	id := makeID(fmt.Sprintf("%s:%s:%d", computer, provider, eventID))
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

	id := makeID(fmt.Sprintf("%s:%s:%s", target, m["computer"], m["ip"]))
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
		setWinLogonPenalty(e)
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
	setWinLogonPenalty(e)
	datastore.AddWinLogon(e)
}

func setWinLogonPenalty(e *datastore.WinLogonEnt) {
	if e.Failed > 0 {
		e.Penalty = 1
	}
	if e.Count > 0 {
		e.Penalty += (10 * e.Failed) / e.Count
	}
}

func calcWinLogonScore() {
	var xs []float64
	datastore.ForEachWinLogon(func(e *datastore.WinLogonEnt) bool {
		if e.Penalty > 100 {
			e.Penalty = 100
		}
		xs = append(xs, float64(100-e.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachWinLogon(func(e *datastore.WinLogonEnt) bool {
		if sd != 0 {
			e.Score = ((10 * (float64(100-e.Penalty) - m) / sd) + 50)
		} else {
			e.Score = 50.0
		}
		e.ValidScore = true
		return true
	})
}

// type=Account,subject=%s,target=%s,computer=%s,count=%d,edit=%d,password=%d,other=%d,ft=%s,lt=%s",
func checkWinAccount(m map[string]string) {
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
	id := makeID(fmt.Sprintf("%s:%s:%s", target, m["computer"], m["subject"]))
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
func checkWinKerberos(m map[string]string) {
	target, ok := m["target"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := makeID(fmt.Sprintf("%s:%s:%s:%s:%s", target, m["computer"], m["ip"], m["service"], m["ticketType"]))
	failed := getNumberFromTWLog(m["failed"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinKerberos(id)
	if e != nil {
		e.LastTime = lt
		e.Count += count
		e.Failed += failed
		setWinKerberosPenalty(e)
		return
	}
	e = &datastore.WinKerberosEnt{
		ID:         id,
		Target:     target,
		Computer:   m["computer"],
		Service:    m["service"],
		TicketType: m["ticketType"],
		IP:         m["ip"],
		Count:      count,
		Failed:     failed,
		FirstTime:  getTimeFromTWLog(m["ft"]),
		LastTime:   lt,
	}
	setWinKerberosPenalty(e)
	datastore.AddWinKerberos(e)
}

func setWinKerberosPenalty(e *datastore.WinKerberosEnt) {
	if e.Failed > 0 {
		e.Penalty = 1
	}
	if e.Count > 0 {
		e.Penalty += (10 * e.Failed) / e.Count
	}
}

func calcWinKerberosScore() {
	var xs []float64
	datastore.ForEachWinKerberos(func(e *datastore.WinKerberosEnt) bool {
		if e.Penalty > 100 {
			e.Penalty = 100
		}
		xs = append(xs, float64(100-e.Penalty))
		return true
	})
	m, sd := getMeanSD(&xs)
	datastore.ForEachWinKerberos(func(e *datastore.WinKerberosEnt) bool {
		if sd != 0 {
			e.Score = ((10 * (float64(100-e.Penalty) - m) / sd) + 50)
		} else {
			e.Score = 50.0
		}
		e.ValidScore = true
		return true
	})
}

// type=Privilege,subject=%s,computer=%s,count=%d,ft=%s,lt=%s
func checkWinPrivilege(m map[string]string) {
	subject, ok := m["subject"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := makeID(fmt.Sprintf("%s:%s", subject, m["computer"]))
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
func checkWinProcess(m map[string]string) {
	process, ok := m["process"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := makeID(fmt.Sprintf("%s:%s", process, m["computer"]))
	start := getNumberFromTWLog(m["start"])
	exit := getNumberFromTWLog(m["exit"])
	lt := getTimeFromTWLog(m["lt"])
	e := datastore.GetWinProcess(id)
	if e != nil {
		e.LastTime = lt
		e.Count += count
		e.Start += start
		e.Exit += exit
		if v, ok := m["subject"]; ok && v != "" {
			e.LastSubject = v
		}
		if v, ok := m["status"]; ok && v != "" {
			e.LastStatus = v
		}
		if v, ok := m["parent"]; ok && v != "" {
			e.LastParent = v
		}
		return
	}
	datastore.AddWinProcess(&datastore.WinProcessEnt{
		ID:          id,
		Process:     process,
		Computer:    m["computer"],
		Count:       count,
		Start:       start,
		Exit:        exit,
		LastSubject: m["subject"],
		LastParent:  m["parent"],
		LastStatus:  m["status"],
		FirstTime:   getTimeFromTWLog(m["ft"]),
		LastTime:    lt,
	})
}

// type=Task,subject=%s,taskname=%s,computer=%s,count=%d,ft=%s,lt=%s
func checkWinTask(m map[string]string) {
	taskname, ok := m["taskname"]
	if !ok {
		return
	}
	count := getNumberFromTWLog(m["count"])
	if count < 1 {
		return
	}
	id := makeID(fmt.Sprintf("%s:%s:%s", taskname, m["computer"], m["subject"]))
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

func checkOldWinEventID() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachWinEventID(func(e *datastore.WinEventIDEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("winEventID", ids)
	}
}

func checkOldWinLogon() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachWinLogon(func(e *datastore.WinLogonEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("winLogon", ids)
	}
}

func checkOldWinAccount() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachWinAccount(func(e *datastore.WinAccountEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("winAccount", ids)
	}
}

func checkOldWinKerberos() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachWinKerberos(func(e *datastore.WinKerberosEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("winKerberos", ids)
	}
}

func checkOldWinPrivilege() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachWinPrivilege(func(e *datastore.WinPrivilegeEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("winPrivilege", ids)
	}
}

func checkOldWinProcess() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachWinProcess(func(e *datastore.WinProcessEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("winProcess", ids)
	}
}

func checkOldWinTask() {
	ids := []string{}
	delOld := time.Now().AddDate(0, 0, -datastore.ReportConf.ReportDays).UnixNano()
	datastore.ForEachWinTask(func(e *datastore.WinTaskEnt) bool {
		if e.LastTime < delOld {
			ids = append(ids, e.ID)
		}
		return true
	})
	if len(ids) > 0 {
		datastore.DeleteReport("winTask", ids)
	}
}

func makeID(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}
