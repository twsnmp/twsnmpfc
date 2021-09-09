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
	st := makeTimeFilter(filter.StartDate, filter.StartTime, 24)
	et := makeTimeFilter(filter.EndDate, filter.EndTime, 0)
	i := 0
	datastore.ForEachLog(st, et, "arplog", func(l *datastore.LogEnt) bool {
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
		i++
		return i <= datastore.MapConf.LogDispSize
	})
	return c.JSON(http.StatusOK, r)
}
