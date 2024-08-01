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
	r := new(discoverWebAPI)
	r.Conf = datastore.DiscoverConf
	r.Stat = discover.Stat
	return c.JSON(http.StatusOK, r)
}

func getDiscoverIPRange(c echo.Context) error {
	return c.JSON(http.StatusOK, discover.GetDiscoverIPRange())
}

func postDiscoverStart(c echo.Context) error {
	dc := new(datastore.DiscoverConfEnt)
	if err := c.Bind(dc); err != nil {
		log.Printf("start discover err=%v", err)
		return echo.ErrBadRequest
	}
	if err := c.Validate(dc); err != nil {
		log.Printf("start discover err=%v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if discover.Stat.Running {
		log.Printf("discover already start")
		return echo.ErrBadRequest
	}
	datastore.DiscoverConf = *dc
	if err := datastore.SaveDiscoverConf(); err != nil {
		log.Printf("start discover err=%v", err)
		return echo.ErrBadRequest
	}
	if err := discover.StartDiscover(); err != nil {
		log.Printf("start discover err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postDiscoverStop(c echo.Context) error {
	discover.StopDiscover()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteDiscoverStat(c echo.Context) error {
	discover.ClearStat()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
