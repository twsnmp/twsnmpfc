package webapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

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
	filter := new(syslogFilter)
	if err := c.Bind(filter); err != nil {
		log.Printf("postSyslog err=%v", err)
		return echo.ErrBadRequest
	}
	messageFilter := makeStringFilter(filter.Message)
	typeFilter := makeStringFilter(filter.Type)
	hostFilter := makeStringFilter(filter.Host)
	tagFilter := makeStringFilter(filter.Tag)
	levelFilter := getLogLevelFilter(filter.Level)
	st := makeTimeFilter(filter.StartDate, filter.EndDate, 3)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
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
