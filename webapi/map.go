package webapi

import (
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
