package webapi

import (
	"io"
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
	r.LogTimeout = datastore.MapConf.LogTimeout
	r.SnmpMode = datastore.MapConf.SnmpMode
	r.Community = datastore.MapConf.Community
	r.SnmpUser = datastore.MapConf.SnmpUser
	r.EnableSyslogd = datastore.MapConf.EnableSyslogd
	r.EnableTrapd = datastore.MapConf.EnableTrapd
	r.EnableNetflowd = datastore.MapConf.EnableNetflowd
	r.EnableArpWatch = datastore.MapConf.EnableArpWatch
	r.EnableSshd = datastore.MapConf.EnableSshd
	r.EnableSflowd = datastore.MapConf.EnableSflowd
	r.EnableTcpd = datastore.MapConf.EnableTcpd
	r.AILevel = datastore.MapConf.AILevel
	r.AIThreshold = datastore.MapConf.AIThreshold
	r.AIMode = datastore.MapConf.AIMode
	r.BackImage = datastore.MapConf.BackImage
	r.GeoIPInfo = datastore.MapConf.GeoIPInfo
	r.EnableMobileAPI = datastore.MapConf.EnableMobileAPI
	r.PublicKey = datastore.MapConf.PublicKey
	r.FontSize = datastore.MapConf.FontSize
	r.MapSize = datastore.MapConf.MapSize
	r.IconSize = datastore.MapConf.IconSize
	r.AutoCharCode = datastore.MapConf.AutoCharCode
	r.DisableOperLog = datastore.MapConf.DisableOperLog
	r.ArpWatchRange = datastore.MapConf.ArpWatchRange
	if r.IconSize == 0 {
		r.IconSize = 32
	}
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
	datastore.MapConf.LogTimeout = mc.LogTimeout
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
	datastore.MapConf.EnableSshd = mc.EnableSshd
	datastore.MapConf.EnableTcpd = mc.EnableTcpd
	datastore.MapConf.EnableMobileAPI = mc.EnableMobileAPI
	datastore.MapConf.AILevel = mc.AILevel
	datastore.MapConf.AIThreshold = mc.AIThreshold
	datastore.MapConf.AIMode = mc.AIMode
	datastore.MapConf.FontSize = mc.FontSize
	datastore.MapConf.MapSize = mc.MapSize
	datastore.MapConf.IconSize = mc.IconSize
	datastore.MapConf.AutoCharCode = mc.AutoCharCode
	datastore.MapConf.ArpWatchRange = mc.ArpWatchRange
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

// Icon Listの取得
func getIcons(c echo.Context) error {
	r := datastore.GetIcons()
	return c.JSON(http.StatusOK, r)
}

func postIcon(c echo.Context) error {
	i := new(datastore.IconEnt)
	if err := c.Bind(i); err != nil {
		return echo.ErrBadRequest
	}
	if err := datastore.AddOrUpdateIcon(i); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteIcon(c echo.Context) error {
	icon := c.Param("icon")
	if err := datastore.DeleteIcon(icon); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postBackImage(c echo.Context) error {
	f, err := c.FormFile("file")
	if err != nil {
		f = nil
	}
	if f != nil && f.Size > 1024*1024*2 {
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
			return echo.ErrBadRequest
		}
		defer fp.Close()
		img, err := io.ReadAll(fp)
		if err != nil {
			return echo.ErrBadRequest
		}
		if err = datastore.SaveBackImage(img); err != nil {
			return echo.ErrBadRequest
		}
		datastore.MapConf.BackImage.Path = f.Filename
	}
	datastore.MapConf.BackImage.X = x
	datastore.MapConf.BackImage.Y = y
	datastore.MapConf.BackImage.Width = w
	datastore.MapConf.BackImage.Height = h
	datastore.MapConf.BackImage.Color = c.FormValue("Color")
	if err := datastore.SaveMapConf(); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "マップの背景を更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getBackImage(c echo.Context) error {
	img, err := datastore.GetBackImage()
	if err != nil {
		return echo.ErrNotFound
	}
	ct := http.DetectContentType(img)
	return c.Blob(http.StatusOK, ct, img)
}

func deleteBackImage(c echo.Context) error {
	if err := datastore.SaveBackImage([]byte{}); err != nil {
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
		return echo.ErrBadRequest
	}
	if f.Size > 1024*1024*200 {
		return echo.ErrBadRequest
	}
	api := c.Get("api").(*WebAPI)
	dp := filepath.Join(api.DataStorePath, "geoio.uplaod")
	src, err := f.Open()
	if err != nil {
		return echo.ErrBadRequest
	}
	defer src.Close()
	dst, err := os.Create(dp)
	if err != nil {
		return echo.ErrInternalServerError
	}
	if _, err = io.Copy(dst, src); err != nil {
		dst.Close()
		return echo.ErrInternalServerError
	}
	dst.Close()
	if err := datastore.UpdateGeoIP(dp); err != nil {
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

func postReGenarateSSHKey(c echo.Context) error {
	datastore.InitSecurityKey()
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: "SSHの鍵を再作成しました。",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getSshPublicKey(c echo.Context) error {
	return c.String(http.StatusOK, datastore.GetSshdPublicKeys())
}

type SshPublicKeyPostEnt struct {
	PublicKey string
}

func postSshPublicKey(c echo.Context) error {
	pk := new(SshPublicKeyPostEnt)
	if err := c.Bind(pk); err != nil {
		log.Printf("postSshPublicKey c=%+v err=%v", c, err)
		return echo.ErrBadRequest
	}
	if err := datastore.SaveSshdPublicKeys(pk.PublicKey); err != nil {
		log.Printf("SaveSshdPublicKeys err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
