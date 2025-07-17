package webapi

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xhit/go-str2duration/v2"

	"github.com/twsnmp/twsnmpfc/datastore"
)

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
		mcp.WithString("time_range",
			mcp.Required(),
			mcp.Description(
				`Time range of logs to search
		format is "start date/time, duration"
		or "start date/time, end date/time".

		A duration string is a possibly signed sequence of
		decimal numbers, each with optional fraction and a unit suffix,
		such as "300ms", "-1.5h" or "2h45m".
		Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

		Example:
		2025/05/07 05:59:00,1h
		2025/05/07 05:59:00,2025/05/07 06:59:00
		`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		host := makeRegexFilter(request.GetString("host_filter", ""))
		tag := makeRegexFilter(request.GetString("tag_filter", ""))
		level := makeRegexFilter(request.GetString("level_filter", ""))
		message := makeRegexFilter(request.GetString("message_filter", ""))
		timeRange, err := request.RequireString("time_range")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		st, et := getTimeRange(timeRange)
		limit := request.GetInt("limit_log_count", 100)
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
			log := mcpSyslogEnt{}
			if log.Host, ok = sl["hostname"].(string); !ok {
				return true
			}
			if log.Tag, ok = sl["tag"].(string); !ok {
				if log.Tag, ok = sl["app_name"].(string); !ok {
					return true
				}
				log.Message = ""
				for i, k := range []string{"proc_id", "msg_id", "message", "structured_data"} {
					if m, ok := sl[k].(string); ok && m != "" {
						if i > 0 {
							log.Message += " "
						}
						log.Message += m
					}
				}
			} else {
				if log.Message, ok = sl["content"].(string); !ok {
					return true
				}
			}
			log.Time = time.Unix(0, l.Time).Format(time.RFC3339Nano)
			log.Level = getLevelFromSeverity(int(sv))
			log.Type = getSyslogType(int(sv), int(fac))
			log.Facility = int(fac)
			log.Severity = int(sv)
			if message != nil && !message.MatchString(log.Message) {
				return true
			}
			if tag != nil && !tag.MatchString(log.Tag) {
				return true
			}
			if level != nil && !level.MatchString(log.Level) {
				return true
			}
			if host != nil && !host.MatchString(log.Host) {
				return true
			}
			list = append(list, log)
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
func getTimeRange(timeRange string) (int64, int64) {
	st := time.Unix(0, 0)
	et := time.Now()
	a := strings.SplitN(timeRange, ",", 2)
	if len(a) == 1 && a[0] != "" {
		if d, err := str2duration.ParseDuration(a[0]); err == nil {
			st = et.Add(d * -1)
		} else if t, err := dateparse.ParseLocal(a[0]); err == nil {
			st = t
		}
	} else {
		if t, err := dateparse.ParseLocal(a[0]); err == nil {
			st = t
			if t, err := dateparse.ParseLocal(a[1]); err == nil {
				et = t
			} else if d, err := str2duration.ParseDuration(a[1]); err == nil {
				et = st.Add(d)
			}
		}
	}
	return st.UnixNano(), et.UnixNano()
}
