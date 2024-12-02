package webapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func deleteNetwork(c echo.Context) error {
	id := c.Param("id")
	n := datastore.GetNetwork(id)
	if n == nil {
		return echo.ErrNotFound
	}
	name := n.Name
	if err := datastore.DeleteNetwok(id); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: name,
		Event:    "ネットワークを削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNetwork(c echo.Context) error {
	nu := new(datastore.NetworkEnt)
	if err := c.Bind(nu); err != nil {
		log.Printf("post network err=%v", err)
		return echo.ErrBadRequest
	}
	op := "更新"
	if n := datastore.GetNetwork(nu.ID); n == nil {
		if err := datastore.AddNetwork(nu); err != nil {
			log.Printf("post network err=%v", err)
			return echo.ErrBadRequest
		}
		op = "追加"
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
		Event:    fmt.Sprintf("ネットワークを%sしました", op),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNetworkPos(c echo.Context) error {
	var pos itemPosWebAPI
	if err := c.Bind(&pos); err != nil {
		return echo.ErrBadRequest
	}
	n := datastore.GetNetwork(pos.ID)
	if n == nil {
		return echo.ErrBadRequest
	}
	n.X = pos.X
	n.Y = pos.Y
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getFindNeighborNetworksAndLines(c echo.Context) error {
	id := c.Param("id")
	n := datastore.GetNetwork(id)
	if n == nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, backend.FindNeighborNetworksAndLines(n))
}

func getCheckNetwork(c echo.Context) error {
	id := c.Param("id")
	backend.CheckNetwork(id)
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
