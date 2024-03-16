package webapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
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
	NextTime  int64
	Filter    int
}

type syslogWebAPI struct {
	Logs          []*syslogWebAPILogEnt
	ExtractHeader []string
	ExtractDatas  [][]string
	NextTime      int64
	Process       int
	Filter        int
	Limit         int
}

type syslogWebAPILogEnt struct {
	Time     int64
	Level    string
	Host     string
	Type     string
	Tag      string
	Message  string
	Severity int
	Facility int
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
		return echo.ErrBadRequest
	}
	messageFilter := makePipeFilter(filter.Message)
	typeFilter := makeStringFilter(filter.Type)
	hostFilter := makeStringFilter(filter.Host)
	tagFilter := makeStringFilter(filter.Tag)
	levelFilter := getLogLevelFilter(filter.Level)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 1)
	if filter.NextTime > 0 {
		st = filter.NextTime
	}
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	grokCap := ""
	var grokExtractor *grok.Grok
	regExtractor := getExtractType(filter.Extractor)
	log.Printf("regExtractor=%d", regExtractor)
	if regExtractor == none && filter.Extractor != "" {
		grokEnt := datastore.GetGrokEnt(filter.Extractor)
		if grokEnt != nil {
			var err error
			grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
			if err == nil {
				if err = grokExtractor.AddPattern(filter.Extractor, grokEnt.Pat); err == nil {
					grokCap = fmt.Sprintf("%%{%s}", filter.Extractor)
				}
			}
		}
	}
	r.NextTime = 0
	r.Process = 0
	r.Filter = filter.Filter
	r.ExtractDatas = [][]string{}
	r.ExtractHeader = []string{}
	i := 0
	to := 15
	if datastore.MapConf.LogTimeout > 0 {
		to = datastore.MapConf.LogTimeout
	}
	end := time.Now().Unix() + int64(to)
	var hostMap = make(map[string]string)
	datastore.ForEachLog(st, et, "syslog", func(l *datastore.LogEnt) bool {
		if i > 1000 {
			// 検索期間が15秒を超えた場合
			if time.Now().Unix() > end {
				r.NextTime = l.Time
				return false
			}
			i = 0
		}
		i++
		if r.Filter >= datastore.MapConf.LogDispSize {
			// 検索数が表示件数を超えた場合
			r.NextTime = l.Time
			return false
		}
		r.Process++
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		var ok bool
		re := new(syslogWebAPILogEnt)
		var sv float64
		if sv, ok = sl["severity"].(float64); !ok {
			return true
		}
		var fac float64
		if fac, ok = sl["facility"].(float64); !ok {
			return true
		}
		if re.Host, ok = sl["hostname"].(string); !ok {
			return true
		}
		if h, ok := hostMap[re.Host]; ok {
			re.Host = h
		} else {
			if n := datastore.FindNodeFromIP(re.Host); n != nil {
				h = fmt.Sprintf("%s(%s)", re.Host, n.Name)
				hostMap[re.Host] = h
				re.Host = h
			} else {
				hostMap[re.Host] = re.Host
			}
		}
		if re.Tag, ok = sl["tag"].(string); !ok {
			if re.Tag, ok = sl["app_name"].(string); !ok {
				return true
			}
			re.Message = ""
			for i, k := range []string{"proc_id", "msg_id", "message", "structured_data"} {
				if m, ok := sl[k].(string); ok && m != "" {
					if i > 0 {
						re.Message += " "
					}
					re.Message += m
				}
			}
		} else {
			if re.Message, ok = sl["content"].(string); !ok {
				return true
			}
		}
		re.Time = l.Time
		re.Level = getLevelFromSeverity(int(sv))
		re.Type = getSyslogType(int(sv), int(fac))
		re.Facility = int(fac)
		re.Severity = int(sv)
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
		if grokExtractor != nil || regExtractor > 0 {
			var values map[string]string
			var err error
			if grokExtractor != nil {
				values, err = grokExtractor.Parse(grokCap, re.Message)
			} else {
				values, err = regExtract(regExtractor, re.Message)
			}
			if err != nil {
				log.Printf("parse grok err=%v", err)
			} else if len(values) > 0 {
				if len(r.ExtractHeader) < 1 {
					r.ExtractHeader = append(r.ExtractHeader, "TimeStr")
					r.ExtractHeader = append(r.ExtractHeader, "Level")
					r.ExtractHeader = append(r.ExtractHeader, "Type")
					r.ExtractHeader = append(r.ExtractHeader, "Host")
					r.ExtractHeader = append(r.ExtractHeader, "Tag")
					keys := []string{}
					for k := range values {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					r.ExtractHeader = append(r.ExtractHeader, keys...)
				}
				e := []string{}
				for _, k := range r.ExtractHeader {
					switch k {
					case "TimeStr":
						e = append(e, time.Unix(0, l.Time).Format("2006-01-02 15:04:05"))
					case "Host":
						e = append(e, re.Host)
					case "Tag":
						e = append(e, re.Tag)
					case "Level":
						e = append(e, re.Level)
					case "Type":
						e = append(e, re.Type)
					default:
						e = append(e, values[k])
					}
				}
				r.ExtractDatas = append(r.ExtractDatas, e)
			}
		}
		r.Logs = append(r.Logs, re)
		r.Filter++
		return true
	})
	r.Limit = datastore.MapConf.LogDispSize
	return c.JSON(http.StatusOK, r)
}

// Extract by regexp
var reSplunk = regexp.MustCompile(`([-a-zA-Z0-1_]+)=([^, ;]+)`)
var reNumber = regexp.MustCompile(`-?[.0-9]+`)
var reIPv4 = regexp.MustCompile(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`)
var reMAC = regexp.MustCompile(`[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}`)
var reEMail = regexp.MustCompile(`[A-Za-z]+[A-Za-z0-1]+@[A-Za-z.]+`)
var reJSON = regexp.MustCompile(`\{.+\}`)

const (
	none = iota
	splunk
	number
	json_type
	other
)

func getExtractType(t string) int {
	switch t {
	case "re_splunk":
		return splunk
	case "re_number":
		return number
	case "re_json":
		return json_type
	case "re_other":
		return other
	}
	return none
}

func regExtract(t int, msg string) (map[string]string, error) {
	r := make(map[string]string)
	switch t {
	case splunk:
		m := reSplunk.FindAllStringSubmatch(msg, -1)
		for _, e := range m {
			if len(e) > 2 {
				r[e[1]] = e[2]
			}
		}
	case number:
		m := reNumber.FindAllString(msg, -1)
		for i, e := range m {
			k := fmt.Sprintf("number%d", i+1)
			r[k] = e
		}
	case json_type:
		if m := reJSON.FindString(msg); m != "" {
			jd := make(map[string]interface{})
			if err := json.Unmarshal([]byte(m), &jd); err == nil {
				for k, v := range jd {
					r[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	default:
		m := reIPv4.FindAllString(msg, -1)
		for i, e := range m {
			k := fmt.Sprintf("ip%d", i+1)
			r[k] = e
		}
		m = reMAC.FindAllString(msg, -1)
		for i, e := range m {
			k := fmt.Sprintf("mac%d", i+1)
			r[k] = e
		}
		m = reEMail.FindAllString(msg, -1)
		for i, e := range m {
			k := fmt.Sprintf("email%d", i+1)
			r[k] = e
		}
	}
	return r, nil
}
