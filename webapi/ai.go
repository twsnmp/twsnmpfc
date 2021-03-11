package webapi

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
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
		air, err := datastore.LoadAIReesult(p.ID)
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
		log.Printf("polling not found")
		return echo.ErrBadRequest
	}
	n := datastore.GetNode(p.NodeID)
	if n == nil {
		log.Printf("node not found")
		return echo.ErrBadRequest
	}
	air, err := datastore.LoadAIReesult(p.ID)
	if err != nil || len(air.ScoreData) < 1 {
		log.Printf("ai result not found")
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
				if err := datastore.DeleteAIResult(p.ID); err != nil {
					log.Printf("deleteAIResult err=%v", err)
				}
				return true
			})
		}()
	} else {
		if err := datastore.DeleteAIResult(id); err != nil {
			log.Printf("deleteAIResult err=%v", err)
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
