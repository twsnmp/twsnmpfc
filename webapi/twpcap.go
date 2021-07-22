package webapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getEtherType(c echo.Context) error {
	r := []*datastore.EtherTypeEnt{}
	datastore.ForEachEtherType(func(f *datastore.EtherTypeEnt) bool {
		r = append(r, f)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteEtherType(c echo.Context) error {
	go datastore.ClearReport("ether")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "Ethernetタイプレポートを削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getDNSQ(c echo.Context) error {
	r := []*datastore.DNSQEnt{}
	datastore.ForEachDNSQ(func(e *datastore.DNSQEnt) bool {
		r = append(r, e)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteDNSQ(c echo.Context) error {
	go datastore.ClearReport("dnsq")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "DNS問い合わせレポートを削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getRADIUSFlows(c echo.Context) error {
	r := []*datastore.RADIUSFlowEnt{}
	datastore.ForEachRADIUSFlows(func(f *datastore.RADIUSFlowEnt) bool {
		if f.ValidScore {
			r = append(r, f)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteRADIUSFlow(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("radius")
	} else {
		datastore.DeleteReport("radius", id)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("RADIUSレポートを削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetRADIUSFlows(c echo.Context) error {
	report.ResetRADIUSFlowsScore()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "RADIUSレポートの信用スコアを再計算しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
func getTLSFlows(c echo.Context) error {
	r := []*datastore.TLSFlowEnt{}
	datastore.ForEachTLSFlows(func(f *datastore.TLSFlowEnt) bool {
		if f.ValidScore {
			r = append(r, f)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteTLSFlow(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("tls")
	} else {
		datastore.DeleteReport("tls", id)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("TLSフローを削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetTLSFlows(c echo.Context) error {
	report.ResetTLSFlowsScore()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "TLSフローレポートの信用スコアを再計算しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
