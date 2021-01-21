package webapi

import (
	"log"
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

func postNodeDelete(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	id := new(idWebAPI)
	if err := c.Bind(id); err != nil {
		log.Printf("postNodeDelete err=%v", err)
		return echo.ErrBadRequest
	}
	if err := api.DataStore.DeleteNode(id.ID); err != nil {
		log.Printf("postNodeDelete err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNodeUpdate(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	nu := new(datastore.NodeEnt)
	if err := c.Bind(nu); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	n := api.DataStore.GetNode(nu.ID)
	if n == nil {
		log.Printf("postNodeUpdate Node not found ID=%s", nu.ID)
		return echo.ErrBadRequest
	}
	n.Name = nu.Name
	n.Descr = nu.Descr
	n.IP = nu.IP
	log.Printf("node=%#v", n)
	if err := api.DataStore.UpdateNode(n); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
