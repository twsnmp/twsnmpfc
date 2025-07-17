package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/twsnmp/twsnmpfc/datastore"
)

// get_sensor_list tool
type mcpSensorEnt struct {
	ID        string
	Host      string
	Type      string
	Total     int64
	Send      int64
	State     string
	Monitor   string
	FirstTime string
	LastTime  string
}

func addGetSensorListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_sensor_list",
		mcp.WithDescription("get sensor list from TWSNMP"),
		mcp.WithString("state_filter",
			mcp.Description(
				`state_filter uses a regular expression to specify search criteria for sensor state names.
If blank, all sensor are searched.
State names can be "normal","warn","low","high","repair","unknown"
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		state := makeRegexFilter(request.GetString("state_filter", ""))
		list := []mcpSensorEnt{}
		datastore.ForEachSensors(func(s *datastore.SensorEnt) bool {
			if state != nil && !state.MatchString(s.State) {
				return true
			}
			if s.Ignore {
				return true
			}
			monitor := ""
			if len(s.Monitors) > 0 {
				i := len(s.Monitors) - 1
				monitor = fmt.Sprintf("cpu=%.02f%% mem=%.02f%% load=%.02f%% process=%d",
					s.Monitors[i].CPU,
					s.Monitors[i].Mem,
					s.Monitors[i].Load,
					s.Monitors[i].Process,
				)
			}
			list = append(list, mcpSensorEnt{
				ID:        s.ID,
				Host:      s.Host,
				Type:      s.Type,
				State:     s.State,
				Total:     s.Total,
				Send:      s.Send,
				Monitor:   monitor,
				FirstTime: time.Unix(0, s.FirstTime).Format(time.RFC3339Nano),
				LastTime:  time.Unix(0, s.LastTime).Format(time.RFC3339Nano),
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
