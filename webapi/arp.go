package webapi

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type arpFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	IP        string
	MAC       string
}

type arpWebAPI struct {
	Time      int64
	State     string
	IP        string
	MAC       string
	Vendor    string
	OldMAC    string
	OldVendor string
}

func postArp(c echo.Context) error {
	r := []*arpWebAPI{}
	var err error
	filter := new(arpFilter)
	if err = c.Bind(filter); err != nil {
		return echo.ErrBadRequest
	}
	ipFilter := makeStringFilter(filter.IP)
	macFilter := makeStringFilter(filter.MAC)
	st := makeStartTimeFilter(filter.StartDate, filter.StartTime)
	et := makeEndTimeFilter(filter.EndDate, filter.EndTime)
	datastore.ForEachLogReverse(st, et, "arplog", func(l *datastore.LogEnt) bool {
		a := strings.Split(l.Log, ",")
		if len(a) < 3 {
			return true
		}
		re := arpWebAPI{
			Time:   l.Time,
			State:  a[0],
			IP:     a[1],
			MAC:    a[2],
			Vendor: datastore.FindVendor(a[2]),
		}
		if len(a) > 3 {
			re.OldMAC = a[3]
			re.OldVendor = datastore.FindVendor(a[3])
		}
		re.Time = l.Time
		if ipFilter != nil && !ipFilter.Match([]byte(a[1])) {
			return true
		}
		if macFilter != nil && !macFilter.Match([]byte(a[2])) && !macFilter.Match([]byte(a[3])) {
			return true
		}
		r = append(r, &re)
		return len(r) <= datastore.MapConf.LogDispSize
	})
	// 逆順にする
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return c.JSON(http.StatusOK, r)
}
