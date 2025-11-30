package webapi

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var mcpAllow sync.Map

func startMCPServer(e *echo.Echo, p *WebAPI) {
	log.Printf("start mcp server mode=%s,form=%s", p.MCPMode, p.MCPFrom)
	mcpAllow.Store("127.0.0.1", true)
	mcpAllow.Store("::1", true)
	for _, ip := range strings.Split(p.MCPFrom, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			mcpAllow.Store(ip, true)
		}
	}
	// Create MCP Server
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "TWSNMP FC MCP Server",
			Version: p.Version,
		},
		nil)

	// Add tools to MCP server
	addMCPTools(s)
	// Add prompts to MCP server
	addMCPPrompts(s)
	switch p.MCPMode {
	case "sse":
		handler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
			return s
		}, nil)

		e.Any("/sse", func(c echo.Context) error {
			if !mcpCheckFromAddress(c) {
				return echo.ErrUnauthorized
			}
			handler.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/message", func(c echo.Context) error {
			if !mcpCheckFromAddress(c) {
				return echo.ErrUnauthorized
			}
			handler.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
	case "auth":
		handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
			return s
		}, nil)
		e.Any("/mcp", func(c echo.Context) error {
			if !mcpCheckFromAddress(c) {
				return echo.ErrUnauthorized
			}
			handler.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		}, echojwt.JWT([]byte(p.Password)))
	default:
		handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
			return s
		}, nil)
		e.Any("/mcp", func(c echo.Context) error {
			if !mcpCheckFromAddress(c) {
				return echo.ErrUnauthorized
			}
			handler.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
	}
}

// makeRegexFilter
func makeRegexFilter(s string) *regexp.Regexp {
	if s != "" {
		if f, err := regexp.Compile(s); err == nil && f != nil {
			return f
		}
	}
	return nil
}

// check from address
func mcpCheckFromAddress(c echo.Context) bool {
	if ip, _, err := net.SplitHostPort(c.Request().RemoteAddr); err == nil {
		if _, ok := mcpAllow.Load(ip); ok {
			return true
		}
	}
	if _, ok := mcpAllow.Load(c.RealIP()); ok {
		return true
	}
	return false
}

func addMCPTools(s *mcp.Server) {
	// mcp_map
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_list",
		Description: "get node list from TWSNMP",
	}, mcpGetNodeList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_network_list",
		Description: "get network list from TWSNMP",
	}, mcpGetNetworkList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_polling_list",
		Description: "get polling list from TWSNMP",
	}, mcpGetPollingList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_polling_log",
		Description: "get polling log from TWSNMP",
	}, mcpGetPollingLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_polling_log_data",
		Description: "get polling log data from TWSNMP",
	}, mcpGetPollingLogData)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "do_ping",
		Description: "do ping",
	}, mcpDoPing)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_mib_tree",
		Description: "get MIB tree from TWSNMP",
	}, mcpGetMIBTree)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "snmpwalk",
		Description: "SNMP walk tool",
	}, mcpSnmpWalk)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "snmpset",
		Description: "SNMP set tool",
	}, mcpSnmpSet)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_node",
		Description: "add node to TWSNMP.A PING polling is also added automatically.",
	}, mcpAddNode)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_node",
		Description: "update node name,ip, position,description or icon",
	}, mcpUpdateNode)

	// mcp_report
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_sensor_list",
		Description: "get sensor list from TWSNMP",
	}, mcpGetSensorList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_mac_address_list",
		Description: "get MAC address list from TWSNMP",
	}, mcpGetMACAddressList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_ip_address_list",
		Description: "get IP address list from TWSNMP",
	}, mcpGetIPAddressList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_wifi_ap_list",
		Description: "get Wifi access point list from TWSNMP",
	}, mcpGetWifiAPList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_bluetooth_device_list",
		Description: "get bluetooth device list from TWSNMP",
	}, mcpGetBluetoothDeviceList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_server_certificate_list",
		Description: "get server certificate list from TWSNMP",
	}, mcpGetServerCertificateList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_resource_monitor_list",
		Description: "get resource monitor list from TWSNMP",
	}, mcpGetResourceMonitorList)

	// mcp_log
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_event_log",
		Description: "search event log from TWSNMP",
	}, mcpSearchEventLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_syslog",
		Description: "search syslog from TWSNMP",
	}, mcpSearchSyslog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_syslog_summary",
		Description: "get syslog summary from TWSNMP",
	}, mcpGetSyslogSummary)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_snmp_trap_log",
		Description: "search SNMP trap log from TWSNMP",
	}, mcpSearchSnmpTrapLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_event_log",
		Description: "add event log to TWSNMP",
	}, mcpAddEventLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_ip_address_info",
		Description: "get ip address info.(DNS host,Managed node,Geo location,RDAP)",
	}, mcpGetIPInfo)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_mac_address_info",
		Description: "get mac address info.(IP,Managed node,Vendor)",
	}, mcpGetMACInfo)
}

// Add prompts
func addMCPPrompts(s *mcp.Server) {
	// get_node
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_node_list",
		Title:       "Get node list with filters",
		Description: "Get a list of nodes registered in TWSNMP with filters.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "state_filter",
				Title:       "node state filter",
				Description: "node state filter. state is normal,repair,warn,low,high,unknown.",
				Required:    false,
			},
			{
				Name:        "name_filter",
				Title:       "node name filter",
				Description: "node name filter",
				Required:    false,
			},
			{
				Name:        "ip_filter",
				Title:       "node ip address filter",
				Description: "node ip address filter",
				Required:    false,
			},
		},
	}, getNodeListPrompt)

	// add_node
	s.AddPrompt(&mcp.Prompt{
		Name:        "add_node",
		Title:       "Add a new node to TWSNMP.",
		Description: "Add a new node to TWSNMP.A PING polling is also added automatically.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "name",
				Title:       "Name of new node",
				Description: "Name of new node",
				Required:    true,
			},
			{
				Name:        "ip",
				Title:       "IP address",
				Description: "IP address",
				Required:    true,
			},
			{
				Name:        "icon",
				Title:       "Icon",
				Description: "Icon. (desktop,laptop,server,cloud,router,ip)",
				Required:    false,
			},
			{
				Name:        "description",
				Title:       "Description of new node",
				Description: "Description of new node",
				Required:    false,
			},
			{
				Name:        "position",
				Title:       "Position of new node",
				Description: "Position of new node.(ex. x=100,y=200)",
				Required:    false,
			},
		},
	}, addNodePrompt)

	// update_node
	s.AddPrompt(&mcp.Prompt{
		Name:        "update_node",
		Title:       "Update node on TWSNMP.",
		Description: "Update node on TWSNMP.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "id",
				Title:       "ID of node to update",
				Description: "ID of node to update. ID or name,IP address.",
				Required:    true,
			},
			{
				Name:        "name",
				Title:       "New name of node",
				Description: "New name of node",
				Required:    false,
			},
			{
				Name:        "ip",
				Title:       "New IP address of node",
				Description: "New IP address of node",
				Required:    false,
			},
			{
				Name:        "icon",
				Title:       "New icon of node",
				Description: "new Icon of node. (desktop,laptop,server,cloud,router,ip)",
				Required:    false,
			},
			{
				Name:        "description",
				Title:       "New description of node",
				Description: "New description of node",
				Required:    false,
			},
			{
				Name:        "position",
				Title:       "new position of node",
				Description: "new position of node.(ex. x=100,y=200)",
				Required:    false,
			},
		},
	}, updateNodePrompt)

	// get_network_list
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_network_list",
		Title:       "Get network node list with filters",
		Description: "Get a list of network nodes registered in TWSNMP with filters.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "name_filter",
				Title:       "network node name filter",
				Description: "network node name filter",
				Required:    false,
			},
			{
				Name:        "ip_filter",
				Title:       "network node ip address filter",
				Description: "network node ip address filter",
				Required:    false,
			},
		},
	}, getNetworkListPrompt)

	// get_polling_list
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_polling_list",
		Title:       "Get polling list with filters",
		Description: "Get a list of polling registered in TWSNMP with filters.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "type_filter",
				Title:       "Polling type filter",
				Description: "Polling type filter.(ping,snmp,syslog,http,tcp)",
				Required:    false,
			},
			{
				Name:        "name_filter",
				Title:       "polling name filter",
				Description: "polling name filter",
				Required:    false,
			},
			{
				Name:        "node_name_filter",
				Title:       "node name filter",
				Description: "node name filter",
				Required:    false,
			},
			{
				Name:        "state_filter",
				Title:       "polling state filter.",
				Description: "polling state filter. state is normal,repair,warn,low,high or unknown.",
				Required:    false,
			},
		},
	}, getPollingListPrompt)

	// get_polling_log
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_polling_log",
		Title:       "Get polling log.",
		Description: "Get polling log from TWSNMP",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "id",
				Title:       "Polling ID",
				Description: "Polling ID",
				Required:    true,
			},
			{
				Name:        "limit",
				Title:       "Max number of logs to get.",
				Description: "Max number of logs to get.(default 100)",
				Required:    false,
			},
		},
	}, getPollingLogPrompt)

	// do_ping
	s.AddPrompt(&mcp.Prompt{
		Name:        "do_ping",
		Title:       "Do ping",
		Description: "Do ping to target from TWSNMP.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "target",
				Title:       "Ping target",
				Description: "Ping target(ip,node name or host name)",
				Required:    true,
			},
			{
				Name:        "size",
				Title:       "Ping packet size.",
				Description: "Ping packet size.(default 64)",
				Required:    false,
			},
			{
				Name:        "ttl",
				Title:       "Ping packet TTL.",
				Description: "Ping packet TTL.(default 254)",
				Required:    false,
			},
			{
				Name:        "timeout",
				Title:       "Ping timeout",
				Description: "Ping timeout(sec).(default 3)",
				Required:    false,
			},
		},
	}, doPingPrompt)

	// snmpwalk
	s.AddPrompt(&mcp.Prompt{
		Name:        "snmpwalk",
		Title:       "Do snmpwalk",
		Description: "Do snmpwalk to target from TWSNMP.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "target",
				Title:       "SNMP walk target",
				Description: "SNMP walk target(ip,node name or host name)",
				Required:    true,
			},
			{
				Name:        "mib_object_name",
				Title:       "MIB object name",
				Description: "MIB object name to get",
				Required:    true,
			},
			{
				Name:        "snmp_mode",
				Title:       "SNMP mode",
				Description: "SNMP mode(v2c,v3auth,v3authpriv,v3authprivex)",
				Required:    false,
			},
			{
				Name:        "community",
				Title:       "Community name",
				Description: "Community name for v2c mode.",
				Required:    false,
			},
			{
				Name:        "user",
				Title:       "User name",
				Description: "User name for v3 mode.",
				Required:    false,
			},
			{
				Name:        "password",
				Title:       "Password",
				Description: "Password for v3 mode.",
				Required:    false,
			},
		},
	}, doSnmpWalkPrompt)

	// search_event_log
	s.AddPrompt(&mcp.Prompt{
		Name:        "search_event_log",
		Title:       "Search event log with filters",
		Description: "Search event log with filters",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "type_filter",
				Title:       "Event type filter",
				Description: "Event type filter.(system,polling,arpwatch,mcp)",
				Required:    false,
			},
			{
				Name:        "node_filter",
				Title:       "Node filter",
				Description: "Node filter",
				Required:    false,
			},
			{
				Name:        "level_name_filter",
				Title:       "Level filter",
				Description: "Level filter.(info,normal,warn,low,high)",
				Required:    false,
			},
			{
				Name:        "event_filter",
				Title:       "Event filter.",
				Description: "Event filter.",
				Required:    false,
			},
			{
				Name:        "start_time",
				Title:       "Start time",
				Description: "Start time of search.(default: -1h)",
				Required:    false,
			},
			{
				Name:        "end_time",
				Title:       "End time",
				Description: "End date and time. (default: now)",
				Required:    false,
			},
			{
				Name:        "limit",
				Title:       "Limit",
				Description: "Max number of logs to search.",
				Required:    false,
			},
		},
	}, searchEventLogPrompt)

	// search_syslog
	s.AddPrompt(&mcp.Prompt{
		Name:        "search_syslog",
		Title:       "Search syslog with filters",
		Description: "Search syslog with filters",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "level_filter",
				Title:       "Level filter",
				Description: "Level filter.(warn,low,high,debug,info)",
				Required:    false,
			},
			{
				Name:        "host_filter",
				Title:       "Sender host filter",
				Description: "Sender host filter",
				Required:    false,
			},
			{
				Name:        "tag_filter",
				Title:       "Syslog tag filter",
				Description: "Syslog tag filter",
				Required:    false,
			},
			{
				Name:        "message_filter",
				Title:       "Syslog message filter.",
				Description: "Syslog message filter.",
				Required:    false,
			},
			{
				Name:        "start_time",
				Title:       "Start time",
				Description: "Start time of search.(default: -1h)",
				Required:    false,
			},
			{
				Name:        "end_time",
				Title:       "End time",
				Description: "End date and time. (default: now)",
				Required:    false,
			},
			{
				Name:        "limit",
				Title:       "Limit",
				Description: "Max number of syslogs to search.",
				Required:    false,
			},
		},
	}, searchSyslogPrompt)

	// get_syslog_summary
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_syslog_summary",
		Title:       "Get syslog summary with filters",
		Description: "Get syslog summary with filters",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "level_filter",
				Title:       "Level filter",
				Description: "Level filter.(warn,low,high,debug,info)",
				Required:    false,
			},
			{
				Name:        "host_filter",
				Title:       "Sender host filter",
				Description: "Sender host filter",
				Required:    false,
			},
			{
				Name:        "tag_filter",
				Title:       "Syslog tag filter",
				Description: "Syslog tag filter",
				Required:    false,
			},
			{
				Name:        "message_filter",
				Title:       "Syslog message filter.",
				Description: "Syslog message filter.",
				Required:    false,
			},
			{
				Name:        "start_time",
				Title:       "Start time",
				Description: "Start time of search.(default: -1h)",
				Required:    false,
			},
			{
				Name:        "end_time",
				Title:       "End time",
				Description: "End date and time. (default: now)",
				Required:    false,
			},
			{
				Name:        "top_n",
				Title:       "Top N",
				Description: "Number of top syslog summary.",
				Required:    false,
			},
		},
	}, getSyslogSummaryPrompt)

	// search_snmp_trap_log
	s.AddPrompt(&mcp.Prompt{
		Name:        "search_snmp_trap_log",
		Title:       "Search snmp trap log with filters",
		Description: "Search snmp trap log of TWSNMP with filters",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "from_filter",
				Title:       "Trap from filter",
				Description: "Trap from filter.",
				Required:    false,
			},
			{
				Name:        "trap_type_filter",
				Title:       "Trap type filter",
				Description: "Trap type filter",
				Required:    false,
			},
			{
				Name:        "variable_filter",
				Title:       "Variable of trap filter",
				Description: "Variable of trap filter",
				Required:    false,
			},
			{
				Name:        "start_time",
				Title:       "Start time",
				Description: "Start time of search.(default: -1h)",
				Required:    false,
			},
			{
				Name:        "end_time",
				Title:       "End time",
				Description: "End date and time. (default: now)",
				Required:    false,
			},
			{
				Name:        "limit",
				Title:       "Limit",
				Description: "Number of snmp trap log to search.",
				Required:    false,
			},
		},
	}, searchSnmpTrapLogPrompt)

	// get_mib_tree
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_mib_tree",
		Title:       "Get MIB tree of TWSNMP.",
		Description: "Get MIB tree of TWSNMP by using get_mib_tree tool",
	}, getMIBTreePrompt)

	// get_ip_address_list
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_ip_address_list",
		Title:       "Get the list of IP address managed by TWSNMP.",
		Description: "Get the list of IP address managed by TWSNMP by using get_ip_address_list tool",
	}, getIPAddressListPrompt)

	// get_resource_monitor_list
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_resource_monitor_list",
		Title:       "Get resource monitor info of TWSNMP",
		Description: "Get resource monitor info of TWSNMP by using get_resource_monitor_list tool",
	}, getResourceMonitorPrompt)

	// get_server_certificate_list
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_server_certificate_list",
		Title:       "Get the list of server certificates managed by TWSNMP",
		Description: "Get the list of server certificates managed by TWSNMP by using get_server_certificate_list tool",
	}, getServerCertificateListPrompt)

	// add_event_log
	s.AddPrompt(&mcp.Prompt{
		Name:        "add_event_log",
		Title:       "Add Event log to TWSNMP",
		Description: "Add Event log to TWSNMP.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "level",
				Title:       "Level of event log",
				Description: "Level of event log(info,normal,warn,low,high)",
				Required:    true,
			},
			{
				Name:        "node",
				Title:       "Node name of event log",
				Description: "Node name of event log",
				Required:    false,
			},
			{
				Name:        "event",
				Title:       "Event log content",
				Description: "Event log content",
				Required:    true,
			},
		},
	}, addEventLogPrompt)

	// get_ip_address_info
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_ip_address_info",
		Title:       "Get IP address information",
		Description: "Get IP address information.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "ip",
				Title:       "IP address",
				Description: "IP address",
				Required:    true,
			},
		},
	}, getIPAddressInfoPrompt)

	// get_mac_address_info
	s.AddPrompt(&mcp.Prompt{
		Name:        "get_mac_address_info",
		Title:       "Get MAC address information",
		Description: "Get MAC address information.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "mac",
				Title:       "MAC address",
				Description: "MAC address",
				Required:    true,
			},
		},
	}, getMACAddressInfoPrompt)
}

// Prompts
func getNodeListPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if state, ok := req.Params.Arguments["state_filter"]; ok {
		c = append(c, fmt.Sprintf("- state: %s", state))
	}
	if ip, ok := req.Params.Arguments["ip_filter"]; ok {
		c = append(c, fmt.Sprintf("- IP address: %s", ip))
	}
	if name, ok := req.Params.Arguments["name_filter"]; ok {
		c = append(c, fmt.Sprintf("- Name: %s", name))
	}
	p := "Get a list of nodes registered in TWSNMP by using get_node_list tool"
	if len(c) > 0 {
		p = " with following filter.\n" + strings.Join(c, "\n")
	} else {
		p += "."
	}
	return &mcp.GetPromptResult{
		Description: "get node list prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func addNodePrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if name, ok := req.Params.Arguments["name"]; ok {
		c = append(c, fmt.Sprintf("- Name: %s", name))
	} else {
		return nil, fmt.Errorf("name is required")
	}
	if ip, ok := req.Params.Arguments["ip"]; ok {
		c = append(c, fmt.Sprintf("- IP address: %s", ip))
	} else {
		return nil, fmt.Errorf("ip is required")
	}
	if icon, ok := req.Params.Arguments["icon"]; ok {
		c = append(c, fmt.Sprintf("- Icon: %s", icon))
	}
	if description, ok := req.Params.Arguments["description"]; ok {
		c = append(c, fmt.Sprintf("- Description: %s", description))
	}
	if position, ok := req.Params.Arguments["position"]; ok {
		c = append(c, fmt.Sprintf("- Position: %s", position))
	}
	p := "Add a new node to TWSNMP by using add_node tool with the following information:\n" + strings.Join(c, "\n")
	return &mcp.GetPromptResult{
		Description: "Add a new node to TWSNMP prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func updateNodePrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if id, ok := req.Params.Arguments["id"]; !ok {
		return nil, fmt.Errorf("id is required")
	} else {
		c = append(c, fmt.Sprintf("- ID: %s", id))
	}
	if name, ok := req.Params.Arguments["name"]; ok {
		c = append(c, fmt.Sprintf("- Name: %s", name))
	}
	if ip, ok := req.Params.Arguments["ip"]; ok {
		c = append(c, fmt.Sprintf("- IP address: %s", ip))
	}
	if icon, ok := req.Params.Arguments["icon"]; ok {
		c = append(c, fmt.Sprintf("- Icon: %s", icon))
	}
	if description, ok := req.Params.Arguments["description"]; ok {
		c = append(c, fmt.Sprintf("- Description: %s", description))
	}
	if position, ok := req.Params.Arguments["position"]; ok {
		c = append(c, fmt.Sprintf("- Position: %s", position))
	}
	p := "Update the node on TWSNMP by using update_node tool with the following information:\n" + strings.Join(c, "\n")
	return &mcp.GetPromptResult{
		Description: "Update the node prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func getNetworkListPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if ip, ok := req.Params.Arguments["ip_filter"]; ok {
		c = append(c, fmt.Sprintf("- IP address: %s", ip))
	}
	if name, ok := req.Params.Arguments["name_filter"]; ok {
		c = append(c, fmt.Sprintf("- Name: %s", name))
	}
	p := "Get a list of network nodes registered in TWSNMP by using get_network_list tool"
	if len(c) > 0 {
		p = " with following filter.\n" + strings.Join(c, "\n")
	} else {
		p += "."
	}
	return &mcp.GetPromptResult{
		Description: "get network node list prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func getPollingListPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if t, ok := req.Params.Arguments["type_filter"]; ok {
		c = append(c, fmt.Sprintf("- Type: %s", t))
	}
	if name, ok := req.Params.Arguments["name_filter"]; ok {
		c = append(c, fmt.Sprintf("- Name: %s", name))
	}
	if node, ok := req.Params.Arguments["node_name_filter"]; ok {
		c = append(c, fmt.Sprintf("- Node name: %s", node))
	}
	if name, ok := req.Params.Arguments["state_filter"]; ok {
		c = append(c, fmt.Sprintf("- State: %s", name))
	}
	p := "Get a list of polling registered in TWSNMP by using get_polling_list tool"
	if len(c) > 0 {
		p = " with following filter.\n" + strings.Join(c, "\n")
	} else {
		p += "."
	}
	return &mcp.GetPromptResult{
		Description: "get polling list prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func getPollingLogPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if id, ok := req.Params.Arguments["id"]; !ok {
		return nil, fmt.Errorf("id is required")
	} else {
		c = append(c, fmt.Sprintf("- ID: %s", id))
	}
	if limit, ok := req.Params.Arguments["name"]; ok {
		c = append(c, fmt.Sprintf("- Limit: %s", limit))
	}
	p := "Get polling log by using get_polling_log tool with the following conditions:\n" + strings.Join(c, "\n")
	return &mcp.GetPromptResult{
		Description: "Get polling log prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func doPingPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if target, ok := req.Params.Arguments["target"]; !ok {
		return nil, fmt.Errorf("target is required")
	} else {
		c = append(c, fmt.Sprintf("- Target: %s", target))
	}
	if size, ok := req.Params.Arguments["size"]; ok {
		c = append(c, fmt.Sprintf("- Size: %s", size))
	}
	if ttl, ok := req.Params.Arguments["ttl"]; ok {
		c = append(c, fmt.Sprintf("- TTL: %s", ttl))
	}
	if timeout, ok := req.Params.Arguments["timeout"]; ok {
		c = append(c, fmt.Sprintf("- Timeout: %s", timeout))
	}
	p := "Do ping to target by using do_ping tool with the following conditions:\n" + strings.Join(c, "\n")
	return &mcp.GetPromptResult{
		Description: "Do ping prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func doSnmpWalkPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if target, ok := req.Params.Arguments["target"]; !ok {
		return nil, fmt.Errorf("target is required")
	} else {
		c = append(c, fmt.Sprintf("- Target: %s", target))
	}
	if target, ok := req.Params.Arguments["mib_object_name"]; !ok {
		return nil, fmt.Errorf("mib object name is required")
	} else {
		c = append(c, fmt.Sprintf("- MIB object name: %s", target))
	}
	if mode, ok := req.Params.Arguments["snmp_mode"]; ok {
		c = append(c, fmt.Sprintf("- SNMP Mode: %s", mode))
	}
	if community, ok := req.Params.Arguments["community"]; ok {
		c = append(c, fmt.Sprintf("- Community: %s", community))
	}
	if user, ok := req.Params.Arguments["user"]; ok {
		c = append(c, fmt.Sprintf("- User: %s", user))
	}
	if password, ok := req.Params.Arguments["password"]; ok {
		c = append(c, fmt.Sprintf("- User: %s", password))
	}
	p := "Do snmpwalk to target by using snmpwalk tool with the following conditions:\n" + strings.Join(c, "\n")
	return &mcp.GetPromptResult{
		Description: "snmpwalk prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func searchEventLogPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if t, ok := req.Params.Arguments["type_filter"]; ok {
		c = append(c, fmt.Sprintf("- Type filter: %s", t))
	}
	if node, ok := req.Params.Arguments["node_filter"]; ok {
		c = append(c, fmt.Sprintf("- Node filter: %s", node))
	}
	if level, ok := req.Params.Arguments["level_filter"]; ok {
		c = append(c, fmt.Sprintf("- Level filter: %s", level))
	}
	if event, ok := req.Params.Arguments["event_filter"]; ok {
		c = append(c, fmt.Sprintf("- Event filter: %s", event))
	}
	if st, ok := req.Params.Arguments["start_time"]; ok {
		c = append(c, fmt.Sprintf("- Start Time: %s", st))
	}
	if et, ok := req.Params.Arguments["end_time"]; ok {
		c = append(c, fmt.Sprintf("- End Time: %s", et))
	}
	if limit, ok := req.Params.Arguments["limit"]; ok {
		c = append(c, fmt.Sprintf("- Limit: %s", limit))
	}
	p := "Search event log of TWSNMP by using search_event_log tool"
	if len(c) > 0 {
		p = " with following filter.\n" + strings.Join(c, "\n")
	} else {
		p += "."
	}
	return &mcp.GetPromptResult{
		Description: "Search event log prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func searchSyslogPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if level, ok := req.Params.Arguments["level_filter"]; ok {
		c = append(c, fmt.Sprintf("- Level filter: %s", level))
	}
	if host, ok := req.Params.Arguments["host_filter"]; ok {
		c = append(c, fmt.Sprintf("- Host filter: %s", host))
	}
	if tag, ok := req.Params.Arguments["tag_filter"]; ok {
		c = append(c, fmt.Sprintf("- Tag filter: %s", tag))
	}
	if message, ok := req.Params.Arguments["message_filter"]; ok {
		c = append(c, fmt.Sprintf("- Message filter: %s", message))
	}
	if st, ok := req.Params.Arguments["start_time"]; ok {
		c = append(c, fmt.Sprintf("- Start Time: %s", st))
	}
	if et, ok := req.Params.Arguments["end_time"]; ok {
		c = append(c, fmt.Sprintf("- End Time: %s", et))
	}
	if limit, ok := req.Params.Arguments["limit"]; ok {
		c = append(c, fmt.Sprintf("- Limit: %s", limit))
	}
	p := "Search syslog of TWSNMP by using search_syslog tool"
	if len(c) > 0 {
		p = " with following filter.\n" + strings.Join(c, "\n")
	} else {
		p += "."
	}
	return &mcp.GetPromptResult{
		Description: "Search syslog prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func getSyslogSummaryPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if level, ok := req.Params.Arguments["level_filter"]; ok {
		c = append(c, fmt.Sprintf("- Level filter: %s", level))
	}
	if host, ok := req.Params.Arguments["host_filter"]; ok {
		c = append(c, fmt.Sprintf("- Host filter: %s", host))
	}
	if tag, ok := req.Params.Arguments["tag_filter"]; ok {
		c = append(c, fmt.Sprintf("- Tag filter: %s", tag))
	}
	if message, ok := req.Params.Arguments["message_filter"]; ok {
		c = append(c, fmt.Sprintf("- Message filter: %s", message))
	}
	if st, ok := req.Params.Arguments["start_time"]; ok {
		c = append(c, fmt.Sprintf("- Start Time: %s", st))
	}
	if et, ok := req.Params.Arguments["end_time"]; ok {
		c = append(c, fmt.Sprintf("- End Time: %s", et))
	}
	if topN, ok := req.Params.Arguments["top_n"]; ok {
		c = append(c, fmt.Sprintf("- Top N: %s", topN))
	}
	p := "Get syslog summary of TWSNMP by using get_syslog_summary tool"
	if len(c) > 0 {
		p = " with following filter.\n" + strings.Join(c, "\n")
	} else {
		p += "."
	}
	return &mcp.GetPromptResult{
		Description: "Get syslog summary prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func searchSnmpTrapLogPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if from, ok := req.Params.Arguments["from_filter"]; ok {
		c = append(c, fmt.Sprintf("- From filter: %s", from))
	}
	if trapType, ok := req.Params.Arguments["trap_type_filter"]; ok {
		c = append(c, fmt.Sprintf("- Trap type filter: %s", trapType))
	}
	if v, ok := req.Params.Arguments["variable_filter"]; ok {
		c = append(c, fmt.Sprintf("- Trap variable filter: %s", v))
	}
	if st, ok := req.Params.Arguments["start_time"]; ok {
		c = append(c, fmt.Sprintf("- Start Time: %s", st))
	}
	if et, ok := req.Params.Arguments["end_time"]; ok {
		c = append(c, fmt.Sprintf("- End Time: %s", et))
	}
	if limit, ok := req.Params.Arguments["limit"]; ok {
		c = append(c, fmt.Sprintf("- Limit: %s", limit))
	}
	p := "Search SNMP trap log of TWSNMP by using search_snmp_trap_log tool"
	if len(c) > 0 {
		p = " with following filter.\n" + strings.Join(c, "\n")
	} else {
		p += "."
	}
	return &mcp.GetPromptResult{
		Description: "Search SNMP trap log prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func getMIBTreePrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Get MIB tree of TWSNMP.",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "Get MIB tree of TWSNMP by using get_mib_tree tool."},
			},
		},
	}, nil
}

func getIPAddressListPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Get the list of IP address managed by TWSNMP.",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "Get the list of IP address managed by TWSNMP by using get_ip_address_list tool."},
			},
		},
	}, nil
}

func getResourceMonitorPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Get resource monitor info of TWSNMP by using get_resource_monitor_list tool.",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "Get resource monitor info of TWSNMP by using get_resource_monitor_list tool."},
			},
		},
	}, nil
}

func getServerCertificateListPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Get the list of server certificates managed by TWSNMP.",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "Get the list of server certificates managed by TWSNMP by using get_server_certificate_list tool."},
			},
		},
	}, nil
}

func addEventLogPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	c := []string{}
	if level, ok := req.Params.Arguments["level"]; !ok {
		return nil, fmt.Errorf("level is required")
	} else {
		c = append(c, fmt.Sprintf("- Level: %s", level))
	}
	if event, ok := req.Params.Arguments["event"]; !ok {
		return nil, fmt.Errorf("event is required")
	} else {
		c = append(c, fmt.Sprintf("- Event: %s", event))
	}
	if node, ok := req.Params.Arguments["node"]; ok {
		c = append(c, fmt.Sprintf("- Node: %s", node))
	}
	p := "Add event log to TWSNMP by using add_event_log tool with the following information:\n" + strings.Join(c, "\n")
	return &mcp.GetPromptResult{
		Description: "Add event log to TWSNMP.",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: p},
			},
		},
	}, nil
}

func getIPAddressInfoPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	ip, ok := req.Params.Arguments["ip"]
	if !ok {
		return nil, fmt.Errorf("ip address is required")
	}
	return &mcp.GetPromptResult{
		Description: "Get IP address information.",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf("Get IP address information by using get_ip_info tool. The IP address to look up is %s.", ip)},
			},
		},
	}, nil
}

func getMACAddressInfoPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	mac, ok := req.Params.Arguments["mac"]
	if !ok {
		return nil, fmt.Errorf("mac address is required")
	}
	return &mcp.GetPromptResult{
		Description: "Get MAC address information.",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf("Get MAC address information by using get_mac_info tool. The MAC address to look up is %s.", mac)},
			},
		},
	}, nil
}
