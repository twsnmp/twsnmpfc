package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func getBlueDevice(c echo.Context) error {
	r := []*datastore.BlueDeviceEnt{}
	datastore.ForEachBludeDevice(func(e *datastore.BlueDeviceEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteBlueDevice(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("blueDevice")
	} else {
		datastore.DeleteReport("blueDevice", []string{id})
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getEnvMonitor(c echo.Context) error {
	r := []*datastore.EnvMonitorEnt{}
	datastore.ForEachEnvMonitor(func(e *datastore.EnvMonitorEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteEnvMonitor(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("envMonitor")
	} else {
		datastore.DeleteReport("envMonitor", []string{id})
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getPowerMonitor(c echo.Context) error {
	r := []*datastore.PowerMonitorEnt{}
	datastore.ForEachPowerMonitor(func(e *datastore.PowerMonitorEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deletePowerMonitor(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("powerMonitor")
	} else {
		datastore.DeleteReport("powerMonitor", []string{id})
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getMotionSensor(c echo.Context) error {
	r := []*datastore.MotionSensorEnt{}
	datastore.ForEachMotionSensor(func(e *datastore.MotionSensorEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteMotionSensor(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("motionSensor")
	} else {
		datastore.DeleteReport("motionSensor", []string{id})
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
