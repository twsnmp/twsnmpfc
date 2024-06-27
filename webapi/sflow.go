package webapi

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type sFlowCounterFilter struct {
	Remote    string
	Type      string
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	NextTime  int64
	Filter    int
}

type sFlowCounterWebAPI struct {
	Logs     []*sFlowCounterWebAPILogEnt
	NextTime int64
	Process  int
	Filter   int
	Limit    int
}

type sFlowCounterWebAPILogEnt struct {
	Time int64
	datastore.SFlowCounterEnt
}

func postSFlowCounter(c echo.Context) error {
	r := new(sFlowCounterWebAPI)
	filter := new(sFlowCounterFilter)
	if err := c.Bind(filter); err != nil {
		return echo.ErrBadRequest
	}
	remoteFilter := makeStringFilter(filter.Remote)
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 24)
	if filter.NextTime > 0 {
		st = filter.NextTime
	}
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	r.NextTime = 0
	r.Process = 0
	r.Filter = filter.Filter
	i := 0
	to := 15
	if datastore.MapConf.LogTimeout > 0 {
		to = datastore.MapConf.LogTimeout
	}
	end := time.Now().Unix() + int64(to)
	datastore.ForEachLog(st, et, "sflowCounter", func(l *datastore.LogEnt) bool {
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
		var re sFlowCounterWebAPILogEnt
		if err := json.Unmarshal([]byte(l.Log), &re); err != nil {
			log.Println(err)
			return true
		}
		re.Time = l.Time
		if remoteFilter != nil && !remoteFilter.Match([]byte(re.Remote)) {
			return true
		}
		if filter.Type != "" && filter.Type != re.Type {
			return true
		}
		r.Logs = append(r.Logs, &re)
		r.Filter++
		return true
	})
	r.Limit = datastore.MapConf.LogDispSize
	return c.JSON(http.StatusOK, r)
}
