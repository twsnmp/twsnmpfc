package webapi

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/discover"
)

type discoverWebAPI struct {
	Conf datastore.DiscoverConfEnt
	Stat discover.DiscoverStat
}

func getDiscover(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := new(discoverWebAPI)
	r.Conf = api.DataStore.DiscoverConf
	r.Stat = api.Discover.Stat
	return c.JSON(http.StatusOK, r)
}

func postDiscoverStart(c echo.Context) error {
	dc := new(datastore.DiscoverConfEnt)
	if err := c.Bind(dc); err != nil {
		log.Printf("postDiscoverStart err=%v", err)
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	api.DataStore.DiscoverConf = *dc
	if err := api.DataStore.SaveDiscoverConfToDB(); err != nil {
		log.Printf("postDiscoverStart err=%v", err)
		return echo.ErrBadRequest
	}
	if err := api.Discover.StartDiscover(); err != nil {
		log.Printf("postDiscoverStart err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postDiscoverStop(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	api.Discover.StopDiscover()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
