package webapi

import (
	"encoding/json"
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

var logLevelMap = map[string]*regexp.Regexp{
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
	var ok bool
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
	var levelFilter *regexp.Regexp
	if filter.Level != "" {
		if levelFilter, ok = logLevelMap[filter.Level]; !ok {
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

type syslogFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	Level     string
	Type      string
	Host      string
	Tag       string
	Message   string
}

type syslogWebAPI struct {
	Time    int64
	Level   string
	Host    string
	Type    string
	Tag     string
	Message string
}

func getLevelFromSeverity(sv int) string {
	if sv < 3 {
		return "high"
	}
	if sv < 4 {
		return "low"
	}
	if sv == 4 {
		return "warn"
	}
	if sv == 7 {
		return "debug"
	}
	return "info"
}

var severityNames = []string{
	"emerg",
	"alert",
	"crit",
	"err",
	"warning",
	"notice",
	"info",
	"debug",
}

var facilityNames = []string{
	"kern",
	"user",
	"mail",
	"daemon",
	"auth",
	"syslog",
	"lpr",
	"news",
	"uucp",
	"cron",
	"authpriv",
	"ftp",
	"ntp",
	"logaudit",
	"logalert",
	"clock",
	"local0",
	"local1",
	"local2",
	"local3",
	"local4",
	"local5",
	"local6",
	"local7",
}

func getSyslogType(sv, fac int) string {
	r := ""
	if sv >= 0 && sv < len(severityNames) {
		r += severityNames[sv]
	} else {
		r += "unknown"
	}
	r += ":"
	if fac >= 0 && fac < len(facilityNames) {
		r += facilityNames[fac]
	} else {
		r += "unknown"
	}
	return r
}

func postSyslog(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := []*syslogWebAPI{}
	var err error
	var ok bool
	filter := new(syslogFilter)
	if err = c.Bind(filter); err != nil {
		log.Printf("postSyslog err=%v", err)
		return echo.ErrBadRequest
	}
	var messageFilter *regexp.Regexp
	if filter.Message != "" {
		if messageFilter, err = regexp.Compile(filter.Message); err != nil {
			log.Printf("postSyslog err=%v", err)
			return echo.ErrBadRequest
		}
	}
	var typeFilter *regexp.Regexp
	if filter.Type != "" {
		if typeFilter, err = regexp.Compile(filter.Type); err != nil {
			log.Printf("postSyslog err=%v", err)
			return echo.ErrBadRequest
		}
	}
	var hostFilter *regexp.Regexp
	if filter.Host != "" {
		if hostFilter, err = regexp.Compile(filter.Host); err != nil {
			log.Printf("postSyslog err=%v", err)
			return echo.ErrBadRequest
		}
	}
	var tagFilter *regexp.Regexp
	if filter.Tag != "" {
		if tagFilter, err = regexp.Compile(filter.Tag); err != nil {
			log.Printf("postSyslog err=%v", err)
			return echo.ErrBadRequest
		}
	}
	var levelFilter *regexp.Regexp
	if filter.Level != "" {
		if levelFilter, ok = logLevelMap[filter.Level]; !ok {
			log.Printf("postSyslog no level filter")
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
	i := 0
	api.DataStore.ForEachLog(st, et, "syslog", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			log.Printf("postSyslog err=%v", err)
			return true
		}
		var ok bool
		re := new(syslogWebAPI)
		if re.Message, ok = sl["content"].(string); !ok {
			log.Printf("postSyslog no content")
			return true
		}
		var sv float64
		if sv, ok = sl["severity"].(float64); !ok {
			log.Printf("postSyslog no severity")
			return true
		}
		var fac float64
		if fac, ok = sl["facility"].(float64); !ok {
			log.Printf("postSyslog no facility")
			return true
		}
		if re.Tag, ok = sl["tag"].(string); !ok {
			log.Printf("postSyslog no tag")
			return true
		}
		if re.Host, ok = sl["hostname"].(string); !ok {
			log.Printf("postSyslog no hostname")
			return true
		}
		re.Time = l.Time
		re.Level = getLevelFromSeverity(int(sv))
		re.Type = getSyslogType(int(sv), int(fac))
		if messageFilter != nil && !messageFilter.Match([]byte(re.Message)) {
			return true
		}
		if tagFilter != nil && !tagFilter.Match([]byte(re.Tag)) {
			return true
		}
		if typeFilter != nil && !typeFilter.Match([]byte(re.Type)) {
			return true
		}
		if levelFilter != nil && !levelFilter.Match([]byte(re.Level)) {
			return true
		}
		if hostFilter != nil && !hostFilter.Match([]byte(re.Host)) {
			return true
		}
		r = append(r, re)
		i++
		return i <= api.DataStore.MapConf.LogDispSize
	})
	return c.JSON(http.StatusOK, r)
}
