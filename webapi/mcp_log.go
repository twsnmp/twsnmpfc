package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xhit/go-str2duration/v2"

	"github.com/twsnmp/twsnmpfc/datastore"
)

// search_event_log tool
type mcpEventLogEnt struct {
	Time  string
	Type  string
	Level string
	Node  string
	Event string
}

func addSearchEventLogTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_event_log",
		mcp.WithDescription("search event log from TWSNMP"),
		mcp.WithString("node_filter",
			mcp.Description(
				`node_filter specifies the search criteria for node names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("type_filter",
			mcp.Description(
				`type_filter specifies the search criteria for type names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("level_filter",
			mcp.Description(
				`level_filter specifies the search criteria for level names using regular expressions.
If blank, no filter.
Level names can be "warn","low","high","debug","info" 
`),
		),
		mcp.WithString("event_filter",
			mcp.Description(
				`event_filter specifies the search criteria for events using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithNumber("limit_log_count",
			mcp.DefaultNumber(100),
			mcp.Max(10000),
			mcp.Min(1),
			mcp.Description("Limit on number of logs retrieved. min 100,max 10000"),
		),
		mcp.WithString("start_time",
			mcp.DefaultString("-1h"),
			mcp.Description(
				`start date and time of logs to search
or duration from now

A duration string is a possibly signed sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "-300ms", "-1.5h" or "-2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

Example:
 2025/05/07 05:59:00
 -1h
`),
		),
		mcp.WithString("end_time",
			mcp.DefaultString(""),
			mcp.Description(
				`end date and time of logs to search.
empty or "now" is current time.

Example:
 2025/05/07 06:59:00
 now
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node := makeRegexFilter(request.GetString("node_filter", ""))
		typeFilter := makeRegexFilter(request.GetString("type_filter", ""))
		level := makeRegexFilter(request.GetString("level_filter", ""))
		event := makeRegexFilter(request.GetString("event_filter", ""))
		start := request.GetString("start_time", "-1h")
		end := request.GetString("end_time", "")
		st, et, err := getTimeRange(start, end)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		limit := request.GetInt("limit_log_count", 100)
		log.Printf("mcp search_event_log limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// add_event_log tool
func addAddEventLogTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("add_event_log",
		mcp.WithDescription("add event log to TWSNMP"),
		mcp.WithString("level",
			mcp.Enum("info", "normal", "warn", "low", "high"),
			mcp.Description("Level of event (info,normal,warn,low,high)"),
		),
		mcp.WithString("node",
			mcp.Description("Node name associated with the event"),
		),
		mcp.WithString("event",
			mcp.Description("Event log contents"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		level := request.GetString("level", "info")
		event := request.GetString("event", "")
		node := request.GetString("node", "")
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
		return mcp.NewToolResultText("ok"), nil
	})
}

// search_syslog tool
type mcpSyslogEnt struct {
	Time     string
	Level    string
	Host     string
	Type     string
	Tag      string
	Message  string
	Severity int
	Facility int
}

func addSearchSyslogTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_syslog",
		mcp.WithDescription("search syslog from TWSNMP"),
		mcp.WithString("host_filter",
			mcp.Description(
				`host_filter specifies the search criteria for host names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("tag_filter",
			mcp.Description(
				`tag_filter specifies the search criteria for tag names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("level_filter",
			mcp.Description(
				`level_filter specifies the search criteria for level names using regular expressions.
If blank, no filter.
Level names can be "warn","low","high","debug","info" 
`),
		),
		mcp.WithString("message_filter",
			mcp.Description(
				`message_filter specifies the search criteria for messages using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithNumber("limit_log_count",
			mcp.DefaultNumber(100),
			mcp.Max(10000),
			mcp.Min(1),
			mcp.Description("Limit on number of logs retrieved. min 100,max 10000"),
		),
		mcp.WithString("start_time",
			mcp.DefaultString("-1h"),
			mcp.Description(
				`start date and time of logs to search
or duration from now

A duration string is a possibly signed sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "300ms", "-1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

Example:
 2025/05/07 05:59:00
 -1h
`),
		),
		mcp.WithString("end_time",
			mcp.DefaultString(""),
			mcp.Description(
				`end date and time of logs to search.
empty or "now" is current time.

Example:
 2025/05/07 06:59:00
 now
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		host := makeRegexFilter(request.GetString("host_filter", ""))
		tag := makeRegexFilter(request.GetString("tag_filter", ""))
		level := makeRegexFilter(request.GetString("level_filter", ""))
		message := makeRegexFilter(request.GetString("message_filter", ""))
		start := request.GetString("start_time", "-1h")
		end := request.GetString("end_time", "")
		st, et, err := getTimeRange(start, end)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		limit := request.GetInt("limit_log_count", 100)
		log.Printf("mcp search_syslog limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
		list := []mcpSyslogEnt{}
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
			e := mcpSyslogEnt{}
			if e.Host, ok = sl["hostname"].(string); !ok {
				return true
			}
			if e.Tag, ok = sl["tag"].(string); !ok {
				if e.Tag, ok = sl["app_name"].(string); !ok {
					return true
				}
				e.Message = ""
				for i, k := range []string{"proc_id", "msg_id", "message", "structured_data"} {
					if m, ok := sl[k].(string); ok && m != "" {
						if i > 0 {
							e.Message += " "
						}
						e.Message += m
					}
				}
			} else {
				if e.Message, ok = sl["content"].(string); !ok {
					return true
				}
			}
			e.Time = time.Unix(0, l.Time).Format(time.RFC3339Nano)
			e.Level = getLevelFromSeverity(int(sv))
			e.Type = getSyslogType(int(sv), int(fac))
			e.Facility = int(fac)
			e.Severity = int(sv)
			if message != nil && !message.MatchString(e.Message) {
				return true
			}
			if tag != nil && !tag.MatchString(e.Tag) {
				return true
			}
			if level != nil && !level.MatchString(e.Level) {
				return true
			}
			if host != nil && !host.MatchString(e.Host) {
				return true
			}
			list = append(list, e)
			return len(list) < limit
		})
		j, err := json.Marshal(&list)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// get_syslog_summary tool
type mcpSyslogSummaryEnt struct {
	Pattern string
	Count   int
}

func addGetSyslogSummaryTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_syslog_summary",
		mcp.WithDescription("get syslog summary from TWSNMP"),
		mcp.WithString("host_filter",
			mcp.Description(
				`host_filter specifies the search criteria for host names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("tag_filter",
			mcp.Description(
				`tag_filter specifies the search criteria for tag names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("level_filter",
			mcp.Description(
				`level_filter specifies the search criteria for level names using regular expressions.
If blank, no filter.
Level names can be "warn","low","high","debug","info" 
`),
		),
		mcp.WithString("message_filter",
			mcp.Description(
				`message_filter specifies the search criteria for messages using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithNumber("top_n",
			mcp.DefaultNumber(10),
			mcp.Max(100),
			mcp.Min(5),
			mcp.Description("Top n syslog pattern. min 5,max 100"),
		),
		mcp.WithString("start_time",
			mcp.DefaultString("-1h"),
			mcp.Description(
				`start date and time of logs to search
or duration from now

A duration string is a possibly signed sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "300ms", "-1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

Example:
 2025/05/07 05:59:00
 -1h
`),
		),
		mcp.WithString("end_time",
			mcp.DefaultString(""),
			mcp.Description(
				`end date and time of logs to search.
empty or "now" is current time.

Example:
 2025/05/07 06:59:00
 now
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		host := makeRegexFilter(request.GetString("host_filter", ""))
		tag := makeRegexFilter(request.GetString("tag_filter", ""))
		level := makeRegexFilter(request.GetString("level_filter", ""))
		message := makeRegexFilter(request.GetString("message_filter", ""))
		start := request.GetString("start_time", "-1h")
		end := request.GetString("end_time", "")
		st, et, err := getTimeRange(start, end)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		topN := request.GetInt("top_n", 10)
		log.Printf("mcp get_syslog_summary topn=%d st=%v et=%v", topN, time.Unix(0, st), time.Unix(0, et))
		patternMap := make(map[string]int)
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
			e := mcpSyslogEnt{}
			if e.Host, ok = sl["hostname"].(string); !ok {
				return true
			}
			if e.Tag, ok = sl["tag"].(string); !ok {
				if e.Tag, ok = sl["app_name"].(string); !ok {
					return true
				}
				e.Message = ""
				for i, k := range []string{"proc_id", "msg_id", "message", "structured_data"} {
					if m, ok := sl[k].(string); ok && m != "" {
						if i > 0 {
							e.Message += " "
						}
						e.Message += m
					}
				}
			} else {
				if e.Message, ok = sl["content"].(string); !ok {
					return true
				}
			}
			e.Level = getLevelFromSeverity(int(sv))
			e.Type = getSyslogType(int(sv), int(fac))
			if message != nil && !message.MatchString(e.Message) {
				return true
			}
			if tag != nil && !tag.MatchString(e.Tag) {
				return true
			}
			if level != nil && !level.MatchString(e.Level) {
				return true
			}
			if host != nil && !host.MatchString(e.Host) {
				return true
			}
			patternMap[normalizeLog(fmt.Sprintf("%s %s %s %s", e.Host, e.Type, e.Tag, e.Message))]++
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

var regNum = regexp.MustCompile(`\b-?\d+(\.\d+)?\b`)
var regUUDI = regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`)
var regEmail = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
var regIP = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
var regMAC = regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}[:-]){5}(?:[0-9a-fA-F]{2})\b`)

func normalizeLog(s string) string {
	s = regUUDI.ReplaceAllString(s, "#UUID#")
	s = regEmail.ReplaceAllString(s, "#EMAIL#")
	s = regIP.ReplaceAllString(s, "#IP#")
	s = regMAC.ReplaceAllString(s, "#MAC#")
	s = regNum.ReplaceAllString(s, "#NUM#")
	return s
}

// search_snmp_trap_log tool
type mcpSNMPTrapLogEnt struct {
	Time        string
	FromAddress string
	TrapType    string
	Variables   string
}

func addSearchSNMPTrapLogTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_snmp_trap_log",
		mcp.WithDescription("search SNMP trap log from TWSNMP"),
		mcp.WithString("from_filter",
			mcp.Description(
				`from_filter specifies the search criteria for trap sender address using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("trap_type_filter",
			mcp.Description(
				`trap_type_filter specifies the search criteria for SNMP trap types using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("variable_filter",
			mcp.Description(
				`variable_filter specifies the search criteria for SNMP trap variables using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithNumber("limit_log_count",
			mcp.DefaultNumber(100),
			mcp.Max(10000),
			mcp.Min(1),
			mcp.Description("Limit on number of logs retrieved. min 100,max 10000"),
		),
		mcp.WithString("start_time",
			mcp.DefaultString("-1h"),
			mcp.Description(
				`start date and time of logs to search
or duration from now

A duration string is a possibly signed sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "300ms", "-1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

Example:
 2025/05/07 05:59:00
 -1h
`),
		),
		mcp.WithString("end_time",
			mcp.DefaultString(""),
			mcp.Description(
				`end date and time of logs to search.
empty or "now" is current time.

Example:
 2025/05/07 06:59:00
 now
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		from := makeRegexFilter(request.GetString("host_filter", ""))
		trapType := makeRegexFilter(request.GetString("trap_type_filter", ""))
		variable := makeRegexFilter(request.GetString("variable_filter", ""))
		start := request.GetString("start_time", "-1h")
		end := request.GetString("end_time", "")
		st, et, err := getTimeRange(start, end)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		limit := request.GetInt("limit_log_count", 100)
		log.Printf("mcp search_snmp_trap_log limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
		list := []mcpSNMPTrapLogEnt{}
		datastore.ForEachLog(st, et, "trap", func(l *datastore.LogEnt) bool {
			var sl = make(map[string]interface{})
			if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
				return true
			}
			var ok bool
			e := mcpSNMPTrapLogEnt{}
			if fa, ok := sl["FromAddress"].(string); !ok {
				return true
			} else {
				a := strings.SplitN(fa, ":", 2)
				if len(a) == 2 {
					e.FromAddress = a[0]
					n := datastore.FindNodeFromIP(a[0])
					if n != nil {
						e.FromAddress += "(" + n.Name + ")"
					}
				} else {
					e.FromAddress = fa
				}
			}
			if e.Variables, ok = sl["Variables"].(string); !ok {
				return true
			}
			var ent string
			if ent, ok = sl["Enterprise"].(string); !ok || ent == "" {
				a := trapOidRegexp.FindStringSubmatch(e.Variables)
				if len(a) > 1 {
					e.TrapType = a[1]
				} else {
					e.TrapType = ""
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
				e.TrapType = fmt.Sprintf("%s:%d:%d", ent, int(gen), int(spe))
			}
			e.Time = time.Unix(0, l.Time).Format(time.RFC3339Nano)
			if from != nil && !from.MatchString(e.FromAddress) {
				return true
			}
			if variable != nil && !variable.MatchString(e.Variables) {
				return true
			}
			if trapType != nil && !trapType.MatchString(e.TrapType) {
				return true
			}
			list = append(list, e)
			return len(list) < limit
		})
		j, err := json.Marshal(&list)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
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
