package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getSensors(c echo.Context) error {
	r := []*datastore.SensorEnt{}
	datastore.ForEachSensors(func(s *datastore.SensorEnt) bool {
		r = append(r, s)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteSensor(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("sensor")
	} else {
		datastore.DeleteReport("sensor", []string{id})
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

// センサーをモニタするしないを切り替え
func postSensor(c echo.Context) error {
	id := c.Param("id")
	s := datastore.GetSensor(id)
	if s == nil {
		return echo.ErrBadRequest
	}
	s.Ignore = !s.Ignore
	s.State = "off"
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
