package webapi

import (
	"io"
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
	Networks   map[string]*datastore.NetworkEnt
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
		Networks: make(map[string]*datastore.NetworkEnt),
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
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		r.Networks[n.ID] = n
		return true
	})
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		r.Lines = append(r.Lines, l)
		return true
	})
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		r.Pollings[p.NodeID] = append(r.Pollings[p.NodeID], &datastore.PollingEnt{
			ID:        p.ID,
			Name:      p.Name,
			NodeID:    p.NodeID,
			Type:      p.Type,
			Mode:      p.Mode,
			Params:    p.Params,
			Filter:    p.Filter,
			Extractor: p.Extractor,
			State:     p.State,
			LogMode:   p.LogMode,
			LastTime:  p.LastTime,
		})
		return true
	})
	i := 0
	datastore.ForEachLastEventLog(0, func(e *datastore.EventLogEnt) bool {
		r.Logs = append(r.Logs, e)
		i++
		return i < 100
	})
	r.LastUpdate = time.Now().Unix()
	return c.JSON(http.StatusOK, r)
}

func setPollingLogValuesForLine(di *datastore.DrawItemEnt) {
	st := time.Now().Add(-time.Hour * 24).UnixNano()
	et := time.Now().UnixNano()
	di.Values = []float64{}
	datastore.ForEachPollingLog(st, et, di.PollingID, func(l *datastore.PollingLogEnt) bool {
		if v, ok := l.Result[di.VarName]; ok {
			if val, ok := v.(float64); ok {
				di.Values = append(di.Values, val)
			}
		}
		return true
	})
}

func autoGetPollingSetting(di *datastore.DrawItemEnt, p *datastore.PollingEnt) (varName, format string, scale float64) {
	varName = di.VarName
	format = di.Format
	scale = di.Scale
	if scale == 0.0 {
		scale = 1.0
	}
	// ポーリングだけ選択して変数が空欄なら自動で設定する
	if varName != "" {
		return
	}
	// 値があるものを優先的に返す
	if _, ok := p.Result["bps"]; ok {
		varName = "bps"
		if format == "" {
			format = "BPS"
		}
		scale = 1.0
		return
	}
	if _, ok := p.Result["rtt"]; ok {
		varName = "rtt"
		if format == "" {
			format = "RTT=%.3fSec"
		}
		scale = 0.000000001
		return
	}
	if _, ok := p.Result["state"]; ok {
		varName = "state"
		format = "%s"
		return
	}
	if _, ok := p.Result["avg"]; ok {
		varName = "avg"
		if format == "" {
			format = "AVG=%.2f"
		}
		return
	}
	if _, ok := p.Result["count"]; ok {
		varName = "count"
		if format == "" {
			format = "COUNT=%.0f"
		}
		return
	}
	// 自動選択できないものは、値なしを表示する
	return
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
	img, err := io.ReadAll(fp)
	if err != nil {
		return echo.ErrBadRequest
	}
	if err = datastore.SaveImage(path, img); err != nil {
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getImageIconList(c echo.Context) error {
	return c.JSON(http.StatusOK, datastore.ImageIcons)
}

func getImageIcon(c echo.Context) error {
	id := c.Param("id")
	img, err := datastore.GetImageIcon(id)
	if err != nil {
		return echo.ErrNotFound
	}
	ct := http.DetectContentType(img)
	return c.Blob(http.StatusOK, ct, img)
}
