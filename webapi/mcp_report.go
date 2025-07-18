package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/twsnmp/twsnmpfc/datastore"
)

// get_sensor_list tool
type mcpSensorEnt struct {
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
State names can be "normal","warn","low","high"
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

type mcpMACEnt struct {
	MAC       string
	Name      string
	IP        string
	Vendor    string
	Score     float64
	Penalty   int64
	FirstTime string
	LastTime  string
}

func addGetMACAddressListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_mac_address_list",
		mcp.WithDescription("get MAC address list from TWSNMP"),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		list := []mcpMACEnt{}
		datastore.ForEachDevices(func(d *datastore.DeviceEnt) bool {
			if !d.ValidScore {
				return true
			}
			name := d.Name
			if name == d.IP && d.NodeID != "" {
				if n := datastore.GetNode(d.NodeID); n != nil {
					name = n.Name
				}
			}
			list = append(list, mcpMACEnt{
				MAC:       d.ID,
				Name:      name,
				IP:        d.IP,
				Vendor:    d.Vendor,
				Score:     d.Score,
				Penalty:   d.Penalty,
				FirstTime: time.Unix(0, d.FirstTime).Format(time.RFC3339Nano),
				LastTime:  time.Unix(0, d.LastTime).Format(time.RFC3339Nano),
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

type mcpIPEnt struct {
	IP        string
	MAC       string
	Name      string
	Location  string
	Vendor    string
	Count     int64
	Change    int64
	Score     float64
	Penalty   int64
	FirstTime string
	LastTime  string
}

func addGetIPAddressListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_ip_address_list",
		mcp.WithDescription("get IP address list from TWSNMP"),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		list := []mcpIPEnt{}
		datastore.ForEachIPReport(func(i *datastore.IPReportEnt) bool {
			if !i.ValidScore {
				return true
			}
			name := i.Name
			if name == i.IP && i.NodeID != "" {
				if n := datastore.GetNode(i.NodeID); n != nil {
					name = n.Name
				}
			}
			list = append(list, mcpIPEnt{
				IP:        i.IP,
				MAC:       i.MAC,
				Name:      name,
				Vendor:    i.Vendor,
				Score:     i.Score,
				Penalty:   i.Penalty,
				Location:  i.Loc,
				Count:     i.Count,
				Change:    i.Change,
				FirstTime: time.Unix(0, i.FirstTime).Format(time.RFC3339Nano),
				LastTime:  time.Unix(0, i.LastTime).Format(time.RFC3339Nano),
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

type mcpWifiAPEnt struct {
	Host      string
	BSSID     string
	SSID      string
	RSSI      int
	Channel   string
	Vendor    string
	Info      string
	Count     int
	Change    int
	FirstTime string
	LastTime  string
}

func addGetWifiAPListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_wifi_ap_list",
		mcp.WithDescription("get Wifi access point list from TWSNMP"),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		list := []mcpWifiAPEnt{}
		datastore.ForEachWifiAP(func(e *datastore.WifiAPEnt) bool {
			rssi := 0
			if len(e.RSSI) > 0 {
				rssi = e.RSSI[len(e.RSSI)-1].Value
			}
			list = append(list, mcpWifiAPEnt{
				Host:      e.Host,
				BSSID:     e.BSSID,
				SSID:      e.SSID,
				Channel:   e.Channel,
				Vendor:    e.Vendor,
				Info:      e.Info,
				RSSI:      rssi,
				Count:     e.Count,
				Change:    e.Change,
				FirstTime: time.Unix(0, e.FirstTime).Format(time.RFC3339Nano),
				LastTime:  time.Unix(0, e.LastTime).Format(time.RFC3339Nano),
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

type mcpBluetoothDeviceEnt struct {
	Host        string
	Address     string
	Name        string
	AddressType string
	RSSI        int
	Info        string
	Vendor      string
	Count       int64
	FirstTime   string
	LastTime    string
}

func addGetBluetoothDeviceListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_bluetooth_device_list",
		mcp.WithDescription("get bluetooth device list from TWSNMP"),
		mcp.WithBoolean("public_address_only",
			mcp.DefaultBool(true),
			mcp.Description("List only devices with Bluetooth address type Public")))
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		onlyPublic := request.GetBool("public_address_only", true)
		list := []mcpBluetoothDeviceEnt{}
		datastore.ForEachBlueDevice(func(b *datastore.BlueDeviceEnt) bool {
			if onlyPublic && !strings.Contains(b.AddressType, "Public") {
				return true
			}
			rssi := 0
			if len(b.RSSI) > 0 {
				rssi = b.RSSI[len(b.RSSI)-1].Value
			}
			list = append(list, mcpBluetoothDeviceEnt{
				Host:        b.Host,
				Address:     b.Address,
				AddressType: b.AddressType,
				RSSI:        rssi,
				Info:        b.Info,
				Vendor:      b.Vendor,
				Count:       b.Count,
				FirstTime:   time.Unix(0, b.FirstTime).Format(time.RFC3339Nano),
				LastTime:    time.Unix(0, b.LastTime).Format(time.RFC3339Nano),
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
