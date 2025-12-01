package webapi

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type logsWebAPI struct {
	EventLogs []*datastore.EventLogEnt
	NodeList  []selectEntWebAPI
}

type eventLogFilter struct {
	Level     string
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	Type      string
	NodeID    string
	Event     string
}

func postEventLogs(c echo.Context) error {
	r := logsWebAPI{
		NodeList:  []selectEntWebAPI{},
		EventLogs: []*datastore.EventLogEnt{},
	}
	filter := new(eventLogFilter)
	if err := c.Bind(filter); err != nil {
		log.Println(err)
		return echo.ErrBadRequest
	}
	eventFilter := makeStringFilter(filter.Event)
	levelFilter := getLogLevelFilter(filter.Level)
	st := makeStartTimeFilter(filter.StartDate, filter.StartTime)
	et := makeEndTimeFilter(filter.EndDate, filter.EndTime)

	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r.NodeList = append(r.NodeList, selectEntWebAPI{Text: n.Name, Value: n.ID})
		return true
	})
	i := 0
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if eventFilter != nil && !eventFilter.Match([]byte(l.Event)) {
			return true
		}
		if levelFilter != nil && !levelFilter.Match([]byte(l.Level)) {
			return true
		}
		if filter.Type != "" && filter.Type != l.Type {
			return true
		}
		if filter.NodeID != "" && filter.NodeID != l.NodeID {
			return true
		}
		r.EventLogs = append(r.EventLogs, l)
		i++
		return i <= datastore.MapConf.LogDispSize
	})
	// 逆順にする
	for i, j := 0, len(r.EventLogs)-1; i < j; i, j = i+1, j-1 {
		r.EventLogs[i], r.EventLogs[j] = r.EventLogs[j], r.EventLogs[i]
	}
	return c.JSON(http.StatusOK, r)
}

func postLastEventLogs(c echo.Context) error {
	r := []*datastore.EventLogEnt{}
	sts := c.Param("st")
	st := time.Now().Add(-time.Hour * 1).UnixNano()
	if sts != "" {
		if nst, err := strconv.ParseInt(sts, 10, 64); err == nil && nst > st {
			st = nst
		}
	}
	et := time.Now().UnixNano()
	i := 0
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		r = append(r, l)
		i++
		return i <= datastore.MapConf.LogDispSize
	})
	//逆順にソートする
	sort.Slice(r, func(i, j int) bool {
		return r[i].Time > r[j].Time
	})
	return c.JSON(http.StatusOK, r)
}
