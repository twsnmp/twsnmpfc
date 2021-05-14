package webapi

import (
	"log"
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
	if err := datastore.SaveBackup(); err != nil {
		log.Printf("postBackup err=%v", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteLogs(c echo.Context) error {
	go datastore.DeleteAllLogs()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "全てのログを削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteArp(c echo.Context) error {
	logger.ResetArpTable()
	datastore.DeleteArp()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "ARP監視のログを削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
