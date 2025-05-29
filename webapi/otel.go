package webapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type OTelMetricEnt struct {
	ID      string `json:"ID"`
	Host    string `json:"Host"`
	Service string `json:"Service"`
	Scope   string `json:"Scope"`
	Name    string `json:"Name"`
	Type    string `json:"Type"`
	Count   int    `json:"Count"`
	First   int64  `json:"First"`
	Last    int64  `json:"Last"`
}

func getOTelMetrics(c echo.Context) error {
	r := []*OTelMetricEnt{}
	datastore.ForEachOTelMetric(func(id string, m *datastore.OTelMetricEnt) bool {
		r = append(r, &OTelMetricEnt{
			ID:      id,
			Host:    m.Host,
			Service: m.Service,
			Scope:   m.Scope,
			Name:    m.Name,
			Type:    m.Type,
			Count:   m.Count,
			First:   m.First,
			Last:    m.Last,
		})
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func getOTelMetric(c echo.Context) error {
	id := c.Param("id")
	m := datastore.FindOTelMetricByID(id)
	if m == nil {
		return echo.ErrNotFound
	}
	return c.JSON(http.StatusOK, m)
}

type OTelTraceEnt struct {
	Bucket   string  `json:"Bucket"`
	TraceID  string  `json:"TraceID"`
	Hosts    string  `json:"Hosts"`
	Services string  `json:"Services"`
	Scopes   string  `json:"Scopes"`
	Start    int64   `json:"Start"`
	End      int64   `json:"End"`
	Dur      float64 `json:"Dur"`
	NumSpan  int     `json:"NumSpan"`
}

func getOTelTraceBucketList(c echo.Context) error {
	return c.JSON(http.StatusOK, datastore.GetOTelTraceBucketList())
}

func postOTelTraces(c echo.Context) error {
	req := new([]string)
	err := c.Bind(req)
	if err != nil {
		log.Printf("postOTelTraces err=%v", err)
		return echo.ErrBadRequest
	}
	ret := []*OTelTraceEnt{}
	for _, b := range *req {
		datastore.ForEachOTelTrace(b, func(t *datastore.OTelTraceEnt) bool {
			hosts := []string{}
			services := []string{}
			scopes := []string{}
			hostMap := make(map[string]bool)
			serviceMap := make(map[string]bool)
			scopeMap := make(map[string]bool)
			for _, span := range t.Spans {
				if _, ok := hostMap[span.Host]; !ok {
					hostMap[span.Host] = true
					hosts = append(hosts, span.Host)
				}
				if _, ok := serviceMap[span.Service]; !ok {
					serviceMap[span.Service] = true
					services = append(services, span.Service)
				}
				if _, ok := scopeMap[span.Scope]; !ok {
					scopeMap[span.Scope] = true
					scopes = append(scopes, span.Scope)
				}
			}
			ret = append(ret, &OTelTraceEnt{
				Bucket:   b,
				TraceID:  t.TraceID,
				Hosts:    strings.Join(hosts, " "),
				Services: strings.Join(services, " "),
				Scopes:   strings.Join(scopes, " "),
				Start:    t.Start,
				End:      t.End,
				Dur:      t.Dur,
				NumSpan:  len(t.Spans),
			})
			return len(ret) < 100000
		})
	}
	return c.JSON(http.StatusOK, ret)
}

type OTelTraceReq struct {
	Bucket  string `json:"Bucket"`
	TraceID string `json:"TraceID"`
}

func postOTelTrace(c echo.Context) error {
	req := new(OTelTraceReq)
	err := c.Bind(req)
	if err != nil {
		log.Printf("postOTelTrace err=%v", err)
		return echo.ErrBadRequest
	}
	t := datastore.GetOTelTrace(req.Bucket, req.TraceID)
	if t == nil {
		log.Printf("req=%+v", req)
		return echo.ErrNotFound
	}
	return c.JSON(http.StatusOK, t)
}

func getOTelLastLog(c echo.Context) error {
	ret := []*syslogWebAPILogEnt{}
	i := 0
	end := time.Now().Add(time.Second * 30).Unix()
	st := time.Now().Add(time.Hour * time.Duration(datastore.MapConf.OTelRetention) * -1).UnixNano()
	datastore.ForEachLastLog("syslog", func(l *datastore.LogEnt) bool {
		if l.Time < st {
			return false
		}
		if i > 1000 {
			if time.Now().Unix() > end {
				return false
			}
			i = 0
		}
		i++
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
		if re.Tag != "otel" {
			return true
		}
		re.Time = l.Time
		re.Level = getLevelFromSeverity(int(sv))
		re.Type = getSyslogType(int(sv), int(fac))
		re.Facility = int(fac)
		re.Severity = int(sv)
		ret = append(ret, re)
		return len(ret) < 50000
	})
	return c.JSON(http.StatusOK, ret)
}

func deleteOTelAllData(c echo.Context) error {
	datastore.DeleteAllOTelData()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "OpenTelemetryのデータを全て削除しました。",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type OTelTraceDAGNodeEnt struct {
	Name  string `json:"Name"`
	Count int    `json:"Count"`
}

type OTelTraceDAGLinkEnt struct {
	Src   string `json:"Src"`
	Dst   string `json:"Dst"`
	Count int    `json:"Count"`
}

type OTelTraceDAGEnt struct {
	Nodes []OTelTraceDAGNodeEnt `json:"Nodes"`
	Links []OTelTraceDAGLinkEnt `json:"Links"`
}

func postOTelDAG(c echo.Context) error {
	req := new([]string)
	err := c.Bind(req)
	if err != nil {
		log.Printf("postOTelDAG err=%v", err)
		return echo.ErrBadRequest
	}
	ret := OTelTraceDAGEnt{
		Nodes: []OTelTraceDAGNodeEnt{},
		Links: []OTelTraceDAGLinkEnt{},
	}
	spanMap := make(map[string]string)
	nodeMap := make(map[string]int)
	spanLinkMap := make(map[string]int)

	for _, b := range *req {
		datastore.ForEachOTelTrace(b, func(t *datastore.OTelTraceEnt) bool {
			for _, span := range t.Spans {
				sk := fmt.Sprintf("%s:%s", t.TraceID, span.SpanID)
				spanMap[sk] = span.Service
				nodeMap[span.Service]++
				if span.ParentSpanID != "" {
					lk := fmt.Sprintf("%s:%s\t%s:%s", t.TraceID, span.ParentSpanID, t.TraceID, span.SpanID)
					spanLinkMap[lk]++
				}
			}
			return true
		})
	}
	linkMap := make(map[string]int)
	for k, c := range spanLinkMap {
		a := strings.SplitN(k, "\t", 2)
		if len(a) != 2 {
			continue
		}
		if src, ok := spanMap[a[0]]; ok {
			if dst, ok := spanMap[a[1]]; ok {
				if src != dst {
					linkMap[fmt.Sprintf("%s\t%s", src, dst)] += c
				}
			}
		}
	}
	for n, c := range nodeMap {
		ret.Nodes = append(ret.Nodes, OTelTraceDAGNodeEnt{
			Name:  n,
			Count: c,
		})
	}

	for l, c := range linkMap {
		if a := strings.SplitN(l, "\t", 2); len(a) == 2 {
			ret.Links = append(ret.Links, OTelTraceDAGLinkEnt{
				Src:   a[0],
				Dst:   a[1],
				Count: c,
			})
		}
	}
	return c.JSON(http.StatusOK, ret)
}
