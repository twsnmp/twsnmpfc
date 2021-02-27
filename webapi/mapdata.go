package webapi

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type mapWebAPI struct {
	LastUpdate int64
	MapConf    *datastore.MapConfEnt
	Nodes      map[string]*datastore.NodeEnt
	Lines      []*datastore.LineEnt
	Pollings   map[string][]*datastore.PollingEnt
	Logs       []*datastore.EventLogEnt
}

func getMap(c echo.Context) error {
	r := &mapWebAPI{
		MapConf:  &datastore.MapConf,
		Nodes:    make(map[string]*datastore.NodeEnt),
		Lines:    []*datastore.LineEnt{},
		Pollings: make(map[string][]*datastore.PollingEnt),
	}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r.Nodes[n.ID] = n
		return true
	})
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		r.Lines = append(r.Lines, l)
		return true
	})
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		r.Pollings[p.NodeID] = append(r.Pollings[p.NodeID], p)
		return true
	})
	i := 0
	datastore.ForEachLastEventLog("", func(e *datastore.EventLogEnt) bool {
		r.Logs = append(r.Logs, e)
		i++
		return i < 100
	})
	r.LastUpdate = time.Now().Unix()
	return c.JSON(http.StatusOK, r)
}

type nodePosWebAPI struct {
	ID string
	X  int
	Y  int
}

func postMapUpdate(c echo.Context) error {
	list := []nodePosWebAPI{}
	if err := c.Bind(&list); err != nil {
		log.Printf("postNodePosUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	for _, nu := range list {
		n := datastore.GetNode(nu.ID)
		if n == nil {
			log.Printf("postNodePosUpdate Node not found ID=%s", nu.ID)
			return echo.ErrBadRequest
		}
		n.X = nu.X
		n.Y = nu.Y
		if err := datastore.UpdateNode(n); err != nil {
			log.Printf("postNodePosUpdate err=%v", err)
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postMapDelete(c echo.Context) error {
	list := []string{}
	if err := c.Bind(&list); err != nil {
		log.Printf("postMapDelete err=%v", err)
		return echo.ErrBadRequest
	}
	for _, id := range list {
		if err := datastore.DeleteNode(id); err != nil {
			log.Printf("postMapDelete err=%v", err)
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getNodes(c echo.Context) error {
	r := []*datastore.NodeEnt{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r = append(r, n)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func postNodeDelete(c echo.Context) error {
	id := new(idWebAPI)
	if err := c.Bind(id); err != nil {
		log.Printf("postNodeDelete err=%v", err)
		return echo.ErrBadRequest
	}
	if err := datastore.DeleteNode(id.ID); err != nil {
		log.Printf("postNodeDelete err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNodeUpdate(c echo.Context) error {
	nu := new(datastore.NodeEnt)
	if err := c.Bind(nu); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	if nu.ID == "" {
		if err := datastore.AddNode(nu); err != nil {
			log.Printf("postNodeUpdate err=%v", err)
			return echo.ErrBadRequest
		}
		return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
	}
	// ここで入力チェック
	n := datastore.GetNode(nu.ID)
	if n == nil {
		log.Printf("postNodeUpdate Node not found ID=%s", nu.ID)
		return echo.ErrBadRequest
	}
	n.Name = nu.Name
	n.Descr = nu.Descr
	n.IP = nu.IP
	n.Icon = nu.Icon
	n.SnmpMode = nu.SnmpMode
	n.Community = nu.Community
	n.User = nu.User
	n.Password = nu.Password
	n.PublicKey = nu.PublicKey
	n.URL = nu.URL
	n.Type = nu.Type
	n.AddrMode = nu.AddrMode
	if err := datastore.UpdateNode(n); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postLineDelete(c echo.Context) error {
	l := new(datastore.LineEnt)
	if err := c.Bind(l); err != nil {
		log.Printf("postLineDelete err=%v", err)
		return echo.ErrBadRequest
	}
	if err := datastore.DeleteLine(l.ID); err != nil {
		log.Printf("postLineDelete err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postLineAdd(c echo.Context) error {
	lu := new(datastore.LineEnt)
	if err := c.Bind(lu); err != nil {
		log.Printf("postLineAdd err=%v", err)
		return echo.ErrBadRequest
	}
	if p := datastore.GetPolling(lu.PollingID1); p != nil {
		lu.State1 = p.State
	}
	if p := datastore.GetPolling(lu.PollingID2); p != nil {
		lu.State2 = p.State
	}
	l := datastore.GetLine(lu.ID)
	if l == nil {
		if err := datastore.AddLine(lu); err != nil {
			log.Printf("postLineAdd err=%v", err)
			return echo.ErrBadRequest
		}
	} else {
		l.NodeID1 = lu.NodeID1
		l.NodeID2 = lu.NodeID2
		l.PollingID1 = lu.PollingID1
		l.PollingID2 = lu.PollingID2
		l.State1 = lu.State1
		l.State2 = lu.State2
		if err := datastore.UpdateLine(l); err != nil {
			log.Printf("postLineAdd err=%v", err)
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type pollingsWebAPI struct {
	Pollings []*datastore.PollingEnt
	NodeList []selectEntWebAPI
}

func getPollings(c echo.Context) error {
	r := pollingsWebAPI{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r.NodeList = append(r.NodeList, selectEntWebAPI{Text: n.Name, Value: n.ID})
		return true
	})
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		r.Pollings = append(r.Pollings, p)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func postPollingDelete(c echo.Context) error {
	id := new(idWebAPI)
	if err := c.Bind(id); err != nil {
		log.Printf("postPollingDelete err=%v", err)
		return echo.ErrBadRequest
	}
	if err := datastore.DeletePolling(id.ID); err != nil {
		log.Printf("postPollingDelete err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postPollingUpdate(c echo.Context) error {
	pu := new(datastore.PollingEnt)
	if err := c.Bind(pu); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	// ここで入力データのチェックをする
	p := datastore.GetPolling(pu.ID)
	if p == nil {
		log.Printf("postPollingUpdate Node not found ID=%s", pu.ID)
		return echo.ErrBadRequest
	}
	p.Name = pu.Name
	p.NodeID = pu.NodeID
	p.Type = pu.Type
	p.Polling = pu.Polling
	p.Level = pu.Level
	p.PollInt = pu.PollInt
	p.Timeout = pu.Timeout
	p.Retry = pu.Retry
	p.LogMode = pu.LogMode
	p.NextTime = 0
	p.State = "unknown"
	if err := datastore.UpdatePolling(p); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postPollingAdd(c echo.Context) error {
	p := new(datastore.PollingEnt)
	if err := c.Bind(p); err != nil {
		log.Printf("postPollingAdd err=%v", err)
		return echo.ErrBadRequest
	}
	// ここで入力データのチェックをする
	p.NextTime = 0
	p.State = "unknown"
	if err := datastore.AddPolling(p); err != nil {
		log.Printf("postPollingAdd err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type nodeWebAPI struct {
	Node     *datastore.NodeEnt
	Logs     []*datastore.EventLogEnt
	Pollings []*datastore.PollingEnt
}

func getNode(c echo.Context) error {
	id := c.Param("id")
	r := nodeWebAPI{}
	r.Node = datastore.GetNode(id)
	if r.Node == nil {
		log.Printf("node not found")
		return echo.ErrBadRequest
	}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == id {
			r.Pollings = append(r.Pollings, p)
		}
		return true
	})
	i := 0
	st := time.Now().Add(-time.Hour * 24).UnixNano()
	et := time.Now().UnixNano()
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if l.NodeID != id {
			return true
		}
		r.Logs = append(r.Logs, l)
		i++
		return i <= datastore.MapConf.LogDispSize
	})

	return c.JSON(http.StatusOK, r)
}

type pollingWebAPI struct {
	Node    *datastore.NodeEnt
	Polling *datastore.PollingEnt
	Logs    []*datastore.PollingLogEnt
}

type timeFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
}

func postPolling(c echo.Context) error {
	id := c.Param("id")
	r := pollingWebAPI{}
	r.Polling = datastore.GetPolling(id)
	if r.Polling == nil {
		log.Printf("polling not found id=%s", id)
		return echo.ErrBadRequest
	}
	r.Node = datastore.GetNode(r.Polling.NodeID)
	if r.Node == nil {
		log.Printf("node not found id=%s", r.Polling.NodeID)
		return echo.ErrBadRequest
	}
	filter := new(timeFilter)
	if err := c.Bind(filter); err != nil {
		log.Printf("postEventLogs err=%v", err)
		return echo.ErrBadRequest
	}
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 24)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	log.Printf("%d %d %v", st, et, filter)
	i := 0
	datastore.ForEachPollingLog(st, et, id, func(l *datastore.PollingLogEnt) bool {
		r.Logs = append(r.Logs, l)
		i++
		return i <= datastore.MapConf.LogDispSize
	})
	return c.JSON(http.StatusOK, r)
}
