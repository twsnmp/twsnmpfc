package webapi

import (
	"fmt"
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
		datastore.DeleteReport("flows", []string{id})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("フローを削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetFlows(c echo.Context) error {
	report.ResetFlowsScore()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "フローレポートの信用スコアを再計算しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type unknownPortEnt struct {
	Name  string
	Count int64
}

func getUnknownPortList(c echo.Context) error {
	r := []unknownPortEnt{}
	for k, v := range report.UnKnownPortMap {
		r = append(r, unknownPortEnt{Name: k, Count: v})
	}
	return c.JSON(http.StatusOK, r)
}

func getFumbleFlows(c echo.Context) error {
	r := []*datastore.FumbleEnt{}
	datastore.ForEachFumbleFlows(func(f *datastore.FumbleEnt) bool {
		r = append(r, f)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteFumbleFlow(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("fumbleFlows")
	} else {
		datastore.DeleteReport("fumbleFlows", []string{id})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("Fumbleフローを削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
