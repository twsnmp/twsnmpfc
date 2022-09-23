package webapi

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getSdrPowerKeys(c echo.Context) error {
	return c.JSON(http.StatusOK, datastore.GetSdrPowerKeys())
}

func getSdrPowerData(c echo.Context) error {
	ids := []string{}
	if err := c.Bind(&ids); err != nil {
		return echo.ErrBadRequest
	}
	r := [][]*datastore.SdrPowerEnt{}
	for _, id := range ids {
		a := strings.Split(id, ":")
		if len(a) != 2 {
			continue
		}
		if t, err := strconv.ParseInt(a[1], 10, 64); err == nil {
			l := []*datastore.SdrPowerEnt{}
			datastore.ForEachSdrPower(t, a[0], func(e *datastore.SdrPowerEnt) bool {
				l = append(l, e)
				return true
			})
			if len(l) > 0 {
				log.Printf("sdrdata len=%d", len(l))
				r = append(r, l)
			}
		}
	}
	return c.JSON(http.StatusOK, r)
}

func deleteSdrPower(c echo.Context) error {
	ids := []string{}
	if err := c.Bind(&ids); err != nil {
		return echo.ErrBadRequest
	}
	for _, id := range ids {
		a := strings.Split(id, ":")
		if len(a) != 2 {
			continue
		}
		if t, err := strconv.ParseInt(a[1], 10, 64); err == nil {
			datastore.DeleteSdrPower(t, a[0])
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
