package webapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getServers(c echo.Context) error {
	r := []*datastore.ServerEnt{}
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		if s.ValidScore {
			r = append(r, s)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteServer(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("servers")
	} else {
		datastore.DeleteReport("servers", id)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("サーバーを削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetServers(c echo.Context) error {
	report.ResetServersScore()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "サーバーレポートの信用スコアを再計算しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
