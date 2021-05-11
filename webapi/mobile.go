package webapi

import (
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"
)

// API
type mobileAPIMapStatusEnt struct {
	High      int
	Low       int
	Warn      int
	Normal    int
	Repair    int
	Unknown   int
	DBSize    int64
	DBSizeStr string
	State     string
}

func getMobileMapStatus(c echo.Context) error {
	ms := new(mobileAPIMapStatusEnt)
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		switch n.State {
		case "high":
			ms.High++
		case "low":
			ms.Low++
		case "warn":
			ms.Warn++
		case "normal":
			ms.Normal++
		case "repair":
			ms.Repair++
		default:
			ms.Unknown++
		}
		return true
	})
	if ms.High > 0 {
		ms.State = "high"
	} else if ms.Low > 0 {
		ms.State = "low"
	} else if ms.Warn > 0 {
		ms.State = "warn"
	} else if ms.Normal+ms.Repair > 0 {
		ms.State = "normal"
	} else {
		ms.State = "unknown"
	}
	ms.DBSize = datastore.DBStats.Size
	ms.DBSizeStr = humanize.Bytes(uint64(datastore.DBStats.Size))
	return c.JSON(http.StatusOK, ms)
}

type mobileAPIMapDataEnt struct {
	LastTime int64
	MapName  string
	Nodes    map[string]*mobileAPINodeEnt
}

type mobileAPINodeEnt struct {
	ID    string
	Name  string
	Descr string
	Icon  string
	State string
	X     int
	Y     int
	IP    string
	MAC   string
}

var mobileAPIMapData = mobileAPIMapDataEnt{
	Nodes: make(map[string]*mobileAPINodeEnt),
}

func makeMobileAPIMapData() {
	if mobileAPIMapData.LastTime > time.Now().Unix()-60 {
		return
	}
	mobileAPIMapData.MapName = datastore.MapConf.MapName
	mobileAPIMapData.LastTime = time.Now().Unix()
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		mobileAPIMapData.Nodes[n.ID] = &mobileAPINodeEnt{
			ID:    n.ID,
			Name:  n.Name,
			Descr: n.Descr,
			Icon:  n.Icon,
			State: n.State,
			X:     n.X,
			Y:     n.Y,
			IP:    n.IP,
			MAC:   n.MAC,
		}
		return true
	})
}

func getMobileMapData(c echo.Context) error {
	makeMobileAPIMapData()
	return c.JSON(http.StatusOK, mobileAPIMapData)
}
