package webapi

import (
	"fmt"
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
		datastore.DeleteReport("ips", []string{ip})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("IPレポートを削除しました(%s)", ip),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetIPReport(c echo.Context) error {
	report.ResetIPReportScore()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "IPレポートの信用スコアを再計算しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
