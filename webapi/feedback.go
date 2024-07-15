// Package webapi : WEB API
package webapi

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
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
		if len(datastore.MonitorDataes) > 0 {
			i := len(datastore.MonitorDataes) - 1
			msg += fmt.Sprintf("CPU=%f,Mem=%f,Load=%f,Disk=%f\n",
				datastore.MonitorDataes[i].CPU,
				datastore.MonitorDataes[i].Mem,
				datastore.MonitorDataes[i].Load,
				datastore.MonitorDataes[i].Disk,
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
	r, err := io.ReadAll(resp.Body)
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
	ba, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("check new version err=%v", err)
		return echo.ErrInternalServerError
	}
	sv := strings.TrimSpace(string(ba))
	return c.JSON(http.StatusOK, map[string]interface{}{"Version": sv, "HasNew": cmpVersion(api.Version, sv) < 0})
}

func cmpVersion(mv, sv string) int {
	mv = strings.ReplaceAll(mv, "(", " ")
	mv = strings.ReplaceAll(mv, "v", "")
	mv = strings.ReplaceAll(mv, "x", "0")
	sv = strings.ReplaceAll(sv, "v", "")
	mva := strings.Split(mv, ".")
	sva := strings.Split(sv, ".")
	for i := 0; i < len(mva) && i < len(sva) && i < 3; i++ {
		sn, err := strconv.ParseInt(sva[i], 10, 64)
		if err != nil {
			log.Println(err)
			return 0
		}
		mn, err := strconv.ParseInt(mva[i], 10, 64)
		if err != nil {
			log.Println(err)
			return 0
		}
		if sn > mn {
			return -1
		}
		if sn < mn {
			return 1
		}
	}
	return 0
}
