package webapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func deleteNetworks(c echo.Context) error {
	ids := []string{}
	if err := c.Bind(&ids); err != nil {
		return echo.ErrBadRequest
	}
	for _, id := range ids {
		if err := datastore.DeleteNetwok(id); err != nil {
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNetwork(c echo.Context) error {
	nu := new(datastore.NetworkEnt)
	if err := c.Bind(nu); err != nil {
		log.Printf("post network err=%v", err)
		return echo.ErrBadRequest
	}
	if n := datastore.GetNetwork(nu.ID); n == nil {
		if err := datastore.AddNetwork(nu); err != nil {
			log.Printf("post network err=%v", err)
			return echo.ErrBadRequest
		}
	} else {
		if err := datastore.UpdateNetwork(nu); err != nil {
			log.Printf("post network err=%v", err)
			return echo.ErrBadRequest
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: nu.Name,
		Event:    fmt.Sprintf("ネットワークを更新しました(%s)", nu.ID),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
