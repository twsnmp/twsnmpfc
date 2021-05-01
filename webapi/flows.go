package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getFlows(c echo.Context) error {
	r := []*datastore.FlowEnt{}
	datastore.ForEachFlows(func(f *datastore.FlowEnt) bool {
		if f.ValidScore {
			r = append(r, f)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteFlow(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("flows")
	} else {
		datastore.DeleteFlow(id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetFlows(c echo.Context) error {
	report.ResetFlowsScore()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
