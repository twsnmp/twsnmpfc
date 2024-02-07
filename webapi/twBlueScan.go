package webapi

import (
	"log"
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

type postBlueScanNameEnt struct {
	Type string
	ID   string
	Name string
}

func postBlueScanName(c echo.Context) error {
	req := new(postBlueScanNameEnt)
	if err := c.Bind(req); err != nil {
		return echo.ErrBadRequest
	}
	ok := false
	switch req.Type {
	case "device":
		ok = datastore.SetBlueDeviceName(req.ID, req.Name)
	case "env":
		ok = datastore.SetEnvMonitorName(req.ID, req.Name)
	case "power":
		ok = datastore.SetPowerMonitorName(req.ID, req.Name)
	case "motion":
		ok = datastore.SetMotionSensorName(req.ID, req.Name)
	}
	if ok {
		return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
	}
	log.Printf("failed to set name %+v", req)
	return echo.ErrBadRequest
}
