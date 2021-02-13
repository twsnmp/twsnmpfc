package webapi

import (
	"log"
	"net/http"

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
	api := c.Get("api").(*WebAPI)
	r := logsWebAPI{
		NodeList:  []selectEntWebAPI{},
		EventLogs: []*datastore.EventLogEnt{},
	}
	filter := new(eventLogFilter)
	if err := c.Bind(filter); err != nil {
		log.Printf("postEventLogs err=%v", err)
		return echo.ErrBadRequest
	}
	eventFilter := makeStringFilter(filter.Event)
	levelFilter := getLogLevelFilter(filter.Level)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 24)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)

	api.DataStore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r.NodeList = append(r.NodeList, selectEntWebAPI{Text: n.Name, Value: n.ID})
		return true
	})
	i := 0
	api.DataStore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
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
		return i <= api.DataStore.MapConf.LogDispSize
	})
	return c.JSON(http.StatusOK, r)
}
