package webapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/backend"
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
		return echo.ErrBadRequest
	}
	if err := datastore.DeletePollings(ids); err != nil {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "user",
			Level: "warn",
			Event: fmt.Sprintf("%d件のポーリングを削除できません", len(ids)),
		})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("%d件のポーリングを削除しました", len(ids)),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func setPollingLevel(c echo.Context) error {
	var params = struct {
		Level string
		IDs   []string
	}{}
	if err := c.Bind(&params); err != nil {
		return echo.ErrBadRequest
	}
	for _, id := range params.IDs {
		p := datastore.GetPolling(id)
		if p != nil {
			p.Level = params.Level
			p.State = "unknown"
			p.NextTime = 0
			if err := datastore.UpdatePolling(p); err != nil {
				return echo.ErrBadRequest
			}
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("%d件のポーリングのレベルを%sに変更しました", len(params.IDs), params.Level),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func setPollingLogMode(c echo.Context) error {
	var params = struct {
		LogMode int
		IDs     []string
	}{}
	if err := c.Bind(&params); err != nil {
		return echo.ErrBadRequest
	}
	for _, id := range params.IDs {
		p := datastore.GetPolling(id)
		if p != nil {
			p.LogMode = params.LogMode
			if err := datastore.UpdatePolling(p); err != nil {
				return echo.ErrBadRequest
			}
		}
	}
	modeName := "しない"
	switch params.LogMode {
	case datastore.LogModeAlways:
		modeName = "常時記録"
	case datastore.LogModeOnChange:
		modeName = "状態変化時のみ記録"
	case datastore.LogModeAI:
		modeName = "AI分析"
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("%d件のポーリングのログモードを%sに変更しました", len(params.IDs), modeName),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func setPollingParams(c echo.Context) error {
	var params = struct {
		Timeout int
		Retry   int
		PollInt int
		IDs     []string
	}{}
	if err := c.Bind(&params); err != nil {
		return echo.ErrBadRequest
	}
	for _, id := range params.IDs {
		p := datastore.GetPolling(id)
		if p != nil {
			p.Timeout = params.Timeout
			p.PollInt = params.PollInt
			p.Retry = params.Retry
			if err := datastore.UpdatePolling(p); err != nil {
				return echo.ErrBadRequest
			}
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("%d件のポーリングのパラメータを変更しました", len(params.IDs)),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getPollingCheck(c echo.Context) error {
	id := c.Param("id")
	all := id == "all"
	if all {
		polling.CheckAllPoll()
	} else {
		polling.PollNowNode(id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postPollingUpdate(c echo.Context) error {
	pu := new(datastore.PollingEnt)
	if err := c.Bind(pu); err != nil {
		return echo.ErrBadRequest
	}
	// ここで入力データのチェックをする
	p := datastore.GetPolling(pu.ID)
	if p == nil {
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
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postPollingAdd(c echo.Context) error {
	p := new(datastore.PollingEnt)
	if err := c.Bind(p); err != nil {
		return echo.ErrBadRequest
	}
	// ここで入力データのチェックをする
	p.NextTime = 0
	p.State = "unknown"
	if err := datastore.AddPolling(p); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type pollingWebAPI struct {
	Node    *datastore.NodeEnt
	Polling *datastore.PollingEnt
}

type timeFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
}

func getPolling(c echo.Context) error {
	id := c.Param("id")
	r := pollingWebAPI{}
	r.Polling = datastore.GetPolling(id)
	if r.Polling == nil {
		return echo.ErrBadRequest
	}
	r.Node = datastore.GetNode(r.Polling.NodeID)
	if r.Node == nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, r)
}

func postPollingLogs(c echo.Context) error {
	id := c.Param("id")
	r := []*datastore.PollingLogEnt{}
	polling := datastore.GetPolling(id)
	if polling == nil {
		return echo.ErrBadRequest
	}
	filter := new(timeFilter)
	if err := c.Bind(filter); err != nil {
		return echo.ErrBadRequest
	}
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 24*7)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	i := 0
	datastore.ForEachPollingLog(st, et, id, func(l *datastore.PollingLogEnt) bool {
		r = append(r, l)
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
		return echo.ErrBadRequest
	}
	n := datastore.GetNode(paa.NodeID)
	if n == nil {
		return echo.ErrBadRequest
	}
	for _, id := range paa.PollingTemplateIDs {
		pt := datastore.GetPollingTemplate(id)
		if pt == nil {
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
		p.Retry = datastore.MapConf.Retry
		p.LogMode = 0
		p.NextTime = 0
		p.State = "unknown"
		if err := datastore.AddPolling(p); err != nil {
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getPollingAIData(c echo.Context) error {
	id := c.Param("id")
	r := backend.AIReq{
		PollingID: id,
	}
	if err := backend.MakeAIData(&r); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, r)
}

func getPollingLogTimeAnalyze(c echo.Context) error {
	id := c.Param("id")
	r, err := backend.TimeAnalyzePollingLog(id)
	if err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, r)
}

func deletePollingLog(c echo.Context) error {
	id := c.Param("id")
	p := datastore.GetPolling(id)
	if p == nil {
		return echo.ErrBadRequest
	}
	p.NextTime = 0
	p.State = "unknown"
	p.Result = make(map[string]interface{})
	datastore.UpdatePolling(p)
	if err := datastore.ClearPollingLog(id); err != nil {
		return echo.ErrInternalServerError
	}
	datastore.DeleteAIResult(id)
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
