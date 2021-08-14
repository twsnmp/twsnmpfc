package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getWinEventID(c echo.Context) error {
	r := []*datastore.WinEventIDEnt{}
	datastore.ForEachWinEventID(func(s *datastore.WinEventIDEnt) bool {
		r = append(r, s)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWinEventID(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("WinEventID")
	} else {
		datastore.DeleteReport("WinEventID", id)
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
