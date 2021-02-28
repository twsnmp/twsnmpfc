package webapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/logger"
)

type dataStoreWebAPI struct {
	Backup     *datastore.DBBackupEnt
	DBStats    *datastore.DBStatsEnt
	DBStatsLog *[]datastore.DBStatsLogEnt
}

func getDataStore(c echo.Context) error {
	r := new(dataStoreWebAPI)
	r.Backup = &datastore.Backup
	r.DBStats = &datastore.DBStats
	r.DBStatsLog = &datastore.DBStatsLog
	return c.JSON(http.StatusOK, r)
}

func postBackup(c echo.Context) error {
	bc := new(datastore.DBBackupEnt)
	if err := c.Bind(bc); err != nil {
		return echo.ErrBadRequest
	}
	datastore.Backup.Mode = bc.Mode
	datastore.Backup.ConfigOnly = bc.ConfigOnly
	datastore.Backup.Generation = bc.Generation
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteLogs(c echo.Context) error {
	go datastore.DeleteAllLogs()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteArp(c echo.Context) error {
	logger.ResetArpTable()
	datastore.DeleteArp()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteReport(c echo.Context) error {
	go datastore.ClearAllReport()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteAIResult(c echo.Context) error {
	go func() {
		datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
			datastore.DeleteAIResult(p.ID)
			return true
		})
	}()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
