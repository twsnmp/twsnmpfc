package datastore

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

// type=Logon,target=%s,sid=%s,count=%d,logon=%d,failed=%d,logoff=%d,changeSubject=%d,changeLogonType=%d,changeIP=%d,subject=%s,subsid=%s,logonType=%s,ip=%s,failCode=%s,ft=%s,lt=%s

type WinLogonEnt struct {
	ID              string // target
	SID             string
	Count           int64
	Logon           int64
	Logoff          int64
	Failed          int64
	ChangeSubject   int64
	ChangeLogonType int64
	ChangeIP        int64
	LastSubject     string
	LastSubjectSID  string
	LastIP          string
	LastLogonType   string
	LastFailedCode  string
	FirstTime       int64
	LastTime        int64
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

// type=Account,target=%s,sid=%s,computer=%s,count=%d,edit=%d,password=%d,other=%d,changesubject=%d,subject=%s,sbjectsid=%s,ft=%s,lt=%s

type WinAccountEnt struct {
	ID             string // host + target
	Host           string
	Target         string
	SID            string
	Computer       string
	Count          int64
	Edit           int64
	Password       int64
	Other          int64
	ChangeSubject  int64
	LastSubject    string
	LastSubjectSID string
	FirstTime      int64
	LastTime       int64
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

// type=KerberosTGT,target=%s,sid=%s,ip=%s,computer=%s,count=%d,failed=%d,changeStatus=%d,changeCert=%d,status=%s,cert=%s,ft=%s,lt=%s
type WinKerberosTGTEnt struct {
	ID         string // target + ip
	Target     string
	SID        string
	IP         string
	Computer   string
	Count      int64
	Failed     int64
	LastCert   string
	LastStatus string
	FirstTime  int64
	LastTime   int64
}

func GetWinKerberosTGT(id string) *WinKerberosTGTEnt {
	if v, ok := winKerberosTGT.Load(id); ok {
		return v.(*WinKerberosTGTEnt)
	}
	return nil
}

func AddWinKerberosTGT(e *WinKerberosTGTEnt) {
	winKerberosTGT.Store(e.ID, e)
}

func ForEachWinKerberosTGT(f func(*WinKerberosTGTEnt) bool) {
	winKerberosTGT.Range(func(k, v interface{}) bool {
		e := v.(*WinKerberosTGTEnt)
		return f(e)
	})
}

// type=KerberosST,target=%s,servcie=%s,sid=%s,ip=%s,computer=%s,count=%d,failed=%d,changeStatus=%d,status=%s,ft=%s,lt=%s
type WinKerberosSTEnt struct {
	ID           string // host + target
	Host         string
	Target       string
	Service      string
	SID          string
	IP           string
	Computer     string
	Count        int64
	Failed       int64
	ChangeStatus int64
	LastStatus   string
	FirstTime    int64
	LastTime     int64
}

func GetWinKerberosST(id string) *WinKerberosSTEnt {
	if v, ok := winKerberosST.Load(id); ok {
		return v.(*WinKerberosSTEnt)
	}
	return nil
}

func AddWinKerberosST(e *WinKerberosSTEnt) {
	winKerberosST.Store(e.ID, e)
}

func ForEachWinKerberosST(f func(*WinKerberosSTEnt) bool) {
	winKerberosST.Range(func(k, v interface{}) bool {
		e := v.(*WinKerberosSTEnt)
		return f(e)
	})
}

// type=Privilege,subject=%s,sid=%s,computer=%s,count=%d,ft=%s,lt=%s
type WinPrivilegeEnt struct {
	ID         string // host + target
	Host       string
	Subject    string
	SID        string
	Computer   string
	Count      int64
	LastStatus string
	FirstTime  int64
	LastTime   int64
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

// type=Process,computer=%s,process=%s,count=%d,start=%d,exit=%d,changeSubject=%d,changeStatus=%d,changeParent=%d,subject=%s,status=%s,parent=%s,ft=%s,lt=%s
type WinProcessEnt struct {
	ID            string // Computer + Process
	Computer      string
	Process       string
	Count         int64
	Start         int64
	Exit          int64
	ChangeSubject int64
	ChangeStatus  int64
	ChangeParent  int64
	LastParent    string
	LastSubject   string
	LastStatus    string
	FirstTime     int64
	LastTime      int64
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

// type=Task,taskname=%s,computer=%s,subject=%s,sid=%s,count=%d,ft=%s,lt=%s
type WinTaskEnt struct {
	ID        string // Computer + TaskName
	TaskName  string
	Computer  string
	Subject   string
	SID       string
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

func ForEachTask(f func(*WinTaskEnt) bool) {
	winTask.Range(func(k, v interface{}) bool {
		e := v.(*WinTaskEnt)
		return f(e)
	})
}
