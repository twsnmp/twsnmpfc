package webapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func deleteLine(c echo.Context) error {
	l := new(datastore.LineEnt)
	if err := c.Bind(l); err != nil {
		log.Printf("delete line err=%v", err)
		return echo.ErrBadRequest
	}
	if err := datastore.DeleteLine(l.ID); err != nil {
		log.Printf("delete line err=%v", err)
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:   "user",
		Level:  "info",
		NodeID: l.NodeID1,
		Event:  fmt.Sprintf("ラインを削除しました(%s)", l.ID),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postLine(c echo.Context) error {
	lu := new(datastore.LineEnt)
	if err := c.Bind(lu); err != nil {
		log.Printf("post line err=%v", err)
		return echo.ErrBadRequest
	}
	lu.State1 = "unknown"
	if lu.PollingID1 != "" {
		if p := datastore.GetPolling(lu.PollingID1); p != nil {
			lu.State1 = p.State
		}
	}
	lu.State2 = lu.State1
	if lu.PollingID2 != "" {
		if p := datastore.GetPolling(lu.PollingID2); p != nil {
			lu.State2 = p.State
		}
	}
	l := datastore.GetLine(lu.ID)
	if l == nil {
		if err := datastore.AddLine(lu); err != nil {
			log.Printf("post line err=%v", err)
			return echo.ErrBadRequest
		}
	} else {
		l.NodeID1 = lu.NodeID1
		l.NodeID2 = lu.NodeID2
		l.PollingID1 = lu.PollingID1
		l.PollingID2 = lu.PollingID2
		l.State1 = lu.State1
		l.State2 = lu.State2
		l.Info = lu.Info
		l.PollingID = lu.PollingID
		l.Width = lu.Width
		l.Port = lu.Port
		if err := datastore.UpdateLine(l); err != nil {
			log.Printf("post line err=%v", err)
			return echo.ErrBadRequest
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:   "user",
		Level:  "info",
		NodeID: lu.NodeID1,
		Event:  fmt.Sprintf("ラインを更新しました(%s)", lu.ID),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
