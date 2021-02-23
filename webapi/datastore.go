package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type dataStoreWebAPI struct {
	Backup     *datastore.DBBackupEnt
	DBStats    *datastore.DBStatsEnt
	DBStatsLog *[]datastore.DBStatsLogEnt
}

func getDataStore(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := new(dataStoreWebAPI)
	r.Backup = &api.DataStore.Backup
	r.DBStats = &api.DataStore.DBStats
	r.DBStatsLog = &api.DataStore.DBStatsLog
	return c.JSON(http.StatusOK, r)
}

func postBackup(c echo.Context) error {
	bc := new(datastore.DBBackupEnt)
	if err := c.Bind(bc); err != nil {
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	api.DataStore.Backup.Mode = bc.Mode
	api.DataStore.Backup.ConfigOnly = bc.ConfigOnly
	api.DataStore.Backup.Generation = bc.Generation
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteLogs(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	go api.DataStore.DeleteAllLogs()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteReport(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	go api.DataStore.ClearAllReport()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteAIResult(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	go func() {
		api.DataStore.ForEachPollings(func(p *datastore.PollingEnt) bool {
			api.DataStore.DeleteAIResult(p.ID)
			return true
		})
	}()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
