package webapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type WebAPIIfPortEnt struct {
	IfIndex         int
	Status          string
	Node            string
	Name            string
	InBPS           uint64
	InOctets        uint64
	OutBPS          uint64
	OutOctets       uint64
	FirstCheckTime  int64
	LastCheckTime   int64
	LastChangedTime int64
	Changed         int
}

func getIfPortTable(c echo.Context) error {
	r := []*WebAPIIfPortEnt{}
	datastore.ForEachIfPortTable(func(id string, l *[]datastore.IfPortEnt) bool {
		node := "Unknow"
		if n := datastore.GetNetwork(id); n != nil {
			node = n.Name
		} else if n := datastore.GetNode(id); n != nil {
			node = n.Name
		}
		for _, e := range *l {
			r = append(r, &WebAPIIfPortEnt{
				Status:          getIfStatus(e.OperStatus, e.AdminStatus),
				Node:            node,
				IfIndex:         e.IfIndex,
				Name:            e.Name,
				InOctets:        e.InOctets,
				InBPS:           e.InBPS,
				OutOctets:       e.OutOctets,
				OutBPS:          e.OutBPS,
				FirstCheckTime:  e.FirstCheckTime,
				LastCheckTime:   e.LastCheckTime,
				LastChangedTime: e.LastChangedTime,
				Changed:         e.Changed,
			})
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func getIfStatus(o, a int) string {
	if o == 1 {
		return "up"
	}
	if a == 2 {
		return "off"
	}
	if o == 2 {
		return "down"
	}
	return "unknown"
}

func deleteIfPortTable(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("ifPortTable")
	} else {
		datastore.DeleteReport("ifPortTable", []string{id})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("IFポートテーブルを削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
