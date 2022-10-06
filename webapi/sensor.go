package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type SensorEnt struct {
	ID          string
	Host        string
	Type        string
	Param       string
	Total       int64
	Send        int64
	State       string
	Ignore      bool
	StatsLen    int
	MonitorsLen int
	FirstTime   int64
	LastTime    int64
}

func getSensors(c echo.Context) error {
	r := []SensorEnt{}
	datastore.ForEachSensors(func(s *datastore.SensorEnt) bool {
		r = append(r, SensorEnt{
			ID:          s.ID,
			Host:        s.Host,
			Type:        s.Type,
			Param:       s.Param,
			Total:       s.Total,
			Send:        s.Send,
			State:       s.State,
			Ignore:      s.Ignore,
			StatsLen:    len(s.Stats),
			MonitorsLen: len(s.Monitors),
			FirstTime:   s.FirstTime,
			LastTime:    s.LastTime,
		})
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func getSensorStats(c echo.Context) error {
	id := c.Param("id")
	s := datastore.GetSensor(id)
	if s == nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, s.Stats)
}

func getSensorMonitors(c echo.Context) error {
	id := c.Param("id")
	s := datastore.GetSensor(id)
	if s == nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, s.Monitors)
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
