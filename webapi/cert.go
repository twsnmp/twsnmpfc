package webapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/report"
)

func getCert(c echo.Context) error {
	r := []*datastore.CertEnt{}
	datastore.ForEachCerts(func(f *datastore.CertEnt) bool {
		r = append(r, f)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func postCert(c echo.Context) error {
	pc := new(struct {
		Target string
		Port   int
	})
	if err := c.Bind(pc); err != nil {
		log.Printf("postCert err=%v", err)
		return echo.ErrBadRequest
	}
	id := fmt.Sprintf("%s:%d", pc.Target, pc.Port)
	if datastore.GetCert(id) != nil {
		log.Printf("postCert duplicate id")
		return echo.ErrBadRequest
	}
	datastore.AddCert(&datastore.CertEnt{
		ID:     id,
		Target: pc.Target,
		Score:  50.0,
		Port:   uint16(pc.Port),
	})
	report.DoCheckCert()
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteCert(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {
		go datastore.ClearReport("cert")
	} else {
		datastore.DeleteReport("cert", id)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("サーバー証明書を削除しました(%s)", id),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func resetCert(c echo.Context) error {
	report.ResetCertScore()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "サーバー証明書の信用スコアを再計算しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}