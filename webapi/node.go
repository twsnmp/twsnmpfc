package webapi

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/backend"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/logger"
	"github.com/twsnmp/twsnmpfc/ping"
	"github.com/twsnmp/twsnmpfc/wol"
)

type nodePosWebAPI struct {
	ID string
	X  int
	Y  int
}

func postNodePos(c echo.Context) error {
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
		log.Printf("update node bind err=%v", err)
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
		log.Printf("update node node not found node=%v", nu)
		return echo.ErrBadRequest
	}
	n.Name = nu.Name
	n.Descr = nu.Descr
	n.IP = nu.IP
	n.Icon = nu.Icon
	n.Image = nu.Image
	n.SnmpMode = nu.SnmpMode
	n.Community = nu.Community
	n.User = nu.User
	n.Password = nu.Password
	n.GNMIUser = nu.GNMIUser
	n.GNMIPassword = nu.GNMIPassword
	n.GNMIEncoding = nu.GNMIEncoding
	n.GNMIPort = nu.GNMIPort
	n.PublicKey = nu.PublicKey
	n.URL = nu.URL
	n.AddrMode = nu.AddrMode
	n.AutoAck = nu.AutoAck
	if n.MAC != nu.MAC {
		if nu.MAC != "" {
			mac := logger.NormMACAddr(nu.MAC)
			v := datastore.FindVendor(mac)
			if v != "" {
				mac += fmt.Sprintf("(%s)", v)
			}
			n.MAC = mac
		} else {
			n.MAC = ""
		}
	}
	datastore.UpdateNode(n)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		NodeID:   n.ID,
		Event:    "ノードを更新しました",
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

// vpanelデータ
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

// ホストリソースデータ
type hostResorceWebAPI struct {
	Node         *datastore.NodeEnt
	HostResource *backend.HostResourceEnt
}

func getHostResource(c echo.Context) error {
	id := c.Param("id")
	r := hostResorceWebAPI{}
	r.Node = datastore.GetNode(id)
	if r.Node == nil {
		log.Printf("host resorce node not found id=%s", id)
		return echo.ErrBadRequest
	}
	r.HostResource = backend.GetHostResource(r.Node)
	return c.JSON(http.StatusOK, r)
}

type PingReq struct {
	IP   string
	Size int
	TTL  int
}

type PingRes struct {
	Stat      int
	TimeStamp int64
	Time      int64
	Size      int
	SendTTL   int
	RecvTTL   int
	RecvSrc   string
	Loc       string
}

func postPing(c echo.Context) error {
	req := new(PingReq)
	if err := c.Bind(req); err != nil {
		log.Println(err)
		return echo.ErrBadRequest
	}
	ipreg := regexp.MustCompile(`^[0-9.]+$`)
	if !ipreg.MatchString(req.IP) {
		if ips, err := net.LookupIP(req.IP); err == nil {
			for _, ip := range ips {
				if ip.IsGlobalUnicast() {
					s := ip.To4().String()
					if ipreg.MatchString(s) {
						req.IP = s
						break
					}
				}
			}
		}
	}
	res := new(PingRes)
	pe := ping.DoPing(req.IP, 2, 0, req.Size, req.TTL)
	res.Stat = int(pe.Stat)
	res.TimeStamp = time.Now().Unix()
	res.Time = pe.Time
	res.Size = pe.Size
	res.RecvSrc = pe.RecvSrc
	res.RecvTTL = pe.RecvTTL
	res.SendTTL = req.TTL
	if pe.RecvSrc != "" {
		res.Loc = datastore.GetLoc(pe.RecvSrc)
	}
	return c.JSON(http.StatusOK, res)
}

type portWebAPI struct {
	Node     *datastore.NodeEnt
	TCPPorts []*backend.PortEnt
	UDPPorts []*backend.PortEnt
}

func getPortList(c echo.Context) error {
	id := c.Param("id")
	r := portWebAPI{}
	r.Node = datastore.GetNode(id)
	if r.Node == nil {
		log.Printf("host resorce node not found id=%s", id)
		return echo.ErrBadRequest
	}
	r.TCPPorts, r.UDPPorts = backend.GetPortList(r.Node)
	return c.JSON(http.StatusOK, r)
}

type rmonWebAPI struct {
	Node *datastore.NodeEnt
	RMON *backend.RMONEnt
}

func getRMON(c echo.Context) error {
	id := c.Param("id")
	t := c.Param("type")
	ret := new(rmonWebAPI)
	ret.Node = datastore.GetNode(id)
	if ret.Node == nil {
		log.Printf("rmon node not found id=%s", id)
		return echo.ErrBadRequest
	}
	ret.RMON = backend.GetRMON(ret.Node, t)
	return c.JSON(http.StatusOK, ret)
}

type NodeMemo struct {
	ID   string
	Memo string
}

func getNodeMemo(c echo.Context) error {
	id := c.Param("id")
	memo := datastore.GetNodeMemo(id)
	return c.JSON(http.StatusOK, NodeMemo{
		ID:   id,
		Memo: memo,
	})
}

func postNodeMemo(c echo.Context) error {
	req := new(NodeMemo)
	if err := c.Bind(req); err != nil {
		log.Printf("post node memo err=%v", err)
		return echo.ErrBadRequest
	}
	if err := datastore.SaveNodeMemo(req.ID, req.Memo); err != nil {
		log.Printf("post node memo err=%v", err)
		return echo.ErrBadRequest
	}
	return c.JSON(http.StatusOK, map[string]string{"resp": "ok"})
}
