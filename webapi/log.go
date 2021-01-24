package webapi

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
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

var eventLogLevelMap = map[string]*regexp.Regexp{
	"high": regexp.MustCompile("high"),
	"low":  regexp.MustCompile("(high|low)"),
	"warn": regexp.MustCompile("(high|low|warn)"),
}

func postEventLogs(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := logsWebAPI{
		NodeList:  []selectEntWebAPI{},
		EventLogs: []*datastore.EventLogEnt{},
	}
	var err error
	filter := new(eventLogFilter)
	if err = c.Bind(filter); err != nil {
		log.Printf("postEventLogs err=%v", err)
		return echo.ErrBadRequest
	}
	var eventFilter *regexp.Regexp
	if filter.Event != "" {
		if eventFilter, err = regexp.Compile(filter.Event); err != nil {
			log.Printf("postEventLogs err=%v", err)
			return echo.ErrBadRequest
		}
	}
	if filter.Level != "" {
		if _, ok := eventLogLevelMap[filter.Level]; !ok {
			log.Printf("postEventLogs err=%v", err)
			return echo.ErrBadRequest
		}
	}
	st := time.Now().Add(-time.Hour * 24).UnixNano()
	if t, err := time.Parse("2006-01-02T15:04 MST", fmt.Sprintf("%sT%s JST", filter.StartDate, filter.StartTime)); err == nil {
		st = t.UnixNano()
	}
	et := time.Now().UnixNano()
	if t, err := time.Parse("2006-01-02T15:04 MST", fmt.Sprintf("%sT%s JST", filter.EndDate, filter.EndTime)); err == nil {
		et = t.UnixNano()
	}

	api.DataStore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r.NodeList = append(r.NodeList, selectEntWebAPI{Text: n.Name, Value: n.ID})
		return true
	})
	api.DataStore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if eventFilter != nil && !eventFilter.Match([]byte(l.Event)) {
			return true
		}
		if filter.Level != "" && filter.Level != l.Level {
			return true
		}
		if filter.Type != "" && filter.Type != l.Type {
			return true
		}
		r.EventLogs = append(r.EventLogs, l)
		return true
	})
	return c.JSON(http.StatusOK, r)
}
