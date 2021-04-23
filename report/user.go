package report

import (
	"fmt"
	"net"
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
	if len(u.ClientMap) > 3 {
		u.Penalty++
	}
	for c, e := range u.ClientMap {
		// クライアント毎に場所
		if !checkUserClient(c) {
			u.Penalty++
		}
		if e.Total > 0 {
			r := float32(e.Ok) / float32(e.Total)
			if r < 0.95 {
				u.Penalty++
				if r < 0.8 {
					u.Penalty++
					if r < 0.2 {
						u.Penalty++
					}
					setBadIPFromClient(c)
				}
			}
		}
	}
}

func setBadIPFromClient(c string) {
	if _, err := net.ParseMAC(c); err == nil {
		mac := normMACAddr(c)
		d := datastore.GetDevice(mac)
		if d == nil {
			return
		}
		c = d.IP
	}
	if ip := net.ParseIP(c); ip != nil {
		if _, ok := badIPs[c]; !ok {
			badIPs[c] = true
		}
	}
}

func checkUserReport(ur *userReportEnt) {
	now := time.Now().UnixNano()
	id := fmt.Sprintf("%s:%s", ur.UserID, ur.Server)
	u := datastore.GetUser(id)
	if u != nil {
		if u.ClientMap == nil {
			u.ClientMap = make(map[string]datastore.UserClientEnt)
		}
		e, ok := u.ClientMap[ur.Client]
		if !ok {
			e = datastore.UserClientEnt{}
		}
		u.Total++
		e.Total++
		if ur.Ok {
			e.Ok++
			u.Ok++
		}
		u.ClientMap[ur.Client] = e
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
		ClientMap:  make(map[string]datastore.UserClientEnt),
		Total:      1,
		FirstTime:  ur.Time,
		LastTime:   ur.Time,
		UpdateTime: now,
	}
	u.ServerName, u.ServerNodeID = findNodeInfoFromIP(ur.Server)
	e := datastore.UserClientEnt{Total: 1}
	if ur.Ok {
		e.Ok = 1
		u.Ok = 1
	}
	u.ClientMap[ur.Client] = e
	setUserPenalty(u)
	datastore.AddUser(u)
}

func checkUserClient(client string) bool {
	if _, err := net.ParseMAC(client); err == nil {
		mac := normMACAddr(client)
		d := datastore.GetDevice(mac)
		if d != nil && d.Penalty > 0 {
			return false
		}
		return true
	}
	if ip := net.ParseIP(client); ip == nil {
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
