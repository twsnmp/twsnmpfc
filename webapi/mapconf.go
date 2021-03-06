package webapi

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/security"
)

func getMapConf(c echo.Context) error {
	r := new(datastore.MapConfEnt)
	r.MapName = datastore.MapConf.MapName
	r.UserID = datastore.MapConf.UserID
	//	r.Password = datastore.MapConf.Password
	r.PollInt = datastore.MapConf.PollInt
	r.Timeout = datastore.MapConf.Timeout
	r.Retry = datastore.MapConf.Retry
	r.LogDays = datastore.MapConf.LogDays
	r.LogDispSize = datastore.MapConf.LogDispSize
	r.SnmpMode = datastore.MapConf.SnmpMode
	r.Community = datastore.MapConf.Community
	r.SnmpUser = datastore.MapConf.SnmpUser
	//	r.SnmpPassword = datastore.MapConf.SmmpPassword
	r.EnableSyslogd = datastore.MapConf.EnableSyslogd
	r.EnableTrapd = datastore.MapConf.EnableTrapd
	r.EnableNetflowd = datastore.MapConf.EnableNetflowd
	r.EnableArpWatch = datastore.MapConf.EnableArpWatch
	r.AILevel = datastore.MapConf.AILevel
	r.AIThreshold = datastore.MapConf.AIThreshold
	r.BackImage = datastore.MapConf.BackImage
	r.GeoIPInfo = datastore.MapConf.GeoIPInfo
	return c.JSON(http.StatusOK, r)
}

func postMapConf(c echo.Context) error {
	mc := new(datastore.MapConfEnt)
	if err := c.Bind(mc); err != nil {
		return echo.ErrBadRequest
	}
	datastore.MapConf.MapName = mc.MapName
	datastore.MapConf.UserID = mc.UserID
	if mc.Password != "" {
		datastore.MapConf.Password = security.PasswordHash(mc.Password)
	}
	datastore.MapConf.PollInt = mc.PollInt
	datastore.MapConf.Timeout = mc.Timeout
	datastore.MapConf.Retry = mc.Retry
	datastore.MapConf.LogDays = mc.LogDays
	datastore.MapConf.LogDispSize = mc.LogDispSize
	datastore.RestartSnmpTrapd = datastore.MapConf.SnmpMode != mc.SnmpMode ||
		datastore.MapConf.Community != mc.Community ||
		datastore.MapConf.SnmpUser != mc.SnmpUser
	datastore.MapConf.SnmpMode = mc.SnmpMode
	datastore.MapConf.Community = mc.Community
	datastore.MapConf.SnmpUser = mc.SnmpUser
	if mc.SnmpPassword != "" {
		datastore.RestartSnmpTrapd = datastore.RestartSnmpTrapd || datastore.MapConf.SnmpPassword != mc.SnmpPassword
		datastore.MapConf.SnmpPassword = mc.SnmpPassword
	}
	datastore.MapConf.EnableSyslogd = mc.EnableSyslogd
	datastore.MapConf.EnableTrapd = mc.EnableTrapd
	datastore.MapConf.EnableNetflowd = mc.EnableNetflowd
	datastore.MapConf.EnableArpWatch = mc.EnableArpWatch
	datastore.MapConf.EnableMobileAPI = mc.EnableMobileAPI
	datastore.MapConf.AILevel = mc.AILevel
	datastore.MapConf.AIThreshold = mc.AIThreshold
	if err := datastore.SaveMapConf(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "マップの設定を更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postBackImage(c echo.Context) error {
	f, err := c.FormFile("file")
	if err != nil {
		log.Printf("postBackImage err=%v", err)
		f = nil
	}
	if f != nil && f.Size > 1024*1024*2 {
		log.Printf("postBackImage size over=%v", f)
		return echo.ErrBadRequest
	}
	x, _ := strconv.Atoi(c.FormValue("X"))
	y, _ := strconv.Atoi(c.FormValue("Y"))
	w, _ := strconv.Atoi(c.FormValue("Width"))
	h, _ := strconv.Atoi(c.FormValue("Height"))
	if w == 0 || h == 0 {
		w = 0
		h = 0
	}
	if f != nil {
		fp, err := f.Open()
		if err != nil {
			log.Printf("postBackImage err=%v", err)
			return echo.ErrBadRequest
		}
		defer fp.Close()
		img, err := ioutil.ReadAll(fp)
		if err != nil {
			log.Printf("postBackImage err=%v", err)
			return echo.ErrBadRequest
		}
		if err = datastore.SaveBackImage(img); err != nil {
			log.Printf("postBackImage err=%v", err)
			return echo.ErrBadRequest
		}
		datastore.MapConf.BackImage.Path = f.Filename
	}
	datastore.MapConf.BackImage.X = x
	datastore.MapConf.BackImage.Y = y
	datastore.MapConf.BackImage.Width = w
	datastore.MapConf.BackImage.Height = h
	if err := datastore.SaveMapConf(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "マップの背景画像を更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getBackImage(c echo.Context) error {
	img, err := datastore.GetBackImage()
	if err != nil {
		log.Printf("postBackImage err=%v", err)
		return echo.ErrNotFound
	}
	ct := http.DetectContentType(img)
	return c.Blob(http.StatusOK, ct, img)
}

func deleteBackImage(c echo.Context) error {
	if err := datastore.SaveBackImage([]byte{}); err != nil {
		log.Printf("postBackImage err=%v", err)
		return echo.ErrBadRequest
	}
	datastore.MapConf.BackImage.Path = ""
	if err := datastore.SaveMapConf(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "マップの背景画像を削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postGeoIP(c echo.Context) error {
	f, err := c.FormFile("file")
	if err != nil {
		log.Printf("postGeoIP err=%v", err)
		return echo.ErrBadRequest
	}
	if f.Size > 1024*1024*200 {
		log.Printf("postGeoIP size over=%v", f)
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	dp := filepath.Join(api.DataStorePath, "geoio.uplaod")
	src, err := f.Open()
	if err != nil {
		log.Printf("postGeoIP err=%v", err)
		return echo.ErrBadRequest
	}
	defer src.Close()
	dst, err := os.Create(dp)
	if err != nil {
		log.Printf("postGeoIP err=%v", err)
		return echo.ErrInternalServerError
	}
	if _, err = io.Copy(dst, src); err != nil {
		dst.Close()
		log.Printf("postGeoIP err=%v", err)
		return echo.ErrInternalServerError
	}
	dst.Close()
	if err := datastore.UpdateGeoIP(dp); err != nil {
		log.Printf("postGeoIP err=%v", err)
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "IP位置情報DBを更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteGeoIP(c echo.Context) error {
	datastore.DeleteGeoIP()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "IP位置情報DBを削除しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
