package webapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type WebAPIFDBTableEnt struct {
	Node            string
	Port            string
	LinkedNode      string
	MAC             string
	Vendor          string
	VLanID          int
	LastCheckTime   int64
	FirstCheckTime  int64
	LastChangedTime int64
	Changed         int
}

func getIfPortName(id string, ifIndex int) string {
	if l := datastore.GetIfPortTable(id); l != nil {
		for _, e := range *l {
			if e.IfIndex == ifIndex {
				return e.Name
			}
		}
	}
	return ""
}

func getFDBTable(c echo.Context) error {
	r := []*WebAPIFDBTableEnt{}
	datastore.ForEachFDBTable(func(id string, l *[]datastore.FDBTableEnt) bool {
		node := "Unknow"
		if n := datastore.GetNetwork(id); n != nil {
			node = n.Name
		} else if n := datastore.GetNode(id); n != nil {
			node = n.Name
		}
		for _, e := range *l {
			port := getIfPortName(id, e.IfIndex)
			linkedNode := ""
			if n := datastore.FindNodeFromMAC(e.MAC); n != nil {
				linkedNode = fmt.Sprintf("%s(%s)", n.Name, n.IP)
			}
			r = append(r, &WebAPIFDBTableEnt{
				Node:            node,
				Port:            port,
				LinkedNode:      linkedNode,
				MAC:             e.MAC,
				Vendor:          datastore.FindVendor(e.MAC),
				VLanID:          e.VLanID,
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

func deleteFDBTable(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("fdbTable")
	} else {
		datastore.DeleteReport("fdbTable", []string{id})
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("FDBテーブルを削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
