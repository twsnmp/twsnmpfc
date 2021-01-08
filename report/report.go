// Package report : ポーリング処理
package report

import (
	"context"
	"net"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"github.com/twsnmp/twsnmpfc/datastore"
)

type Report struct {
	ds              *datastore.DataStore
	devices         map[string]*deviceEnt
	users           map[string]*userEnt
	flows           map[string]*flowEnt
	servers         map[string]*serverEnt
	dennyRules      map[string]bool
	allowRules      map[string]*allowRuleEnt
	deviceReportCh  chan *deviceReportEnt
	userReportCh    chan *userReportEnt
	flowReportCh    chan *flowReportEnt
	serviceMap      map[string]string
	badIPs          map[string]int64
	protMap         map[int]string
	privateIPBlocks []*net.IPNet
	geoip           *geoip2.Reader
	geoipMap        map[string]string
}

type deviceReportEnt struct {
	Time int64
	MAC  string
	IP   string
}

type userReportEnt struct {
	Time   int64
	UserID string
	Server string
	Client string
	Ok     bool
}

type flowReportEnt struct {
	Time    int64
	SrcIP   string
	SrcPort int
	DstIP   string
	DstPort int
	Prot    int
	Bytes   int64
}

type deviceEnt struct {
	ID         string // MAC Addr
	Name       string
	IP         string
	Vendor     string
	Services   map[string]int64
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type userEnt struct {
	ID         string // User ID + Server
	UserID     string
	Server     string
	ServerName string
	Clients    map[string]int64
	Total      int
	Ok         int
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type serverEnt struct {
	ID         string //  ID Server
	Server     string
	Services   map[string]int64
	Count      int64
	Bytes      int64
	ServerName string
	Loc        string
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type flowEnt struct {
	ID         string // ID Client:Server
	Client     string
	Server     string
	Services   map[string]int64
	Count      int64
	Bytes      int64
	ClientName string
	ClientLoc  string
	ServerName string
	ServerLoc  string
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

// allowRuleEnt : 特定のサービスは特定のサーバーに限定するルール
type allowRuleEnt struct {
	Service string // Service
	Servers map[string]bool
}

func NewPolling(ctx context.Context, ds *datastore.DataStore) (*Report, error) {
	r := &Report{
		ds:             ds,
		devices:        make(map[string]*deviceEnt),
		users:          make(map[string]*userEnt),
		flows:          make(map[string]*flowEnt),
		servers:        make(map[string]*serverEnt),
		dennyRules:     make(map[string]bool),
		allowRules:     make(map[string]*allowRuleEnt),
		deviceReportCh: make(chan *deviceReportEnt, 100),
		userReportCh:   make(chan *userReportEnt, 100),
		flowReportCh:   make(chan *flowReportEnt, 500),
		serviceMap:     make(map[string]string),
		badIPs:         make(map[string]int64),
		protMap: map[int]string{
			1:   "icmp",
			2:   "igmp",
			6:   "tcp",
			8:   "egp",
			17:  "udp",
			112: "vrrp",
		},
		privateIPBlocks: []*net.IPNet{},
		geoip:           nil,
		geoipMap:        make(map[string]string),
	}
	return r, nil
}

func normMACAddr(m string) string {
	m = strings.Replace(m, "-", ":", -1)
	a := strings.Split(m, ":")
	r := ""
	for _, e := range a {
		if r != "" {
			r += ":"
		}
		if len(e) == 1 {
			r += "0"
		}
		r += e
	}
	return strings.ToUpper(r)
}

func (r *Report) ReportDevice(mac, ip string, t int64) {
	mac = normMACAddr(mac)
	r.deviceReportCh <- &deviceReportEnt{
		Time: t,
		MAC:  mac,
		IP:   ip,
	}
}

func (r *Report) ReportUser(user, server, client string, ok bool, t int64) {
	r.userReportCh <- &userReportEnt{
		Time:   t,
		Server: server,
		Client: client,
		UserID: user,
		Ok:     ok,
	}
}

func (r *Report) ReportFlow(src string, sp int, dst string, dp, prot int, bytes int64, t int64) {
	r.flowReportCh <- &flowReportEnt{
		Time:    t,
		SrcIP:   src,
		SrcPort: sp,
		DstIP:   dst,
		DstPort: dp,
		Prot:    prot,
		Bytes:   bytes,
	}
}
