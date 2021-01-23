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
	// ここで入力チェック
	n := api.DataStore.GetNode(nu.ID)
	if n == nil {
		log.Printf("postNodeUpdate Node not found ID=%s", nu.ID)
		return echo.ErrBadRequest
	}
	n.Name = nu.Name
	n.Descr = nu.Descr
	n.IP = nu.IP
	n.Icon = nu.Icon
	n.SnmpMode = nu.SnmpMode
	n.Community = nu.Community
	n.User = nu.User
	n.Password = nu.Password
	n.PublicKey = nu.PublicKey
	n.URL = nu.URL
	n.Type = nu.Type
	n.AddrMode = nu.AddrMode
	if err := api.DataStore.UpdateNode(n); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type pollingsWebAPI struct {
	Pollings []*datastore.PollingEnt
	NodeList []selectEntWebAPI
}

func getPollings(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := pollingsWebAPI{}
	api.DataStore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r.NodeList = append(r.NodeList, selectEntWebAPI{Text: n.Name, Value: n.ID})
		return true
	})
	api.DataStore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		r.Pollings = append(r.Pollings, p)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func postPollingDelete(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	id := new(idWebAPI)
	if err := c.Bind(id); err != nil {
		log.Printf("postPollingDelete err=%v", err)
		return echo.ErrBadRequest
	}
	if err := api.DataStore.DeletePolling(id.ID); err != nil {
		log.Printf("postPollingDelete err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postPollingUpdate(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	pu := new(datastore.PollingEnt)
	if err := c.Bind(pu); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	// ここで入力データのチェックをする
	p := api.DataStore.GetPolling(pu.ID)
	if p == nil {
		log.Printf("postPollingUpdate Node not found ID=%s", pu.ID)
		return echo.ErrBadRequest
	}
	p.Name = pu.Name
	p.NodeID = pu.NodeID
	p.Type = pu.Type
	p.Polling = pu.Polling
	p.Level = pu.Level
	p.PollInt = pu.PollInt
	p.Timeout = pu.Timeout
	p.Retry = pu.Retry
	p.LogMode = pu.LogMode
	p.NextTime = 0
	p.State = "unknown"
	if err := api.DataStore.UpdatePolling(p); err != nil {
		log.Printf("postNodeUpdate err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
