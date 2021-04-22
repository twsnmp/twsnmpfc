package webapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/vjeantet/grok"
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
	Extractor string
}

type syslogWebAPI struct {
	Logs          []*syslogWebAPILogEnt
	ExtractHeader []string
	ExtractDatas  [][]string
}

type syslogWebAPILogEnt struct {
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
	r := new(syslogWebAPI)
	filter := new(syslogFilter)
	if err := c.Bind(filter); err != nil {
		log.Printf("postSyslog err=%v", err)
		return echo.ErrBadRequest
	}
	messageFilter := makePipeFilter(filter.Message)
	typeFilter := makeStringFilter(filter.Type)
	hostFilter := makeStringFilter(filter.Host)
	tagFilter := makeStringFilter(filter.Tag)
	levelFilter := getLogLevelFilter(filter.Level)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 3)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	grokCap := ""
	var grokExtractor *grok.Grok
	if filter.Extractor != "" {
		grokEnt := datastore.GetGrokEnt(filter.Extractor)
		if grokEnt != nil {
			var err error
			grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
			if err == nil {
				if err = grokExtractor.AddPattern(filter.Extractor, grokEnt.Pat); err == nil {
					grokCap = fmt.Sprintf("%%{%s}", filter.Extractor)
				}
			}
		} else {
			log.Printf("no grok %s", filter.Extractor)
		}
	}
	i := 0
	r.ExtractDatas = [][]string{}
	r.ExtractHeader = []string{}
	datastore.ForEachLog(st, et, "syslog", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			log.Printf("postSyslog err=%v", err)
			return true
		}
		var ok bool
		re := new(syslogWebAPILogEnt)
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
		for _, mf := range messageFilter {
			if mf.reg.Match([]byte(re.Message)) {
				if mf.not {
					return true
				}
			} else {
				if !mf.not {
					return true
				}
			}
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
		if grokExtractor != nil {
			values, err := grokExtractor.Parse(grokCap, re.Message)
			if err != nil {
				log.Printf("grock err=%v", err)
			} else if len(values) > 0 {
				if len(r.ExtractHeader) < 1 {
					r.ExtractHeader = append(r.ExtractHeader, "TimeStr")
					for k := range values {
						r.ExtractHeader = append(r.ExtractHeader, k)
						sort.Strings(r.ExtractHeader)
					}
				}
				e := []string{}
				for _, k := range r.ExtractHeader {
					if k == "TimeStr" {
						e = append(e, time.Unix(0, l.Time).Format("2006/01/02T15:04:05"))
					} else {
						e = append(e, values[k])
					}
				}
				r.ExtractDatas = append(r.ExtractDatas, e)
			}
		}
		r.Logs = append(r.Logs, re)
		i++
		return i <= datastore.MapConf.LogDispSize
	})
	return c.JSON(http.StatusOK, r)
}
