package webapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type aiWebAPI struct {
	ID          string
	NodeID      string
	NodeName    string
	PollingName string
	AIResult    *datastore.AIResult
}

type aiListEntWebAPI struct {
	ID          string
	NodeID      string
	NodeName    string
	PollingName string
	Score       float64
	Count       int
	LastTime    int64
}

func getAIList(c echo.Context) error {
	r := []*aiListEntWebAPI{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.LogMode != datastore.LogModeAI {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		air, err := datastore.GetAIReesult(p.ID)
		if err != nil || len(air.ScoreData) < 1 {
			return true
		}
		e := &aiListEntWebAPI{
			ID:          p.ID,
			NodeName:    n.Name,
			PollingName: p.Name,
			Score:       air.ScoreData[len(air.ScoreData)-1][1],
			Count:       len(air.ScoreData),
			LastTime:    air.LastTime,
		}
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func getAIResult(c echo.Context) error {
	id := c.Param("id")
	p := datastore.GetPolling(id)
	if p == nil {
		return echo.ErrBadRequest
	}
	n := datastore.GetNode(p.NodeID)
	if n == nil {
		return echo.ErrBadRequest
	}
	air, err := datastore.GetAIReesult(p.ID)
	if err != nil || len(air.ScoreData) < 1 {
		return echo.ErrNotFound
	}
	r := aiWebAPI{
		ID:          p.ID,
		NodeID:      p.NodeID,
		NodeName:    n.Name,
		PollingName: p.Name,
		AIResult:    air,
	}
	return c.JSON(http.StatusOK, r)
}

func deleteAIResult(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go func() {
			datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
				if err := backend.DeleteAIResult(p.ID); err != nil {
					log.Printf("delete ai result err=%v", err)
				}
				return true
			})
		}()
	} else {
		if err := backend.DeleteAIResult(id); err != nil {
			log.Printf("delete ai result err=%v", err)
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("AI分析結果を削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
