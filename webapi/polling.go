package webapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/polling"
)

func copyPolling(to, from string) {
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == from {
			pn := *p
			pn.ID = ""
			pn.NodeID = to
			pn.NextTime = 0
			pn.State = "unknown"
			pn.Result = make(map[string]interface{})
			datastore.AddPolling(&pn)
		}
		return true
	})
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

func deletePollings(c echo.Context) error {
	ids := []string{}
	if err := c.Bind(&ids); err != nil {
		log.Printf("deletePolling err=%v", err)
		return echo.ErrBadRequest
	}
	for _, id := range ids {
		if err := datastore.DeletePolling(id); err != nil {
			log.Printf("deletePolling err=%v", err)
			return echo.ErrBadRequest
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("%dのポーリングを削除しました", len(ids)),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func setPollingLevel(c echo.Context) error {
	var params = struct {
		Level string
		IDs   []string
	}{}
	if err := c.Bind(&params); err != nil {
		log.Printf("setPollingLevel err=%v", err)
		return echo.ErrBadRequest
	}
	for _, id := range params.IDs {
		p := datastore.GetPolling(id)
		if p != nil {
			p.Level = params.Level
			p.State = "unknown"
			p.NextTime = 0
			if err := datastore.UpdatePolling(p); err != nil {
				log.Printf("setPollingLevel err=%v", err)
				return echo.ErrBadRequest
			}
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("%dのポーリングのレベルを%sに変更しました", len(params.IDs), params.Level),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getPollingCheck(c echo.Context) error {
	id := c.Param("id")
	all := id == "all"
	nodeName := "全て"
	if !all {
		n := datastore.GetNode(id)
		if n == nil {
			log.Printf("getPollingCheck node not fond id=%s", id)
			return echo.ErrBadRequest
		}
		nodeName = n.Name
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   id,
		NodeName: nodeName,
		Event:    "ポーリングの再確認",
	})
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if (all || p.NodeID == id) && p.State != "normal" {
			p.NextTime = 0
			p.State = "unknown"
			datastore.SetNodeStateChanged(p.NodeID)
		}
		return true
	})
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
	p.Params = pu.Params
	p.Mode = pu.Mode
	p.Script = pu.Script
	p.Extractor = pu.Extractor
	p.Filter = pu.Filter
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

func getPollingTemplate(c echo.Context) error {
	r := []*datastore.PollingTemplateEnt{}
	datastore.ForEachPollingTemplate(func(pt *datastore.PollingTemplateEnt) bool {
		r = append(r, pt)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

type pollingAutoAddEnt struct {
	NodeID             string
	PollingTemplateIDs []string
}

func postPollingAutoAdd(c echo.Context) error {
	paa := new(pollingAutoAddEnt)
	if err := c.Bind(paa); err != nil {
		log.Printf("postPollingAutoAdd err=%v", err)
		return echo.ErrBadRequest
	}
	n := datastore.GetNode(paa.NodeID)
	if n == nil {
		log.Printf("node not found id=%s", paa.NodeID)
		return echo.ErrBadRequest
	}
	for _, id := range paa.PollingTemplateIDs {
		pt := datastore.GetPollingTemplate(id)
		if pt == nil {
			log.Printf("template not found id=%s", id)
			continue
		}
		if pt.AutoMode == "disable" {
			continue
		}
		if pt.AutoMode != "" {
			// インデックスの展開などを行う並列で処理する
			go polling.AutoAddPolling(n, pt)
			continue
		}
		p := new(datastore.PollingEnt)
		p.Name = pt.Name
		p.NodeID = n.ID
		p.Type = pt.Type
		p.Params = pt.Params
		p.Mode = pt.Mode
		p.Script = pt.Script
		p.Extractor = pt.Extractor
		p.Filter = pt.Filter
		p.Level = pt.Level
		p.PollInt = datastore.MapConf.PollInt
		p.Timeout = datastore.MapConf.Timeout
		p.Retry = datastore.MapConf.Timeout
		p.LogMode = 0
		p.NextTime = 0
		p.State = "unknown"
		if err := datastore.AddPolling(p); err != nil {
			log.Printf("postPollingAutoAdd err=%v", err)
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
