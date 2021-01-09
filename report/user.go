package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfc/datastore"
)

type userReportEnt struct {
	Time   int64
	UserID string
	Server string
	Client string
	Ok     bool
}

func (r *Report) ReportUser(user, server, client string, ok bool, t int64) {
	r.userReportCh <- &userReportEnt{
		Time:   t,
		Server: server,
		Client: client,
		UserID: user,
		Ok:     ok,
	}
}

func (r *Report) checkUserReport(ur *userReportEnt) {
	now := time.Now().UnixNano()
	id := fmt.Sprintf("%s:%s", ur.UserID, ur.Server)
	u := r.ds.GetUser(id)
	if u != nil {
		u.Total++
		if ur.Ok {
			u.Ok++
		} else {
			u.Penalty++
		}
		if _, ok := u.Clients[ur.Client]; ok {
			u.Clients[ur.Client]++
		} else {
			// 複数の場所からログインは問題
			u.Penalty++
			u.Clients[ur.Client] = 1
			r.checkUserClient(u, ur.Client)
		}
		u.LastTime = ur.Time
		u.UpdateTime = now
		return
	}
	u = &datastore.UserEnt{
		ID:         id,
		UserID:     ur.UserID,
		Server:     ur.Server,
		ServerName: r.findNameFromIP(ur.Server),
		Clients:    make(map[string]int64),
		Total:      1,
		FirstTime:  ur.Time,
		LastTime:   ur.Time,
		UpdateTime: now,
	}
	u.Clients[ur.Client] = 1
	r.checkUserClient(u, ur.Client)
	if ur.Ok {
		u.Ok = 1
	} else {
		u.Penalty = 1
	}
	r.ds.AddUser(u)
}

func (r *Report) checkUserClient(u *datastore.UserEnt, client string) {
	if !strings.Contains(client, ".") {
		return
	}
	loc := r.ds.GetLoc(client)
	a := strings.Split(loc, ",")
	if len(a) > 0 {
		loc = a[0]
	}
	// DNSで解決できない場合
	if client == r.findNameFromIP(client) {
		u.Penalty++
	}
	if loc != "" && loc != "LOCAL" {
		id := fmt.Sprintf("*:*:%s", loc)
		if r.ds.GetDennyRule(id) {
			u.Penalty++
		}
	}
	if u.Penalty > 0 {
		if _, ok := r.badIPs[client]; !ok {
			r.badIPs[client] = u.Penalty
		}
	}
}
