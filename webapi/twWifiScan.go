package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getWifiAP(c echo.Context) error {
	r := []*datastore.WifiAPEnt{}
	datastore.ForEachWifiAP(func(e *datastore.WifiAPEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteWifiAP(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("wifiAP")
	} else {
		datastore.DeleteReport("wifiAP", []string{id})
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
