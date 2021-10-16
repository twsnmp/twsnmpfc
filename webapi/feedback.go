// Package webapi : WEB API
package webapi

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type feedbackWebAPI struct {
	Msg            string
	IncludeSysInfo bool
}

func postFeedback(c echo.Context) error {
	fb := new(feedbackWebAPI)
	if err := c.Bind(fb); err != nil {
		log.Printf("send feedback err=%v", err)
		return echo.ErrBadRequest
	}
	if err := sendFeedback(fb); err != nil {
		log.Printf("send feedback err=%v", err)
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "フィードバックを送信しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func sendFeedback(fb *feedbackWebAPI) error {
	msg := fb.Msg
	if fb.IncludeSysInfo {
		msg += fmt.Sprintf("\n-----\nGOOS=%s,GOARCH=%s\n", runtime.GOOS, runtime.GOARCH)
		msg += fmt.Sprintf("DBSize=%d\n", datastore.DBStats.Size)
		if len(backend.MonitorDataes) > 0 {
			i := len(backend.MonitorDataes) - 1
			msg += fmt.Sprintf("CPU=%f,Mem=%f,Load=%f,Disk=%f\n",
				backend.MonitorDataes[i].CPU,
				backend.MonitorDataes[i].Mem,
				backend.MonitorDataes[i].Load,
				backend.MonitorDataes[i].Disk,
			)
		}
	}
	values := url.Values{}
	values.Set("msg", msg)
	values.Add("hash", calcHash(msg))

	req, err := http.NewRequest(
		"POST",
		"https://lhx98.linkclub.jp/twise.co.jp/cgi-bin/twsnmpfb.cgi",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(r) != "OK" {
		return fmt.Errorf("resp is '%s'", r)
	}
	return nil
}

func calcHash(msg string) string {
	h := sha256.New()
	if _, err := h.Write([]byte(msg + time.Now().Format("2006/01/02T15"))); err != nil {
		log.Printf("calc hash err=%v", err)
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getCheckUpdate(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	url := "https://lhx98.linkclub.jp/twise.co.jp/cgi-bin/twsnmpfc.cgi?ver=" + api.Version
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("check new version err=%v", err)
		return echo.ErrInternalServerError
	}
	defer resp.Body.Close()
	ba, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("check new version err=%v", err)
		return echo.ErrInternalServerError
	}
	sv := strings.TrimSpace(string(ba))
	return c.JSON(http.StatusOK, map[string]interface{}{"Version": sv, "HasNew": backend.CmpVersion(api.Version, sv) < 0})
}
