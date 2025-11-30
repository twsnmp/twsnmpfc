package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/twsnmp/twsnmpfc/datastore"
	"github.com/twsnmp/twsnmpfc/ping"
)

// get_node_list tool
type mcpNodeEnt struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IP          string `json:"ip"`
	MAC         string `json:"mac"`
	State       string `json:"state"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

type mcpGetNodeListParams struct {
	NameFilter  string `json:"name_filter" jsonschema:"name_filter specifies the search criteria for node names using regular expressions.If blank, all nodes are searched."`
	IPFilter    string `json:"ip_filter" jsonschema:"ip_filter specifies the search criteria for node IP address using regular expressions.If blank, all nodes are searched."`
	StateFilter string `json:"state_filter" jsonschema:"state_filter uses a regular expression to specify search criteria for node state names(normal,warn,low,high,repair,unknown) If blank, all nodes are searched."`
}

func mcpGetNodeList(ctx context.Context, req *mcp.CallToolRequest, args mcpGetNodeListParams) (*mcp.CallToolResult, any, error) {
	name := makeRegexFilter(args.NameFilter)
	ip := makeRegexFilter(args.IPFilter)
	state := makeRegexFilter(args.StateFilter)

	list := []mcpNodeEnt{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if name != nil && !name.MatchString(n.Name) {
			return true
		}
		if ip != nil && !ip.MatchString(n.IP) {
			return true
		}
		if state != nil && !state.MatchString(n.State) {
			return true
		}
		list = append(list, mcpNodeEnt{
			ID:          n.ID,
			Name:        n.Name,
			IP:          n.IP,
			MAC:         n.MAC,
			X:           n.X,
			Y:           n.Y,
			Icon:        n.Icon,
			Description: n.Descr,
			State:       n.State,
		})
		return true
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_network_list tool

type mcpPortEnt struct {
	Name  string `json:"name"`
	State string `json:"state"`
}
type mcpNetworkEnt struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	IP          string       `json:"ip"`
	Ports       []mcpPortEnt `json:"ports"`
	X           int          `json:"x"`
	Y           int          `json:"y"`
	Description string       `json:"description"`
}
type mcpGetNetworkListParams struct {
	NameFilter string `json:"name_filter" jsonschema:"name_filter specifies the search criteria for network names using regular expressions.If blank, all networks are searched."`
	IPFilter   string `json:"ip_filter" jsonschema:"ip_filter specifies the search criteria for network IP address using regular expressions.If blank, all networks are searched."`
}

func mcpGetNetworkList(ctx context.Context, req *mcp.CallToolRequest, args mcpGetNetworkListParams) (*mcp.CallToolResult, any, error) {
	name := makeRegexFilter(args.NameFilter)
	ip := makeRegexFilter(args.IPFilter)
	list := []mcpNetworkEnt{}
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		if name != nil && !name.MatchString(n.Name) {
			return true
		}
		if ip != nil && !ip.MatchString(n.IP) {
			return true
		}
		ports := []mcpPortEnt{}
		for _, p := range n.Ports {
			ports = append(ports, mcpPortEnt{
				Name:  p.Name,
				State: p.State,
			})
		}
		list = append(list, mcpNetworkEnt{
			ID:          n.ID,
			Name:        n.Name,
			IP:          n.IP,
			X:           n.X,
			Y:           n.Y,
			Description: n.Descr,
			Ports:       ports,
		})
		return true
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_polling_list tool
type mcpPollingEnt struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	NodeID   string         `json:"node_id"`
	NodeName string         `json:"node_name"`
	Type     string         `json:"type"`
	Level    string         `json:"level"`
	State    string         `json:"state"`
	LastTime string         `json:"last_time"`
	Result   map[string]any `json:"result"`
}
type mcpGetPollingListParams struct {
	TypeFilter     string `json:"type_filter" jsonschema:"type_filter uses a regular expression to specify search criteria for polling type names.If blank, all pollings are searched.Type names can be ping,tcp,http,dns,twsnmp,syslog"`
	NameFilter     string `json:"name_filter" jsonschema:"name_filter specifies the search criteria for polling names using regular expressions.If blank, all pollings are searched."`
	NodeNameFilter string `json:"node_name_filter" jsonschema:"node_name_filter specifies the search criteria for node names of polling using regular expressions.If blank, all pollings are searched."`
	StateFilter    string `json:"state_filter" jsonschema:"state_filter uses a regular expression to specify search criteria for polling state names.If blank, all pollings are searched.State names can be normal,warn,low,high,repair,unknown"`
}

func mcpGetPollingList(ctx context.Context, req *mcp.CallToolRequest, args mcpGetPollingListParams) (*mcp.CallToolResult, any, error) {
	name := makeRegexFilter(args.NameFilter)
	nodeName := makeRegexFilter(args.NodeNameFilter)
	typeFilter := makeRegexFilter(args.TypeFilter)
	state := makeRegexFilter(args.StateFilter)

	list := []mcpPollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if name != nil && !name.MatchString(p.Name) {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(p.Type) {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		if nodeName != nil && !nodeName.MatchString(n.Name) {
			return true
		}
		if state != nil && !state.MatchString(p.State) {
			return true
		}
		list = append(list, mcpPollingEnt{
			ID:       p.ID,
			Name:     p.Name,
			NodeName: n.Name,
			NodeID:   n.ID,
			Type:     p.Type,
			Level:    p.Level,
			LastTime: time.Unix(0, p.LastTime).Format(time.RFC3339Nano),
			State:    p.State,
			Result:   p.Result,
		})
		return true
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_polling_log tool
type mcpPollingLogEnt struct {
	Time   string         `json:"time"`
	State  string         `json:"state"`
	Result map[string]any `json:"result"`
}
type mcpGetPollingLogParams struct {
	ID    string `json:"id" jsonschema:"The ID of the polling to retrieve the polling log"`
	Limit int    `json:"limit" jsonschema:"Limit on number of logs retrieved. The value must be between 100 and 2000. If outside this range, it defaults to 100."`
}

func mcpGetPollingLog(ctx context.Context, req *mcp.CallToolRequest, args mcpGetPollingLogParams) (*mcp.CallToolResult, any, error) {
	id := args.ID
	if id == "" {
		return nil, nil, fmt.Errorf("no id")
	}
	polling := datastore.GetPolling(id)
	if polling == nil {
		return nil, nil, fmt.Errorf("polling not found")
	}
	limit := args.Limit
	if limit < 100 || limit > 2000 {
		limit = 100
	}
	list := []mcpPollingLogEnt{}

	datastore.ForEachLastPollingLog(id, func(l *datastore.PollingLogEnt) bool {
		list = append(list, mcpPollingLogEnt{
			Time:   time.Unix(0, l.Time).Format(time.RFC3339),
			State:  l.State,
			Result: l.Result,
		})
		return len(list) < limit
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

func mcpGetPollingLogData(ctx context.Context, req *mcp.CallToolRequest, args mcpGetPollingLogParams) (*mcp.CallToolResult, any, error) {
	id := args.ID
	if id == "" {
		return nil, nil, fmt.Errorf("no id")
	}
	polling := datastore.GetPolling(id)
	if polling == nil {
		return nil, nil, fmt.Errorf("polling not found")
	}
	limit := args.Limit
	if limit < 100 || limit > 2000 {
		limit = 100
	}
	list := []mcpPollingLogEnt{}
	datastore.ForEachLastPollingLog(id, func(l *datastore.PollingLogEnt) bool {
		list = append(list, mcpPollingLogEnt{
			Time:   time.Unix(0, l.Time).Format(time.RFC3339),
			State:  l.State,
			Result: l.Result,
		})
		return len(list) < limit
	})
	if len(list) < 1 {
		return nil, nil, fmt.Errorf("polling log not found")
	}
	keys := []string{}
	for k, v := range list[0].Result {
		if k == "lastTime" {
			continue
		}
		if _, ok := v.(float64); !ok {
			continue
		}
		keys = append(keys, k)
	}
	csv := []string{"time,state," + strings.Join(keys, ",")}
	for _, l := range list {
		s := fmt.Sprintf("%s,%s", l.Time, l.State)
		for _, k := range keys {
			if v, ok := l.Result[k].(float64); ok {
				s += "," + fmt.Sprintf("%f", v)
			} else {
				s += ","
			}
		}
		csv = append(csv, s)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(csv, "\n")},
		},
	}, nil, nil
}

// do_ping tool
type mcpPingEnt struct {
	Result       string `json:"result"`
	Time         string `json:"time"`
	RTT          string `json:"rtt"`
	RTTNano      int64  `json:"rtt_nano"`
	Size         int    `json:"size"`
	TTL          int    `json:"ttl"`
	ResponseFrom string `json:"response_from"`
	Location     string `json:"location"`
}

type mcpDoPingParams struct {
	Target  string `json:"target" jsonschema:"ping target ip address or host name"`
	Size    int    `json:"size" jsonschema:"ping packet size. min 1,max 1500,default 64"`
	TTL     int    `json:"ttl" jsonschema:"ip packet TTL. min 1,max 255,default 254"`
	Timeout int    `json:"timeout" jsonschema:"timeout sec of ping. min 1,max 10,default 3"`
}

func mcpDoPing(ctx context.Context, req *mcp.CallToolRequest, args mcpDoPingParams) (*mcp.CallToolResult, any, error) {
	target := getTargetIP(args.Target)
	if target == "" {
		return nil, nil, fmt.Errorf("target ip not found")
	}
	timeout := args.Timeout
	if timeout < 1 || timeout > 10 {
		timeout = 3
	}

	size := args.Size
	if size < 1 || size > 1500 {
		size = 64
	}
	ttl := args.TTL
	if ttl < 1 || ttl > 255 {
		ttl = 254
	}
	pe := ping.DoPing(target, timeout, 0, size, ttl)
	res := mcpPingEnt{
		Result:       pe.Stat.String(),
		Time:         time.Now().Format(time.RFC3339),
		RTT:          time.Duration(pe.Time).String(),
		Size:         pe.Size,
		ResponseFrom: pe.RecvSrc,
		TTL:          pe.RecvTTL,
		RTTNano:      pe.Time,
	}
	if pe.RecvSrc != "" {
		res.Location = datastore.GetLoc(pe.RecvSrc)
	}
	j, err := json.Marshal(&res)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// getTargetIP: targetからIPアドレスを取得する、targetはノード名、ホスト名、IPアドレス
func getTargetIP(target string) string {
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

// get_mib_tree tool
func mcpGetMIBTree(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	j, err := json.Marshal(&datastore.MIBTree)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpMIBEnt struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type mcpSnmpWalkParams struct {
	Target        string `json:"target" jsonschema:"target IP address or node name."`
	MIBObjectName string `json:"mib_object_name" jsonschema:"mib object name"`
	Community     string `json:"community" jsonschema:"community name for snmp v2c mode. Can be omitted if the target is a managed node name."`
	User          string `json:"user" jsonschema:"User name for snmp v3 mode. Can be omitted if the target is a managed node name."`
	Password      string `json:"password" jsonschema:"Password for snmp v3 mode. Can be omitted if the target is a managed node name."`
	SnmpMode      string `json:"snmp_mode" jsonschema:"snmp mode (v2c,v3auth,v3authpriv,v3authprivex). Can be omitted if the target is a managed node name."`
}

func mcpSnmpWalk(ctx context.Context, req *mcp.CallToolRequest, args mcpSnmpWalkParams) (*mcp.CallToolResult, any, error) {
	community := args.Community
	user := args.User
	password := args.Password
	snmpMode := args.SnmpMode
	name := args.MIBObjectName
	if name == "" {
		return nil, nil, fmt.Errorf("no mib_object_name")
	}
	target := args.Target
	if target == "" {
		return nil, nil, fmt.Errorf("no target")
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
		target = getTargetIP(target)
		if target == "" {
			return nil, nil, fmt.Errorf("target not found")
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
	err := agent.Connect()
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}
	j, err := json.Marshal(&res)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpSnmpSetParams struct {
	Target        string `json:"target" jsonschema:"target IP address or node name."`
	MIBObjectName string `json:"mib_object_name" jsonschema:"mib object name"`
	Community     string `json:"community" jsonschema:"community name for snmp v2c mode. Can be omitted if the target is a managed node name."`
	User          string `json:"user" jsonschema:"User name for snmp v3 mode. Can be omitted if the target is a managed node name."`
	Password      string `json:"password" jsonschema:"Password for snmp v3 mode. Can be omitted if the target is a managed node name."`
	SnmpMode      string `json:"snmp_mode" jsonschema:"snmp mode (v2c,v3auth,v3authpriv,v3authprivex). Can be omitted if the target is a managed node name."`
	Type          string `json:"type" jsonschema:"Type of set value(integer or string)"`
	Value         string `json:"value" jsonschema:"Set value"`
}

func mcpSnmpSet(ctx context.Context, req *mcp.CallToolRequest, args mcpSnmpSetParams) (*mcp.CallToolResult, any, error) {
	community := args.Community
	user := args.User
	password := args.Password
	snmpMode := args.SnmpMode
	name := args.MIBObjectName
	if name == "" {
		return nil, nil, fmt.Errorf("no mib_object_name")
	}
	target := args.Target
	if target == "" {
		return nil, nil, fmt.Errorf("no target")
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
		target = getTargetIP(target)
		if target == "" {
			return nil, nil, fmt.Errorf("target not found")
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
	err := agent.Connect()
	if err != nil {
		return nil, nil, err
	}
	defer agent.Conn.Close()
	setPDU := []gosnmp.SnmpPDU{}
	switch args.Type {
	case "integer":
		i, err := strconv.Atoi(args.Value)
		if err != nil {
			return nil, nil, err
		}
		setPDU = append(setPDU, gosnmp.SnmpPDU{
			Name:  nameToOID(name),
			Type:  gosnmp.Integer,
			Value: i,
		})
	default:
		// string
		setPDU = append(setPDU, gosnmp.SnmpPDU{
			Name:  nameToOID(name),
			Type:  gosnmp.OctetString,
			Value: []byte(args.Value),
		})
	}
	r, err := agent.Set(setPDU)
	if err != nil {
		return nil, nil, err
	}
	if r.Error != gosnmp.NoError {
		return nil, nil, fmt.Errorf("snmp set %s", r.Error.String())
	}
	for _, variable := range r.Variables {
		name := datastore.MIBDB.OIDToName(variable.Name)
		value := datastore.GetMIBValueString(name, &variable, false)
		res = append(res, mcpMIBEnt{
			Name:  name,
			Value: value,
		})
	}
	j, err := json.Marshal(&res)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpAddNodeParams struct {
	Name        string `json:"name" jsonschema:"node name. A PING polling is also added automatically for the new node."`
	IP          string `json:"ip" jsonschema:"node ip address"`
	Icon        string `json:"icon" jsonschema:"icon of node. Defaults to 'desktop' if not specified."`
	Description string `json:"description" jsonschema:"description of node"`
	X           int    `json:"x" jsonschema:"x position of node. min 64,max 1000"`
	Y           int    `json:"y" jsonschema:"y position of node. min 64,max 1000"`
}

func mcpAddNode(ctx context.Context, req *mcp.CallToolRequest, args mcpAddNodeParams) (*mcp.CallToolResult, any, error) {
	icon := args.Icon
	if icon == "" {
		icon = "desktop"
	}
	descr := args.Description
	name := args.Name
	if name == "" {
		return nil, nil, fmt.Errorf("node name is empty")
	}
	ip := args.IP
	if ip == "" {
		return nil, nil, fmt.Errorf("ip is empty")
	}
	x := args.X
	if x < 64 || x > 1000 {
		return nil, nil, fmt.Errorf("invalid x")
	}
	y := args.Y
	if y < 64 || y > 1000 {
		return nil, nil, fmt.Errorf("invalid y")
	}
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
		return nil, nil, err
	}
	datastore.AddPolling(&datastore.PollingEnt{
		Name:   "PING",
		Type:   "ping",
		NodeID: n.ID})
	j, err := json.Marshal(&mcpNodeEnt{
		ID:          n.ID,
		Name:        n.Name,
		Description: n.Descr,
		IP:          n.IP,
		State:       n.State,
		X:           n.X,
		Y:           n.Y,
		Icon:        n.Icon,
	})
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// update_node
type mcpUpdateNodeParams struct {
	ID          string `json:"id" jsonschema:"node id or current name or current ip"`
	Name        string `json:"name" jsonschema:"new node name or empty"`
	IP          string `json:"ip" jsonschema:"new ip address or empty"`
	Icon        string `json:"icon" jsonschema:"new icon or empty"`
	Description string `json:"description" jsonschema:"description of node"`
	X           int    `json:"x" jsonschema:"x position of node. min 64,max 1000. Specify 0 to skip updates."`
	Y           int    `json:"y" jsonschema:"y position of node. min 64,max 1000. Specify 0 to skip updates."`
}

func mcpUpdateNode(ctx context.Context, req *mcp.CallToolRequest, args mcpUpdateNodeParams) (*mcp.CallToolResult, any, error) {
	id := args.ID
	n := datastore.GetNode(id)
	if n == nil {
		n = datastore.FindNodeFromName(id)
		if n == nil {
			n = datastore.FindNodeFromIP(id)
			if n == nil {
				return nil, nil, fmt.Errorf("node not found")
			}
		}
	}
	icon := args.Icon
	descr := args.Description
	name := args.Name
	ip := args.IP
	x := args.X
	if x != 0 {
		if x < 64 || x > 1000 {
			return nil, nil, fmt.Errorf("invalid x")
		}
	}
	y := args.Y
	if y != 0 {
		if y < 64 || y > 1000 {
			return nil, nil, fmt.Errorf("invalid y")
		}
	}
	if x > 0 {
		n.X = x
	}
	if y > 0 {
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
	if ip != "" {
		n.IP = ip
	}
	j, err := json.Marshal(&mcpNodeEnt{
		ID:          n.ID,
		Name:        n.Name,
		Description: n.Descr,
		IP:          n.IP,
		State:       n.State,
		X:           n.X,
		Y:           n.Y,
		Icon:        n.Icon,
	})
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}
