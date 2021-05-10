package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getIPReport(c echo.Context) error {
	r := []*datastore.IPReportEnt{}
	datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
		if i.ValidScore {
			r = append(r, i)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteIPReport(c echo.Context) error {
	ip := c.Param("ip")
	if ip == "all" {
		go datastore.ClearReport("ips")
	} else {
		datastore.DeleteReport("ips", ip)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetIPReport(c echo.Context) error {
	report.ResetIPReportScore()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
