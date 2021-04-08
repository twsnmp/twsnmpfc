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

func ReportUser(user, server, client string, ok bool, t int64) {
	userReportCh <- &userReportEnt{
		Time:   t,
		Server: server,
		Client: client,
		UserID: user,
		Ok:     ok,
	}
}

func ResetUsersScore() {
	datastore.ForEachUsers(func(u *datastore.UserEnt) bool {
		setUserPenalty(u)
		u.UpdateTime = time.Now().UnixNano()
		return true
	})
	calcUserScore()
}

func setUserPenalty(u *datastore.UserEnt) {
	u.Penalty = 0
	if len(u.Clients) > 5 {
		u.Penalty++
	}
	for c, p := range u.Clients {
		// クライアント毎に場所
		if !checkUserClient(c) {
			u.Penalty++
		}
		if p > 0 {
			u.Penalty++
			if n, ok := badIPs[c]; !ok && n < 5 {
				badIPs[c]++
			}
		}
	}
}

func checkUserReport(ur *userReportEnt) {
	now := time.Now().UnixNano()
	id := fmt.Sprintf("%s:%s", ur.UserID, ur.Server)
	u := datastore.GetUser(id)
	if u != nil {
		u.Total++
		if ur.Ok {
			u.Ok++
		}
		if _, ok := u.Clients[ur.Client]; !ok {
			u.Clients[ur.Client] = 0
		}
		if ur.Ok {
			u.Clients[ur.Client]--
		} else {
			u.Clients[ur.Client]++
		}
		setUserPenalty(u)
		if u.ServerName == "" {
			u.ServerName, u.ServerNodeID = findNodeInfoFromIP(ur.Server)
		}
		u.LastTime = ur.Time
		u.UpdateTime = now
		return
	}
	u = &datastore.UserEnt{
		ID:         id,
		UserID:     ur.UserID,
		Server:     ur.Server,
		Clients:    make(map[string]int64),
		Total:      1,
		FirstTime:  ur.Time,
		LastTime:   ur.Time,
		UpdateTime: now,
	}
	u.ServerName, u.ServerNodeID = findNodeInfoFromIP(ur.Server)
	if ur.Ok {
		u.Clients[ur.Client] = -1
	} else {
		u.Clients[ur.Client] = 1
	}
	if ur.Ok {
		u.Ok = 1
	}
	setUserPenalty(u)
	datastore.AddUser(u)
}

func checkUserClient(client string) bool {
	if !strings.Contains(client, ".") {
		return false
	}
	loc := datastore.GetLoc(client)
	if !isSafeCountry(loc) {
		return false
	}
	// DNSで解決できない場合
	name, _ := findNodeInfoFromIP(client)
	return client == name
}
