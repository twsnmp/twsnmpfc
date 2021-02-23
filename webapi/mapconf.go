package webapi

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/security"
)

func getMapConf(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	r := new(datastore.MapConfEnt)
	r.MapName = api.DataStore.MapConf.MapName
	r.UserID = api.DataStore.MapConf.UserID
	//	r.Password = api.DataStore.MapConf.Password
	r.PollInt = api.DataStore.MapConf.PollInt
	r.Timeout = api.DataStore.MapConf.Timeout
	r.Retry = api.DataStore.MapConf.Retry
	r.LogDays = api.DataStore.MapConf.LogDays
	r.LogDispSize = api.DataStore.MapConf.LogDispSize
	r.SnmpMode = api.DataStore.MapConf.SnmpMode
	r.Community = api.DataStore.MapConf.Community
	r.SnmpUser = api.DataStore.MapConf.SnmpUser
	//	r.SnmpPassword = api.DataStore.MapConf.SmmpPassword
	r.EnableSyslogd = api.DataStore.MapConf.EnableSyslogd
	r.EnableTrapd = api.DataStore.MapConf.EnableTrapd
	r.EnableNetflowd = api.DataStore.MapConf.EnableNetflowd
	r.AILevel = api.DataStore.MapConf.AILevel
	r.AIThreshold = api.DataStore.MapConf.AIThreshold
	r.BackImage = api.DataStore.MapConf.BackImage
	return c.JSON(http.StatusOK, r)
}

func postMapConf(c echo.Context) error {
	mc := new(datastore.MapConfEnt)
	if err := c.Bind(mc); err != nil {
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	api.DataStore.MapConf.MapName = mc.MapName
	api.DataStore.MapConf.UserID = mc.UserID
	if mc.Password != "" {
		api.DataStore.MapConf.Password = security.PasswordHash(mc.Password)
	}
	api.DataStore.MapConf.PollInt = mc.PollInt
	api.DataStore.MapConf.Timeout = mc.Timeout
	api.DataStore.MapConf.Retry = mc.Retry
	api.DataStore.MapConf.LogDays = mc.LogDays
	api.DataStore.MapConf.LogDispSize = mc.LogDispSize
	api.DataStore.MapConf.SnmpMode = mc.SnmpMode
	api.DataStore.MapConf.Community = mc.Community
	api.DataStore.MapConf.SnmpUser = mc.SnmpUser
	if mc.SnmpPassword != "" {
		api.DataStore.MapConf.SnmpPassword = mc.SnmpPassword
	}
	api.DataStore.MapConf.EnableSyslogd = mc.EnableSyslogd
	api.DataStore.MapConf.EnableTrapd = mc.EnableTrapd
	api.DataStore.MapConf.EnableNetflowd = mc.EnableNetflowd
	api.DataStore.MapConf.AILevel = mc.AILevel
	api.DataStore.MapConf.AIThreshold = mc.AIThreshold
	if err := api.DataStore.SaveMapConfToDB(); err != nil {
		return echo.ErrBadRequest
	}
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
	api := c.Get("api").(*WebAPI)
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
		if err = api.DataStore.SaveBackImage(img); err != nil {
			log.Printf("postBackImage err=%v", err)
			return echo.ErrBadRequest
		}
		api.DataStore.MapConf.BackImage.Path = f.Filename
	}
	api.DataStore.MapConf.BackImage.X = x
	api.DataStore.MapConf.BackImage.Y = y
	api.DataStore.MapConf.BackImage.Width = w
	api.DataStore.MapConf.BackImage.Height = h
	if err := api.DataStore.SaveMapConfToDB(); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getBackImage(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	img, err := api.DataStore.GetBackImage()
	if err != nil {
		log.Printf("postBackImage err=%v", err)
		return echo.ErrNotFound
	}
	ct := http.DetectContentType(img)
	return c.Blob(http.StatusOK, ct, img)
}

func deleteBackImage(c echo.Context) error {
	api := c.Get("api").(*WebAPI)
	if err := api.DataStore.SaveBackImage([]byte{}); err != nil {
		log.Printf("postBackImage err=%v", err)
		return echo.ErrBadRequest
	}
	api.DataStore.MapConf.BackImage.Path = ""
	if err := api.DataStore.SaveMapConfToDB(); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
