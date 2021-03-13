package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getServers(c echo.Context) error {
	r := []*datastore.ServerEnt{}
	datastore.ForEachServers(func(s *datastore.ServerEnt) bool {
		r = append(r, s)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteServer(c echo.Context) error {
	id := c.Param("id")
	datastore.DeleteServer(id)
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetServers(c echo.Context) error {
	report.ResetServersScore()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}