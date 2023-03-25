package webapi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type mapWebAPI struct {
	LastUpdate int64
	MapConf    *datastore.MapConfEnt
	Nodes      map[string]*datastore.NodeEnt
	Items      map[string]*datastore.DrawItemEnt
	Lines      []*datastore.LineEnt
	Pollings   map[string][]*datastore.PollingEnt
	Logs       []*datastore.EventLogEnt
	Images     []string
}

func getMap(c echo.Context) error {
	r := &mapWebAPI{
		MapConf:  &datastore.MapConf,
		Nodes:    make(map[string]*datastore.NodeEnt),
		Items:    make(map[string]*datastore.DrawItemEnt),
		Lines:    []*datastore.LineEnt{},
		Pollings: make(map[string][]*datastore.PollingEnt),
		Images:   datastore.GetImageList(),
	}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r.Nodes[n.ID] = n
		return true
	})
	datastore.ForEachItems(func(di *datastore.DrawItemEnt) bool {
		checkDrawItem(di)
		r.Items[di.ID] = di
		return true
	})
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		r.Lines = append(r.Lines, l)
		return true
	})
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		r.Pollings[p.NodeID] = append(r.Pollings[p.NodeID], p)
		return true
	})
	i := 0
	datastore.ForEachLastEventLog("", func(e *datastore.EventLogEnt) bool {
		r.Logs = append(r.Logs, e)
		i++
		return i < 100
	})
	r.LastUpdate = time.Now().Unix()
	return c.JSON(http.StatusOK, r)
}

func checkDrawItem(di *datastore.DrawItemEnt) {
	if di.Type < 4 || di.PollingID == "" {
		return
	}
	if di.Type == 4 {
		di.Text = "値なし"
	}
	if di.Type == 5 {
		di.Value = 0.0
	}
	p := datastore.GetPolling(di.PollingID)
	if p == nil {
		return
	}
	i, ok := p.Result[di.VarName]
	if !ok {
		return
	}
	text := ""
	val := 0.0
	switch v := i.(type) {
	case string:
		if di.Format == "" {
			text = v
		} else {
			text = fmt.Sprintf(di.Format, v)
		}
	case float64:
		if di.Scale != 0.0 {
			v *= di.Scale
		}
		if di.Format == "" {
			text = fmt.Sprintf("%f", v)
		} else {
			text = fmt.Sprintf(di.Format, v)
		}
		val = v
	}
	if text == "" {
		text = "値が空"
	}
	switch di.Type {
	case datastore.DrawItemTypePollingGauge:
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
		log.Println(err)
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
		log.Println("no draw item")
		return echo.ErrBadRequest
	}
	odi.W = di.W
	odi.H = di.H
	odi.Path = di.Path
	odi.Text = di.Text
	odi.Size = di.Size
	odi.Color = di.Color
	odi.Format = di.Format
	odi.VarName = di.VarName
	odi.Scale = di.Scale
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getImage(c echo.Context) error {
	path := c.Param("path")
	img, err := datastore.GetImage(path)
	if err != nil {
		return echo.ErrNotFound
	}
	ct := http.DetectContentType(img)
	return c.Blob(http.StatusOK, ct, img)
}

func deleteImage(c echo.Context) error {
	path := c.Param("path")
	err := datastore.DelteImage(path)
	if err != nil {
		return echo.ErrNotFound
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postImage(c echo.Context) error {
	f, err := c.FormFile("file")
	if err != nil {
		f = nil
	}
	path := f.Filename
	if path == "" || f == nil || f.Size > 1024*1024*2 {
		return echo.ErrBadRequest
	}
	fp, err := f.Open()
	if err != nil {
		return echo.ErrBadRequest
	}
	defer fp.Close()
	img, err := ioutil.ReadAll(fp)
	if err != nil {
		return echo.ErrBadRequest
	}
	if err = datastore.SaveImage(path, img); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
