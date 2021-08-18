package datastore

// type=EventID,computer=%s,channel=%s,provider=%s,eventID=%d,total=%d,count=%d,ft=%s,lt=%s
type WinEventIDEnt struct {
	ID        string // Computer + Provider + EventID
	Level     string
	Computer  string
	Provider  string
	Channel   string
	EventID   int
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetWinEventID(id string) *WinEventIDEnt {
	if v, ok := winEventID.Load(id); ok {
		return v.(*WinEventIDEnt)
	}
	return nil
}

func AddWinEventID(e *WinEventIDEnt) {
	winEventID.Store(e.ID, e)
}

func ForEachWinEventID(f func(*WinEventIDEnt) bool) {
	winEventID.Range(func(k, v interface{}) bool {
		e := v.(*WinEventIDEnt)
		return f(e)
	})
}

// type=Logon,target=%s,computer=%s,ip=%s,count=%d,logon=%d,failed=%d,logoff=%d%s%s,ft=%s,lt=%s",
// type=LogonFailed,subject=%s@%s,target=%s@%s,targetsid=%s,logonType=%s,ip=%s,code=%s,time=%s",
type WinLogonEnt struct {
	ID        string // target + computer + IP
	Target    string
	Computer  string
	IP        string
	Count     int64
	Logon     int64
	Logoff    int64
	Failed    int64
	FirstTime int64
	LastTime  int64
}

func GetWinLogon(id string) *WinLogonEnt {
	if v, ok := winLogon.Load(id); ok {
		return v.(*WinLogonEnt)
	}
	return nil
}

func AddWinLogon(e *WinLogonEnt) {
	winLogon.Store(e.ID, e)
}

func ForEachWinLogon(f func(*WinLogonEnt) bool) {
	winLogon.Range(func(k, v interface{}) bool {
		e := v.(*WinLogonEnt)
		return f(e)
	})
}

// type=Account,subject=%s,target=%s,computer=%s,count=%d,edit=%d,password=%d,other=%d,ft=%s,lt=%s",
type WinAccountEnt struct {
	ID        string // subject + target + computer
	Subject   string
	Target    string
	Computer  string
	Count     int64
	Edit      int64
	Password  int64
	Other     int64
	FirstTime int64
	LastTime  int64
}

func GetWinAccount(id string) *WinAccountEnt {
	if v, ok := winAccount.Load(id); ok {
		return v.(*WinAccountEnt)
	}
	return nil
}

func AddWinAccount(e *WinAccountEnt) {
	winAccount.Store(e.ID, e)
}

func ForEachWinAccount(f func(*WinAccountEnt) bool) {
	winAccount.Range(func(k, v interface{}) bool {
		e := v.(*WinAccountEnt)
		return f(e)
	})
}

// type=Kerberos,target=%s,computer=%s,ip=%s,service=%s,ticketType=%s,count=%d,failed=%d,status=%s,cert=%s,ft=%s,lt=%s
// type=KerberosFailed,target=%s,computer=%s,ip=%s,service=%s,ticketType=%s,status=%s,time=%s
type WinKerberosEnt struct {
	ID         string // target + computer + service + ip
	Target     string
	Computer   string
	IP         string
	Service    string
	TicketType string
	Count      int64
	Failed     int64
	LastCert   string
	LastStatus string
	FirstTime  int64
	LastTime   int64
}

func GetWinKerberos(id string) *WinKerberosEnt {
	if v, ok := winKerberos.Load(id); ok {
		return v.(*WinKerberosEnt)
	}
	return nil
}

func AddWinKerberos(e *WinKerberosEnt) {
	winKerberos.Store(e.ID, e)
}

func ForEachWinKerberos(f func(*WinKerberosEnt) bool) {
	winKerberos.Range(func(k, v interface{}) bool {
		e := v.(*WinKerberosEnt)
		return f(e)
	})
}

// type=Privilege,subject=%s,computer=%s,count=%d,ft=%s,lt=%s
type WinPrivilegeEnt struct {
	ID        string //  subject + computer
	Subject   string
	Computer  string
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetWinPrivilege(id string) *WinPrivilegeEnt {
	if v, ok := winPrivilege.Load(id); ok {
		return v.(*WinPrivilegeEnt)
	}
	return nil
}

func AddWinPrivilege(e *WinPrivilegeEnt) {
	winPrivilege.Store(e.ID, e)
}

func ForEachWinPrivilege(f func(*WinPrivilegeEnt) bool) {
	winPrivilege.Range(func(k, v interface{}) bool {
		e := v.(*WinPrivilegeEnt)
		return f(e)
	})
}

// type=Process,computer=%s,process=%s,count=%d,start=%d,exit=%d,subject=%s,status=%s,parent=%s,ft=%s,lt=%s",
type WinProcessEnt struct {
	ID          string // Computer + Process
	Computer    string
	Process     string
	Count       int64
	Start       int64
	Exit        int64
	LastParent  string
	LastSubject string
	LastStatus  string
	FirstTime   int64
	LastTime    int64
}

func GetWinProcess(id string) *WinProcessEnt {
	if v, ok := winProcess.Load(id); ok {
		return v.(*WinProcessEnt)
	}
	return nil
}

func AddWinProcess(e *WinProcessEnt) {
	winProcess.Store(e.ID, e)
}

func ForEachWinProcess(f func(*WinProcessEnt) bool) {
	winProcess.Range(func(k, v interface{}) bool {
		e := v.(*WinProcessEnt)
		return f(e)
	})
}

// type=Task,subject=%s,taskname=%s,computer=%s,count=%d,ft=%s,lt=%s",
type WinTaskEnt struct {
	ID        string // Computer + TaskName
	TaskName  string
	Computer  string
	Subject   string
	Count     int64
	FirstTime int64
	LastTime  int64
}

func GetWinTask(id string) *WinTaskEnt {
	if v, ok := winTask.Load(id); ok {
		return v.(*WinTaskEnt)
	}
	return nil
}

func AddWinTask(e *WinTaskEnt) {
	winTask.Store(e.ID, e)
}

func ForEachWinTask(f func(*WinTaskEnt) bool) {
	winTask.Range(func(k, v interface{}) bool {
		e := v.(*WinTaskEnt)
		return f(e)
	})
}
