package webapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getDevices(c echo.Context) error {
	r := []*datastore.DeviceEnt{}
	datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
		if d.ValidScore {
			r = append(r, d)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteDevice(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("devices")
	} else {
		datastore.DeleteReport("devices", []string{id})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("デバイスを削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetDevices(c echo.Context) error {
	report.ResetDevicesScore()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "デバイスレポートの信用スコアを再計算しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
