package webapi

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/wol"
)

type mapWebAPI struct {
	LastUpdate int64
	MapConf    *datastore.MapConfEnt
	Nodes      map[string]*datastore.NodeEnt
	Lines      []*datastore.LineEnt
	Pollings   map[string][]*datastore.PollingEnt
	Logs       []*datastore.EventLogEnt
}

func getMap(c echo.Context) error {
	r := &mapWebAPI{
		MapConf:  &datastore.MapConf,
		Nodes:    make(map[string]*datastore.NodeEnt),
		Lines:    []*datastore.LineEnt{},
		Pollings: make(map[string][]*datastore.PollingEnt),
	}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r.Nodes[n.ID] = n
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

type nodePosWebAPI struct {
	ID string
	X  int
	Y  int
}

func postMapUpdate(c echo.Context) error {
	list := []nodePosWebAPI{}
	if err := c.Bind(&list); err != nil {
		return echo.ErrBadRequest
	}
	for _, nu := range list {
		n := datastore.GetNode(nu.ID)
		if n == nil {
			return echo.ErrBadRequest
		}
		n.X = nu.X
		n.Y = nu.Y
		if err := datastore.UpdateNode(n); err != nil {
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func getNodes(c echo.Context) error {
	r := []*datastore.NodeEnt{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		r = append(r, n)
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func deleteNodes(c echo.Context) error {
	ids := []string{}
	if err := c.Bind(&ids); err != nil {
		return echo.ErrBadRequest
	}
	for _, id := range ids {
		if err := datastore.DeleteNode(id); err != nil {
			return echo.ErrBadRequest
		}
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postNodeUpdate(c echo.Context) error {
	nu := new(datastore.NodeEnt)
	if err := c.Bind(nu); err != nil {
		return echo.ErrBadRequest
	}
	if nu.ID == "" {
		if err := datastore.AddNode(nu); err != nil {
			return echo.ErrBadRequest
		}
		from := c.QueryParam("from")
		if from != "" {
			copyPolling(nu.ID, from)
		}
		return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
	}
	// ここで入力チェック
	n := datastore.GetNode(nu.ID)
	if n == nil {
		return echo.ErrBadRequest
	}
	n.Name = nu.Name
	n.Descr = nu.Descr
	n.IP = nu.IP
	n.Icon = nu.Icon
	n.SnmpMode = nu.SnmpMode
	n.Community = nu.Community
	n.User = nu.User
	n.Password = nu.Password
	n.PublicKey = nu.PublicKey
	n.URL = nu.URL
	n.Type = nu.Type
	n.AddrMode = nu.AddrMode
	if err := datastore.UpdateNode(n); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		NodeID:   n.ID,
		Event:    "ノードを更新しました",
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func deleteLine(c echo.Context) error {
	l := new(datastore.LineEnt)
	if err := c.Bind(l); err != nil {
		return echo.ErrBadRequest
	}
	if err := datastore.DeleteLine(l.ID); err != nil {
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:   "user",
		Level:  "info",
		NodeID: l.NodeID1,
		Event:  fmt.Sprintf("ラインを削除しました(%s)", l.ID),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

func postLineAdd(c echo.Context) error {
	lu := new(datastore.LineEnt)
	if err := c.Bind(lu); err != nil {
		return echo.ErrBadRequest
	}
	if lu.PollingID1 == "" || lu.PollingID2 == "" {
		return echo.ErrBadRequest
	}
	if p := datastore.GetPolling(lu.PollingID1); p != nil {
		lu.State1 = p.State
	}
	if p := datastore.GetPolling(lu.PollingID2); p != nil {
		lu.State2 = p.State
	}
	l := datastore.GetLine(lu.ID)
	if l == nil {
		if err := datastore.AddLine(lu); err != nil {
			return echo.ErrBadRequest
		}
	} else {
		l.NodeID1 = lu.NodeID1
		l.NodeID2 = lu.NodeID2
		l.PollingID1 = lu.PollingID1
		l.PollingID2 = lu.PollingID2
		l.State1 = lu.State1
		l.State2 = lu.State2
		l.Info = lu.Info
		l.PollingID = lu.PollingID
		l.Width = lu.Width
		if err := datastore.UpdateLine(l); err != nil {
			return echo.ErrBadRequest
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:   "user",
		Level:  "info",
		NodeID: lu.NodeID1,
		Event:  fmt.Sprintf("ラインを更新しました(%s)", lu.ID),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

type nodeWebAPI struct {
	Node     *datastore.NodeEnt
	Logs     []*datastore.EventLogEnt
	Pollings []*datastore.PollingEnt
}

func getNodeLog(c echo.Context) error {
	id := c.Param("id")
	r := nodeWebAPI{}
	r.Node = datastore.GetNode(id)
	if r.Node == nil {
		return echo.ErrBadRequest
	}
	i := 0
	st := time.Now().Add(-time.Hour * 24).UnixNano()
	et := time.Now().UnixNano()
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if l.NodeID != id {
			return true
		}
		r.Logs = append(r.Logs, l)
		i++
		return i <= datastore.MapConf.LogDispSize
	})
	return c.JSON(http.StatusOK, r)
}

func getNodePolling(c echo.Context) error {
	id := c.Param("id")
	r := nodeWebAPI{}
	r.Node = datastore.GetNode(id)
	if r.Node == nil {
		return echo.ErrBadRequest
	}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == id {
			r.Pollings = append(r.Pollings, p)
		}
		return true
	})
	return c.JSON(http.StatusOK, r)
}

func postWOL(c echo.Context) error {
	id := c.Param("id")
	n := datastore.GetNode(id)
	if n == nil || n.MAC == "" {
		log.Printf("postWOL node not found")
		return echo.ErrBadRequest
	}
	a := strings.SplitN(n.MAC, "(", 2)
	if len(a) < 1 || a[0] == "" {
		log.Printf("postWOL no MAC")
		return echo.ErrBadRequest
	}
	if err := wol.SendWakeOnLanPacket(a[0]); err != nil {
		log.Printf("postWOL node not found")
		return echo.ErrBadRequest
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		NodeID:   n.ID,
		Event:    fmt.Sprintf("%sにWake ON LANパケットを送信しました", n.MAC),
	})
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}

//
type vpanelWebAPI struct {
	Node  *datastore.NodeEnt
	Ports []*backend.VPanelPortEnt
	Power bool
}

func getVPanel(c echo.Context) error {
	id := c.Param("id")
	r := vpanelWebAPI{}
	r.Node = datastore.GetNode(id)
	if r.Node == nil {
		return echo.ErrBadRequest
	}
	r.Power = backend.GetVPanelPowerInfo(r.Node)
	r.Ports = backend.GetVPanelPorts(r.Node)
	return c.JSON(http.StatusOK, r)
}
