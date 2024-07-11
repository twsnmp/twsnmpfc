package webapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

func checkDrawItem(di *datastore.DrawItemEnt) {
	if di.Type < 4 || di.PollingID == "" {
		return
	}
	if di.Type == 4 {
		di.Text = "値なし"
	}
	if di.Type >= 5 {
		di.Value = 0.0
	}
	p := datastore.GetPolling(di.PollingID)
	if p == nil {
		return
	}
	if di.Type == datastore.DrawItemTypePollingLine {
		switch p.State {
		case "high":
			di.Color = "#e31a1c"
		case "low":
			di.Color = "#fb9a99"
		case "warn":
			di.Color = "#dfdf22"
		default:
			di.Color = "#1f78b4"
		}
		setPollingLogValuesForLine(di)
		return
	}
	varName, format, scale := autoGetPollingSetting(di, p)
	i, ok := p.Result[varName]
	if !ok {
		return
	}
	text := ""
	val := 0.0
	switch v := i.(type) {
	case string:
		if format == "" {
			text = v
		} else {
			text = fmt.Sprintf(format, v)
		}
	case float64:
		v *= scale
		if format == "" {
			text = fmt.Sprintf("%f", v)
		} else if strings.Contains(format, "BPS") {
			bps := humanize.Bytes(uint64(v)) + "PS"
			text = strings.Replace(format, "BPS", bps, 1)
		} else if strings.Contains(format, "PPS") {
			pps := humanize.Commaf(v) + "PPS"
			text = strings.Replace(format, "PPS", pps, 1)
		} else {
			text = fmt.Sprintf(format, v)
		}
		val = v
	}
	if text == "" {
		text = "値なし"
	}
	switch di.Type {
	case datastore.DrawItemTypePollingGauge, datastore.DrawItemTypePollingNewGauge, datastore.DrawItemTypePollingBar:
		if val > 100.0 {
			val = 100.0
		}
		if val > 90.0 {
			di.Color = "#e31a1c"
		} else if val > 80.0 {
			di.Color = "#dfdf22"
		} else {
			di.Color = "#1f78b4"
		}
		di.Value = val
	case datastore.DrawItemTypePollingText:
		di.Text = text
		switch p.State {
		case "high":
			di.Color = "#e31a1c"
		case "low":
			di.Color = "#fb9a99"
		case "warn":
			di.Color = "#dfdf22"
		default:
			di.Color = "#eee"
		}
		di.Value = val
	}
}

type itemPosWebAPI struct {
	ID string
	X  int
	Y  int
}

func postItemPos(c echo.Context) error {
	list := []itemPosWebAPI{}
	if err := c.Bind(&list); err != nil {
		return echo.ErrBadRequest
	}
	for _, i := range list {
		di := datastore.GetDrawItem(i.ID)
		if di == nil {
			return echo.ErrBadRequest
		}
		di.X = i.X
		di.Y = i.Y
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteDrawItems(c echo.Context) error {
	ids := []string{}
	if err := c.Bind(&ids); err != nil {
		return echo.ErrBadRequest
	}
	for _, id := range ids {
		if err := datastore.DeleteDrawItem(id); err != nil {
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postItemUpdate(c echo.Context) error {
	di := new(datastore.DrawItemEnt)
	if err := c.Bind(di); err != nil {
		return echo.ErrBadRequest
	}
	if di.ID == "" {
		if err := datastore.AddDrawItem(di); err != nil {
			return echo.ErrBadRequest
		}
		return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
	}
	// ここで入力チェック
	odi := datastore.GetDrawItem(di.ID)
	if odi == nil {
		return echo.ErrBadRequest
	}
	odi.Type = di.Type
	odi.W = di.W
	odi.H = di.H
	odi.Path = di.Path
	odi.Text = di.Text
	odi.Size = di.Size
	odi.Color = di.Color
	odi.Format = di.Format
	odi.VarName = di.VarName
	odi.PollingID = di.PollingID
	odi.Scale = di.Scale
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
