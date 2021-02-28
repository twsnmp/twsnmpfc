package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getDevices(c echo.Context) error {
	r := []*datastore.DeviceEnt{}
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		r = append(r, d)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteDevice(c echo.Context) error {
	id := c.Param("id")
	datastore.DeleteDevice(id)
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetDevices(c echo.Context) error {
	report.ResetDeviceScore()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
