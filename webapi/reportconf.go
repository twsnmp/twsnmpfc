package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getReportConf(c echo.Context) error {
	return c.JSON(http.StatusOK, datastore.ReportConf)
}

func postReportConf(c echo.Context) error {
	nc := new(datastore.ReportConfEnt)
	if err := c.Bind(nc); err != nil {
		return echo.ErrBadRequest
	}
	datastore.ReportConf = *nc
	if err := datastore.SaveReportConf(); err != nil {
		return echo.ErrBadRequest
	}
	report.UpdateReportConf()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "レポート設定を更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
