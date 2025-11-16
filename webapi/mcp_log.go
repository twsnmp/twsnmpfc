package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/twsnmp/rdap"
	"github.com/xhit/go-str2duration/v2"

	"github.com/twsnmp/twsnmpfc/datastore"
)

// search_event_log tool
type mcpEventLogEnt struct {
	Time  string `json:"time"`
	Type  string `json:"type"`
	Level string `json:"level"`
	Node  string `json:"node"`
	Event string `json:"event"`
}

type mcpSearchEventLogParams struct {
	NodeFilter  string `json:"node_filter" jsonschema:"node_filter specifies the search criteria for node names using regular expressions.If blank, no filter."`
	TypeFilter  string `json:"type_filter" jsonschema:"type_filter specifies the search criteria for type names using regular expressions.If blank, no filter."`
	LevelFilter string `json:"level_filter" jsonschema:"level_filter specifies the search criteria for level names using regular expressions.If blank, no filter.Level names can be info,normal,warn,low,high,debug"`
	EventFilter string `json:"event_filter" jsonschema:"event_filter specifies the search criteria for events using regular expressions.If blank, no filter."`
	StartTime   string `json:"start_time" jsonschema:"start date and time of logs to search or duration from now. If blank, defaults to the last 1 hour."`
	EndTime     string `json:"end_time" jsonschema:"end date and time of logs to search.empty or now is current time."`
	Limit       int    `json:"limit" jsonschema:"Limit on number of logs retrieved. min 100,max 10000"`
}

func mcpSearchEventLog(ctx context.Context, req *mcp.CallToolRequest, args mcpSearchEventLogParams) (*mcp.CallToolResult, any, error) {
	node := makeRegexFilter(args.NodeFilter)
	typeFilter := makeRegexFilter(args.TypeFilter)
	level := makeRegexFilter(args.LevelFilter)
	event := makeRegexFilter(args.EventFilter)
	start := args.StartTime
	if start == "" {
		start = "-1h"
	}
	end := args.EndTime
	st, et, err := getTimeRange(start, end)
	if err != nil {
		return nil, nil, err
	}
	limit := args.Limit
	if limit < 100 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	list := []mcpEventLogEnt{}
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if event != nil && !event.MatchString(l.Event) {
			return true
		}
		if level != nil && !level.MatchString(l.Level) {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(l.Type) {
			return true
		}
		if node != nil && !node.MatchString(l.NodeName) {
			return true
		}
		list = append(list, mcpEventLogEnt{
			Time:  time.Unix(0, l.Time).Format(time.RFC3339Nano),
			Type:  l.Type,
			Level: l.Level,
			Node:  l.NodeName,
			Event: l.Event,
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

// search_syslog tool
type mcpSyslogEnt struct {
	Time     string `json:"time"`
	Level    string `json:"level"`
	Host     string `json:"host"`
	Type     string `json:"type"`
	Tag      string `json:"tag"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
	Facility int    `json:"facility"`
}
type mcpSearchSyslogParams struct {
	LevelFilter   string `json:"level_filter" jsonschema:"level_filter specifies the search criteria for level names using regular expressions.If blank, no filter.Level names can be info,normal,warn,low,high,debug."`
	HostFilter    string `json:"host_filter" jsonschema:"host_filter specifies the search criteria for host names or IP address using regular expressions.If blank, no filter."`
	TagFilter     string `json:"tag_filter" jsonschema:"tag_filter specifies the search criteria for tag names using regular expressions.If blank, no filter."`
	MessageFilter string `json:"message_filter" jsonschema:"message_filter specifies the search criteria for messages using regular expressions.If blank, no filter."`
	StartTime     string `json:"start_time" jsonschema:"start date and time of logs to search or duration from now. If blank, defaults to the last 1 hour."`
	EndTime       string `json:"end_time" jsonschema:"end date and time of logs to search.empty or now is current time."`
	Limit         int    `json:"limit" jsonschema:"Limit on number of logs retrieved. min 100,max 10000"`
}

func mcpSearchSyslog(ctx context.Context, req *mcp.CallToolRequest, args mcpSearchSyslogParams) (*mcp.CallToolResult, any, error) {
	hostFilter := makeRegexFilter(args.HostFilter)
	tagFilter := makeRegexFilter(args.TagFilter)
	levelFilter := makeRegexFilter(args.LevelFilter)
	messageFilter := makeRegexFilter(args.MessageFilter)
	start := args.StartTime
	if start == "" {
		start = "-1h"
	}
	end := args.EndTime
	st, et, err := getTimeRange(start, end)
	if err != nil {
		return nil, nil, err
	}
	limit := args.Limit
	if limit < 100 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	log.Printf("mcp search_syslog limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
	list := []mcpSyslogEnt{}
	var hostMap = make(map[string]string)
	datastore.ForEachLog(st, et, "syslog", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		var ok bool
		var sv float64
		if sv, ok = sl["severity"].(float64); !ok {
			return true
		}
		var fac float64
		if fac, ok = sl["facility"].(float64); !ok {
			return true
		}
		var host string
		if host, ok = sl["hostname"].(string); !ok {
			return true
		}
		if h, ok := hostMap[host]; ok {
			host = h
		} else {
			if n := datastore.FindNodeFromIP(host); n != nil {
				h = fmt.Sprintf("%s(%s)", host, n.Name)
				hostMap[host] = h
				host = h
			} else {
				hostMap[host] = host
			}
		}
		var tag string
		var message string
		if tag, ok = sl["tag"].(string); !ok {
			if tag, ok = sl["app_name"].(string); !ok {
				return true
			}
			message = ""
			for i, k := range []string{"proc_id", "msg_id", "message", "structured_data"} {
				if m, ok := sl[k].(string); ok && m != "" {
					if i > 0 {
						message += " "
					}
					message += m
				}
			}
		} else {
			if message, ok = sl["content"].(string); !ok {
				return true
			}
		}

		e := mcpSyslogEnt{
			Time:     time.Unix(0, l.Time).Format(time.RFC3339Nano),
			Host:     host,
			Level:    getLevelFromSeverity(int(sv)),
			Type:     getSyslogType(int(sv), int(fac)),
			Tag:      tag,
			Facility: int(fac),
			Severity: int(sv),
			Message:  message,
		}
		if messageFilter != nil && !messageFilter.MatchString(e.Message) {
			return true
		}
		if tagFilter != nil && !tagFilter.MatchString(e.Tag) {
			return true
		}
		if levelFilter != nil && !levelFilter.MatchString(e.Level) {
			return true
		}
		if hostFilter != nil && !hostFilter.MatchString(e.Host) {
			return true
		}
		list = append(list, e)
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

// get_syslog_summary tool
type mcpSyslogSummaryEnt struct {
	Pattern string `json:"pattern"`
	Count   int    `json:"count"`
}

type mcpGetSyslogSummaryParams struct {
	LevelFilter   string `json:"level_filter" jsonschema:"level_filter specifies the search criteria for level names using regular expressions.If blank, no filter.Level names can be info,normal,warn,low,high,debug."`
	HostFilter    string `json:"host_filter" jsonschema:"host_filter specifies the search criteria for host names using regular expressions.If blank, no filter."`
	TagFilter     string `json:"tag_filter" jsonschema:"tag_filter specifies the search criteria for tag names using regular expressions.If blank, no filter."`
	MessageFilter string `json:"message_filter" jsonschema:"message_filter specifies the search criteria for messages using regular expressions.If blank, no filter."`
	StartTime     string `json:"start_time" jsonschema:"start date and time of logs to search or duration from now. If blank, defaults to the last 1 hour."`
	EndTime       string `json:"end_time" jsonschema:"end date and time of logs to search.empty or now is current time."`
	TopN          int    `json:"top_n" jsonschema:"Top n syslog pattern. min 5,max 500"`
}

func mcpGetSyslogSummary(ctx context.Context, req *mcp.CallToolRequest, args mcpGetSyslogSummaryParams) (*mcp.CallToolResult, any, error) {
	hostFilter := makeRegexFilter(args.HostFilter)
	tagFilter := makeRegexFilter(args.TagFilter)
	levelFilter := makeRegexFilter(args.LevelFilter)
	messageFilter := makeRegexFilter(args.MessageFilter)
	start := args.StartTime
	if start == "" {
		start = "-1h"
	}
	end := args.EndTime
	st, et, err := getTimeRange(start, end)
	if err != nil {
		return nil, nil, err
	}
	topN := args.TopN
	if topN < 5 {
		topN = 5
	}
	if topN > 500 {
		topN = 500
	}
	patternMap := make(map[string]int)
	hostMap := make(map[string]string)
	datastore.ForEachLog(st, et, "syslog", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		var ok bool
		var sv float64
		if sv, ok = sl["severity"].(float64); !ok {
			return true
		}
		var fac float64
		if fac, ok = sl["facility"].(float64); !ok {
			return true
		}
		var host string
		if host, ok = sl["hostname"].(string); !ok {
			return true
		}
		if h, ok := hostMap[host]; ok {
			host = h
		} else {
			if n := datastore.FindNodeFromIP(host); n != nil {
				h = fmt.Sprintf("%s(%s)", host, n.Name)
				hostMap[host] = h
				host = h
			} else {
				hostMap[host] = host
			}
		}
		var tag string
		var message string
		if tag, ok = sl["tag"].(string); !ok {
			if tag, ok = sl["app_name"].(string); !ok {
				return true
			}
			message = ""
			for i, k := range []string{"proc_id", "msg_id", "message", "structured_data"} {
				if m, ok := sl[k].(string); ok && m != "" {
					if i > 0 {
						message += " "
					}
					message += m
				}
			}
		} else {
			if message, ok = sl["content"].(string); !ok {
				return true
			}
		}
		level := getLevelFromSeverity(int(sv))
		syslogType := getSyslogType(int(sv), int(fac))

		if messageFilter != nil && !messageFilter.MatchString(message) {
			return true
		}
		if tagFilter != nil && !tagFilter.MatchString(tag) {
			return true
		}
		if levelFilter != nil && !levelFilter.MatchString(level) {
			return true
		}
		if hostFilter != nil && !hostFilter.MatchString(host) {
			return true
		}
		patternMap[normalizeLog(fmt.Sprintf("%s %s %s %s", host, syslogType, tag, message))]++
		return true
	})
	list := []mcpSyslogSummaryEnt{}
	for p, c := range patternMap {
		list = append(list, mcpSyslogSummaryEnt{
			Pattern: p,
			Count:   c,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})
	if len(list) > topN {
		list = list[:topN]
	}
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

var regNum = regexp.MustCompile(`\b-?\d+(\.\d+)?\b`)
var regUUID = regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`)
var regEmail = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
var regIP = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
var regMAC = regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}[:-]){5}(?:[0-9a-fA-F]{2})\b`)

func normalizeLog(s string) string {
	s = regUUID.ReplaceAllString(s, "#UUID#")
	s = regEmail.ReplaceAllString(s, "#EMAIL#")
	s = regIP.ReplaceAllString(s, "#IP#")
	s = regMAC.ReplaceAllString(s, "#MAC#")
	s = regNum.ReplaceAllString(s, "#NUM#")
	return s
}

// search_snmp_trap_log tool
type mcpSNMPTrapLogEnt struct {
	Time        string `json:"time"`
	FromAddress string `json:"from_address"`
	TrapType    string `json:"trap_type"`
	Variables   string `json:"variables"`
}

type mcpSearchSnmpTrapLogParams struct {
	FromFilter     string `json:"from_filter" jsonschema:"from_filter specifies the search criteria for trap sender address using regular expressions.If blank, no filter."`
	TrapTypeFilter string `json:"trap_type_filter" jsonschema:"trap_type_filter specifies the search criteria for SNMP trap types using regular expressions.If blank, no filter."`
	VariableFilter string `json:"variable_filter" jsonschema:"variable_filter specifies the search criteria for SNMP trap variables using regular expressions.If blank, no filter."`
	StartTime      string `json:"start_time" jsonschema:"start date and time of logs to search or duration from now. If blank, defaults to the last 1 hour."`
	EndTime        string `json:"end_time" jsonschema:"end date and time of logs to search.empty or now is current time."`
	Limit          int    `json:"limit" jsonschema:"Limit on number of logs retrieved. min 100,max 10000"`
}

func mcpSearchSnmpTrapLog(ctx context.Context, req *mcp.CallToolRequest, args mcpSearchSnmpTrapLogParams) (*mcp.CallToolResult, any, error) {
	fromFilter := makeRegexFilter(args.FromFilter)
	trapTypeFilter := makeRegexFilter(args.TrapTypeFilter)
	variableFilter := makeRegexFilter(args.VariableFilter)
	start := args.StartTime
	if start == "" {
		start = "-1h"
	}
	end := args.EndTime
	st, et, err := getTimeRange(start, end)
	if err != nil {
		return nil, nil, err
	}
	limit := args.Limit
	if limit < 100 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	log.Printf("mcp search_snmp_trap_log limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
	list := []mcpSNMPTrapLogEnt{}
	datastore.ForEachLog(st, et, "trap", func(l *datastore.LogEnt) bool {
		var sl = make(map[string]interface{})
		if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
			return true
		}
		var ok bool
		var fa string
		if fa, ok = sl["FromAddress"].(string); !ok {
			return true
		} else {
			a := strings.SplitN(fa, ":", 2)
			if len(a) == 2 {
				fa = a[0]
				n := datastore.FindNodeFromIP(a[0])
				if n != nil {
					fa += "(" + n.Name + ")"
				}
			}
		}
		var vars string
		if vars, ok = sl["Variables"].(string); !ok {
			return true
		}
		var ent string
		var trapType string
		if ent, ok = sl["Enterprise"].(string); !ok || ent == "" {
			a := trapOidRegexp.FindStringSubmatch(vars)
			if len(a) > 1 {
				trapType = a[1]
			} else {
				trapType = ""
			}
		} else {
			var gen float64
			if gen, ok = sl["GenericTrap"].(float64); !ok {
				return true
			}
			var spe float64
			if spe, ok = sl["SpecificTrap"].(float64); !ok {
				return true
			}
			trapType = fmt.Sprintf("%s:%d:%d", ent, int(gen), int(spe))
		}

		e := mcpSNMPTrapLogEnt{
			Time:        time.Unix(0, l.Time).Format(time.RFC3339Nano),
			FromAddress: fa,
			TrapType:    trapType,
			Variables:   vars,
		}
		if fromFilter != nil && !fromFilter.MatchString(e.FromAddress) {
			return true
		}
		if variableFilter != nil && !variableFilter.MatchString(e.Variables) {
			return true
		}
		if trapTypeFilter != nil && !trapTypeFilter.MatchString(e.TrapType) {
			return true
		}
		list = append(list, e)
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

// add_event_log tool
type mcpAddEventLogParams struct {
	Level string `json:"level" jsonschema:"Level of event (info,normal,warn,low,high,debug) default info."`
	Node  string `json:"node" jsonschema:"Node name associated with the event"`
	Event string `json:"event" jsonschema:"Event log contents"`
}

func mcpAddEventLog(ctx context.Context, req *mcp.CallToolRequest, args mcpAddEventLogParams) (*mcp.CallToolResult, any, error) {
	level := args.Level
	if level == "" {
		level = "info"
	}
	event := args.Event
	node := args.Node
	nodeID := ""
	if node != "" {
		if n := datastore.FindNodeFromName(node); n != nil {
			nodeID = n.ID
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:     time.Now().UnixNano(),
		Level:    level,
		Type:     "mcp",
		Event:    event,
		NodeName: node,
		NodeID:   nodeID,
	})
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "ok"},
		},
	}, nil, nil

}

// get_ip_address_info
type mcpGetIPInfoParams struct {
	IP string `json:"ip" jsonschema:"IP address"`
}

type mcpIPInfoEnt struct {
	IP              string
	Node            string
	DNSNames        []string
	Location        string
	RDAPIPVersion   string
	RDAPType        string
	RDAPHandle      string
	RDAPName        string
	RDAPCountry     string
	RDAPWhoisServer string
}

func mcpGetIPInfo(ctx context.Context, req *mcp.CallToolRequest, args mcpGetIPInfoParams) (*mcp.CallToolResult, any, error) {
	ip := args.IP
	info := new(mcpIPInfoEnt)
	info.IP = ip
	if n := datastore.FindNodeFromIP(ip); n != nil {
		info.Node = n.Name
	}
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	if names, err := r.LookupAddr(ctx, ip); err == nil && len(names) > 0 {
		info.DNSNames = names
	}
	info.Location = datastore.GetLoc(ip)
	if !strings.HasPrefix(info.Location, "LOCAL") {
		client := &rdap.Client{}
		if ri, err := client.QueryIP(ip); err == nil {
			info.RDAPIPVersion = ri.IPVersion
			info.RDAPName = ri.Name
			info.RDAPCountry = ri.Country
			info.RDAPWhoisServer = ri.Port43
			info.RDAPHandle = ri.Handle
			info.RDAPType = ri.Type
		}
	}
	j, err := json.Marshal(info)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_mac_address_info
type mcpGetMACInfoParams struct {
	MAC string `json:"mac" jsonschema:"MAC address"`
}

type mcpMACInfoEnt struct {
	MAC    string `json:"mac"`
	Node   string `json:"node"`
	IP     string `json:"ip"`
	Vendor string `json:"vendor"`
}

func mcpGetMACInfo(ctx context.Context, req *mcp.CallToolRequest, args mcpGetMACInfoParams) (*mcp.CallToolResult, any, error) {
	mac := normMACAddr(args.MAC)
	info := new(mcpMACInfoEnt)
	info.MAC = mac
	if n := datastore.FindNodeFromMAC(mac); n != nil {
		info.Node = n.Name
		info.IP = n.IP
	} else {
		info.IP = findIPFromArp(mac)
	}
	info.Vendor = datastore.FindVendor(mac)
	j, err := json.Marshal(info)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

func normMACAddr(m string) string {
	if hw, err := net.ParseMAC(m); err == nil {
		m = strings.ToUpper(hw.String())
		return m
	}
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

func findIPFromArp(mac string) string {
	ip := ""
	datastore.ForEachArp(func(a *datastore.ArpEnt) bool {
		if a.MAC == mac {
			ip = a.IP
			return false
		}
		return true
	})
	return ip
}

// getTimeRange
func getTimeRange(start, end string) (int64, int64, error) {
	var st time.Time
	var err error
	et := time.Now()
	if start == "" {
		return 0, 0, fmt.Errorf("start_time must not empty")
	}
	if d, err := str2duration.ParseDuration(start); err == nil {
		st = et.Add(d)
	} else if st, err = dateparse.ParseLocal(start); err != nil {
		return 0, 0, err
	}
	if end != "" && end != "now" {
		if et, err = dateparse.ParseLocal(end); err != nil {
			return 0, 0, err
		}
	}
	if st.UnixNano() > et.UnixNano() {
		return 0, 0, fmt.Errorf("start_time must before end_time")
	}
	return st.UnixNano(), et.UnixNano(), nil
}
