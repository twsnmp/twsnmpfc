package webapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getMqttStat(c echo.Context) error {
	r := []*datastore.MqttStatEnt{}
	datastore.ForEachMqttStat(func(s *datastore.MqttStatEnt) bool {
		r = append(r, s)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteMqttStat(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		if err := datastore.DeleteAllMqttStat(); err != nil {
			return echo.ErrInternalServerError
		}
	} else {
		datastore.DeleteMqttStats([]string{id})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("MQTT統計を削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
