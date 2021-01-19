package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getNodes(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := []*datastore.NodeEnt{}
	api.DataStore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r = append(r, n)
		return true
	})
	return c.JSON(http.StatusOK, r)
}
