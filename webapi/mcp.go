package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/labstack/echo/v4"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
)

var mcpSSEServer *server.SSEServer
var mcpStreamableHTTPServer *server.StreamableHTTPServer
var mcpAllow sync.Map

func startMCPServer(e *echo.Echo, mcpFrom string) {
	log.Println("start mcp server")
	mcpAllow.Store("127.0.0.1", true)
	for _, ip := range strings.Split(mcpFrom, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			mcpAllow.Store(ip, true)
		}
	}
	// Create MCP Server
	s := server.NewMCPServer(
		"TWSNMP MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)
	// Add tools to MCP server
	addGetNodeListTool(s)
	addGetNetworkListTool(s)
	addGetPollingListTool(s)
	addDoPingtTool(s)
	addGetMIBTreeTool(s)
	addSNMPWalkTool(s)
	addAddNodeTool(s)
	addUpdateNodeTool(s)
	mcpSSEServer = server.NewSSEServer(s)
	e.Any("/sse", func(c echo.Context) error {
		if _, ok := mcpAllow.Load(c.RealIP()); !ok {
			return echo.ErrUnauthorized
		}
		mcpSSEServer.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})
	e.Any("/message", func(c echo.Context) error {
		if _, ok := mcpAllow.Load(c.RealIP()); !ok {
			return echo.ErrUnauthorized
		}
		mcpSSEServer.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})
	mcpStreamableHTTPServer = server.NewStreamableHTTPServer(s)
	e.Any("/mcp", func(c echo.Context) error {
		if _, ok := mcpAllow.Load(c.RealIP()); !ok {
			return echo.ErrUnauthorized
		}
		mcpStreamableHTTPServer.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})
}

func stopMCPServer(ctx context.Context) {
	if mcpSSEServer != nil {
		mcpSSEServer.Shutdown(ctx)
	}
	if mcpStreamableHTTPServer != nil {
		mcpStreamableHTTPServer.Shutdown(ctx)
	}
}

// get_node_list tool
type mcpNodeEnt struct {
	ID         string
	Name       string
	IP         string
	MAC        string
	State      string
	X          int
	Y          int
	Icon       string
	Descrption string
}

func addGetNodeListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_node_list",
		mcp.WithDescription("get node list from TWSNMP"),
		mcp.WithString("name_filter",
			mcp.Description("node name filter. Empty is no filter"),
		),
		mcp.WithString("ip_filter",
			mcp.Description("node ip filter. Empty is no filter"),
		),
		mcp.WithString("state_filter",
			mcp.Enum("", "normal", "warn", "low", "high", "repair"),
			mcp.Description(
				`node state filter. Empty is no filter
 select state name.(normal,warn,low,high,repair)
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name_filter", "")
		ip := request.GetString("ip_filter", "")
		state := request.GetString("state_filter", "")
		list := []mcpNodeEnt{}
		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			if name != "" && name != n.Name {
				return true
			}
			if ip != "" && ip != n.IP {
				return true
			}
			if state != "" && state != n.State {
				return true
			}
			list = append(list, mcpNodeEnt{
				ID:         n.ID,
				Name:       n.Name,
				IP:         n.IP,
				MAC:        n.MAC,
				X:          n.X,
				Y:          n.Y,
				Icon:       n.Icon,
				Descrption: n.Descr,
				State:      n.State,
			})
			return true
		})
		j, err := json.Marshal(&list)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// get_network_list tool
type mcpNetworkEnt struct {
	ID         string
	Name       string
	IP         string
	Ports      []string
	X          int
	Y          int
	Descrption string
}

func addGetNetworkListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_network_list",
		mcp.WithDescription("get network list from TWSNMP"),
		mcp.WithString("name_filter",
			mcp.Description("network name filter. Empty is no filter"),
		),
		mcp.WithString("ip_filter",
			mcp.Description("network ip filter. Empty is no filter"),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name_filter", "")
		ip := request.GetString("ip_filter", "")
		list := []mcpNetworkEnt{}
		datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
			if name != "" && name != n.Name {
				return true
			}
			if ip != "" && ip != n.IP {
				return true
			}
			ports := []string{}
			for _, p := range n.Ports {
				ports = append(ports, fmt.Sprintf("%s=%s", p.Name, p.State))
			}
			list = append(list, mcpNetworkEnt{
				ID:         n.ID,
				Name:       n.Name,
				IP:         n.IP,
				X:          n.X,
				Y:          n.Y,
				Descrption: n.Descr,
				Ports:      ports,
			})
			return true
		})
		j, err := json.Marshal(&list)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// get_polling_list tool
type mcpPollingEnt struct {
	ID       string
	Name     string
	NodeID   string
	NodeName string
	Type     string
	Level    string
	State    string
	LastTime string
	Result   map[string]any
}

func addGetPollingListTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("get_polling_list",
		mcp.WithDescription("get polling list from TWSNMP"),
		mcp.WithString("type_filter",
			mcp.Enum("", "ping", "snmp", "tcp", "http", "dns"),
			mcp.Description("polling type filter. Empty is no filter"),
		),
		mcp.WithString("state_filter",
			mcp.Enum("", "normal", "warn", "low", "high", "repair"),
			mcp.Description(
				`node state filter. Empty is no filter
 select state name.(normal,warn,low,high,repair)
`),
		),
		mcp.WithString("name_filter",
			mcp.Description("polling name filter. Empty is no filter"),
		),
		mcp.WithString("node_name_filter",
			mcp.Description("node name filter. Empty is no filter"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name_filter", "")
		nodeName := request.GetString("node_name_filter", "")
		typeFilter := request.GetString("type_filter", "")
		state := request.GetString("state_filter", "")
		list := []mcpPollingEnt{}
		datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
			if name != "" && name != p.Name {
				return true
			}
			if typeFilter != "" && typeFilter != p.Type {
				return true
			}
			n := datastore.GetNode(p.NodeID)
			if n == nil {
				return true
			}
			if nodeName != "" && nodeName != n.Name {
				return true
			}
			if state != "" && state != p.State {
				return true
			}
			list = append(list, mcpPollingEnt{
				ID:       p.ID,
				Name:     p.Name,
				NodeName: n.Name,
				LastTime: time.Unix(0, p.LastTime).Format(time.RFC3339Nano),
				State:    n.State,
				Result:   p.Result,
			})
			return true
		})
		j, err := json.Marshal(&list)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// do_ping tool
type mcpPingEnt struct {
	Result       string `json:"Result"`
	Time         string `json:"Time"`
	RTT          string `json:"RTT"`
	RTTNano      int64  `json:"RTTNano"`
	Size         int    `json:"Size"`
	TTL          int    `json:"TTL"`
	ResponceFrom string `json:"ResponceFrom"`
	Location     string `json:"Location"`
}

func addDoPingtTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("do_ping",
		mcp.WithDescription("do ping"),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("ping target ip address or host name"),
		),
		mcp.WithNumber("size",
			mcp.DefaultNumber(64),
			mcp.Max(1500),
			mcp.Min(64),
			mcp.Description("ping packate size"),
		),
		mcp.WithNumber("ttl",
			mcp.DefaultNumber(254),
			mcp.Max(254),
			mcp.Min(1),
			mcp.Description("ip packet TTL"),
		),
		mcp.WithNumber("timeout",
			mcp.DefaultNumber(2),
			mcp.Max(10),
			mcp.Min(1),
			mcp.Description("timeout sec of ping"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		target, err := request.RequireString("target")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		target = getTragetIP(target)
		if target == "" {
			return mcp.NewToolResultText("target ip not found"), nil
		}
		timeout := request.GetInt("timeout", 3)
		size := request.GetInt("size", 64)
		ttl := request.GetInt("ttl", 254)
		pe := ping.DoPing(target, timeout, 0, size, ttl)
		res := mcpPingEnt{
			Result:       pe.Stat.String(),
			Time:         time.Now().Format(time.RFC3339),
			RTT:          time.Duration(pe.Time).String(),
			Size:         pe.Size,
			ResponceFrom: pe.RecvSrc,
			TTL:          pe.RecvTTL,
			RTTNano:      pe.Time,
		}
		if pe.RecvSrc != "" {
			res.Location = datastore.GetLoc(pe.RecvSrc)
		}
		j, err := json.Marshal(&res)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// getTragetIP: targetからIPアドレスを取得する、targetはノード名、ホスト名、IPアドレス
func getTragetIP(target string) string {
	ipreg := regexp.MustCompile(`^[0-9.]+$`)
	if ipreg.MatchString(target) {
		return target
	}
	n := datastore.FindNodeFromName(target)
	if n != nil {
		return n.IP
	}
	if ips, err := net.LookupIP(target); err == nil {
		for _, ip := range ips {
			if ip.IsGlobalUnicast() {
				s := ip.To4().String()
				if ipreg.MatchString(s) {
					return s
				}
			}
		}
	}
	return ""
}

func addGetMIBTreeTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("get_MIB_tree",
		mcp.WithDescription("get MIB tree from TWSNMP"),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		j, err := json.Marshal(&datastore.MIBTree)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

type mcpMIBEnt struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

func addSNMPWalkTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("snmpwalk",
		mcp.WithDescription("SNMP walk tool"),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("snmpwalk target ip address or host name or node name"),
		),
		mcp.WithString("mib_object_name",
			mcp.Required(),
			mcp.Description("snmpwak mib object name"),
		),
		mcp.WithString("community",
			mcp.Description("snmp v2c comminity name"),
		),
		mcp.WithString("user",
			mcp.Description("snmp v3 user name"),
		),
		mcp.WithString("password",
			mcp.Description("snmp v3 password"),
		),
		mcp.WithString("snmpmode",
			mcp.Enum("", "v2c", "v3auth", "v3authpriv", "v3authprivex"),
			mcp.Description(
				`snmp mode
v2c : SNMP v2 (default)
v3auth: SNMP v3 authentication protocol is SHA1,privacy protocol is none.
v3authpriv: SNMP v3 authentication protocol is SHA1,privacy protocol is AES.
v3authprivex: SNMP v3 authentication protocol is SHA1,privacy protocol is AES256.
`),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		community := request.GetString("community", "")
		user := request.GetString("community", "")
		password := request.GetString("password", "")
		snmpMode := request.GetString("snmpmode", "")
		name, err := request.RequireString("mib_object_name")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		target, err := request.RequireString("target")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		if n := datastore.FindNodeFromName(target); n != nil {
			if community == "" {
				community = n.Community
			}
			if user == "" {
				user = n.User
			}
			if password == "" {
				password = n.Password
			}
			if snmpMode == "" {
				snmpMode = n.SnmpMode
			}
			target = n.IP
		} else {
			target = getTragetIP(target)
			if target == "" {
				return mcp.NewToolResultText("target ip not found"), nil
			}
		}
		agent := &gosnmp.GoSNMP{
			Target:    target,
			Port:      161,
			Transport: "udp",
			Community: community,
			Version:   gosnmp.Version2c,
			Timeout:   time.Duration(datastore.MapConf.Timeout) * time.Second,
			Retries:   datastore.MapConf.Retry,
			MaxOids:   gosnmp.MaxOids,
		}
		switch snmpMode {
		case "v3auth":
			agent.Version = gosnmp.Version3
			agent.SecurityModel = gosnmp.UserSecurityModel
			agent.MsgFlags = gosnmp.AuthNoPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 user,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: password,
			}
		case "v3authpriv":
			agent.Version = gosnmp.Version3
			agent.SecurityModel = gosnmp.UserSecurityModel
			agent.MsgFlags = gosnmp.AuthPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 user,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: password,
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        password,
			}
		case "v3authprivex":
			agent.Version = gosnmp.Version3
			agent.SecurityModel = gosnmp.UserSecurityModel
			agent.MsgFlags = gosnmp.AuthPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 user,
				AuthenticationProtocol:   gosnmp.SHA256,
				AuthenticationPassphrase: password,
				PrivacyProtocol:          gosnmp.AES256,
				PrivacyPassphrase:        password,
			}
		}
		res := []mcpMIBEnt{}
		err = agent.Connect()
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		defer agent.Conn.Close()
		err = agent.Walk(nameToOID(name), func(variable gosnmp.SnmpPDU) error {
			name := datastore.MIBDB.OIDToName(variable.Name)
			value := datastore.GetMIBValueString(name, &variable, false)
			res = append(res, mcpMIBEnt{
				Name:  name,
				Value: value,
			})
			return nil
		})
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		j, err := json.Marshal(&res)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// add_node tool
func addAddNodeTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("add_node",
		mcp.WithDescription("add node to TWSNMP"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("node name"),
		),
		mcp.WithString("ip",
			mcp.Required(),
			mcp.Description("node ip address"),
		),
		mcp.WithString("icon",
			mcp.Description("icon of node"),
		),
		mcp.WithString("description",
			mcp.Description("description of node"),
		),
		mcp.WithNumber("x",
			mcp.Max(1000),
			mcp.Min(64),
			mcp.Description("x positon of node"),
		),
		mcp.WithNumber("y",
			mcp.Max(1000),
			mcp.Min(64),
			mcp.Description("y positon of node"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		icon := request.GetString("icon", "desktop")
		descr := request.GetString("description", "")
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		ip, err := request.RequireString("ip")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		x := request.GetInt("x", 64)
		y := request.GetInt("y", 64)
		n := &datastore.NodeEnt{
			Name:  name,
			IP:    ip,
			Icon:  icon,
			X:     x,
			Y:     y,
			Descr: descr,
			State: "unknown",
		}
		if err := datastore.AddNode(n); err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		datastore.AddPolling(&datastore.PollingEnt{
			Name:   "PING",
			Type:   "ping",
			NodeID: n.ID})
		j, err := json.Marshal(&mcpNodeEnt{
			ID:         n.ID,
			Name:       n.Name,
			Descrption: n.Descr,
			IP:         n.IP,
			State:      n.State,
			X:          n.X,
			Y:          n.Y,
			Icon:       n.Icon,
		})
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// update_node
func addUpdateNodeTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("update_node",
		mcp.WithDescription("update node name,ip, positon,description or icon"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("node id to update"),
		),
		mcp.WithString("name",
			mcp.Description("node name"),
		),
		mcp.WithString("ip",
			mcp.Description("node ip address"),
		),
		mcp.WithString("icon",
			mcp.Description("icon of node"),
		),
		mcp.WithString("description",
			mcp.Description("description of node"),
		),
		mcp.WithNumber("x",
			mcp.Description("x positon of node"),
		),
		mcp.WithNumber("y",
			mcp.Description("y positon of node"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		icon := request.GetString("icon", "")
		descr := request.GetString("description", "")
		name := request.GetString("name", "")
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		x := request.GetInt("x", -1)
		y := request.GetInt("y", -1)
		n := datastore.GetNode(id)
		if n == nil {
			return mcp.NewToolResultText("node not found"), nil
		}
		if x >= 0 {
			n.X = x
		}
		if y >= 0 {
			n.Y = y
		}
		if icon != "" {
			n.Icon = icon
		}
		if descr != "" {
			n.Descr = descr
		}
		if name != "" {
			n.Name = name
		}
		j, err := json.Marshal(&mcpNodeEnt{
			ID:         n.ID,
			Name:       n.Name,
			Descrption: n.Descr,
			IP:         n.IP,
			State:      n.State,
			X:          n.X,
			Y:          n.Y,
			Icon:       n.Icon,
		})
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}
