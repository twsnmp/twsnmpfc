package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/twsnmp/twsnmpfc/datastore"
)

// get_sensor_list tool

type mcpSensorMonitorEnt struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Load    float64 `json:"load"`
	Process int64   `json:"process"`
}
type mcpSensorEnt struct {
	Host      string              `json:"host"`
	Type      string              `json:"type"`
	Total     int64               `json:"total"`
	Send      int64               `json:"send"`
	State     string              `json:"state"`
	Monitor   mcpSensorMonitorEnt `json:"monitor"`
	FirstTime string              `json:"first_time"`
	LastTime  string              `json:"last_time"`
}
type mcpGetSensorListParams struct {
	StateFilter string `json:"state_filter" jsonschema:"state_filter uses a regular expression to specify search criteria for sensor state names(normal,warn,low,high,unknown).If blank, all sensors are searched."`
}

func mcpGetSensorList(ctx context.Context, req *mcp.CallToolRequest, args mcpGetSensorListParams) (*mcp.CallToolResult, any, error) {
	state := makeRegexFilter(args.StateFilter)
	list := []mcpSensorEnt{}
	datastore.ForEachSensors(func(s *datastore.SensorEnt) bool {
		if state != nil && !state.MatchString(s.State) {
			return true
		}
		if s.Ignore {
			return true
		}
		monitor := mcpSensorMonitorEnt{}
		if len(s.Monitors) > 0 {
			i := len(s.Monitors) - 1
			monitor.CPU = s.Monitors[i].CPU
			monitor.Memory = s.Monitors[i].Mem
			monitor.Load = s.Monitors[i].Load
			monitor.Process = s.Monitors[i].Process
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
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpMACEnt struct {
	MAC       string  `json:"mac"`
	Name      string  `json:"name"`
	IP        string  `json:"ip"`
	Vendor    string  `json:"vendor"`
	Score     float64 `json:"score"`
	Penalty   int64   `json:"penalty"`
	FirstTime string  `json:"first_time"`
	LastTime  string  `json:"last_time"`
}

func mcpGetMACAddressList(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
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
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpIPEnt struct {
	IP        string  `json:"ip"`
	MAC       string  `json:"mac"`
	Name      string  `json:"name"`
	Location  string  `json:"location"`
	Vendor    string  `json:"vendor"`
	Count     int64   `json:"count"`
	Change    int64   `json:"change"`
	Score     float64 `json:"score"`
	Penalty   int64   `json:"penalty"`
	FirstTime string  `json:"first_time"`
	LastTime  string  `json:"last_time"`
}

func mcpGetIPAddressList(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
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
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpWifiAPEnt struct {
	Host      string `json:"host"`
	BSSID     string `json:"bssid"`
	SSID      string `json:"ssid"`
	RSSI      int    `json:"rssi"`
	Channel   string `json:"channel"`
	Vendor    string `json:"vendor"`
	Info      string `json:"info"`
	Count     int    `json:"count"`
	Change    int    `json:"change"`
	FirstTime string `json:"first_time"`
	LastTime  string `json:"last_time"`
}

func mcpGetWifiAPList(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
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
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpBluetoothDeviceEnt struct {
	Host        string `json:"host"`
	Address     string `json:"address"`
	Name        string `json:"name"`
	AddressType string `json:"address_type"`
	RSSI        int    `json:"rssi"`
	Info        string `json:"info"`
	Vendor      string `json:"vendor"`
	Count       int64  `json:"count"`
	FirstTime   string `json:"first_time"`
	LastTime    string `json:"last_time"`
}

type mcpGetBluetoothDeviceListParams struct {
	PublicAddressOnly bool `json:"public_address_only" jsonschema:"List only devices with Bluetooth address type Public."`
}

func mcpGetBluetoothDeviceList(ctx context.Context, req *mcp.CallToolRequest, args mcpGetBluetoothDeviceListParams) (*mcp.CallToolResult, any, error) {
	onlyPublic := args.PublicAddressOnly
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
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpServerCertificateEnt struct {
	Server       string  `json:"server"`
	Port         uint16  `json:"port"`
	Subject      string  `json:"subject"`
	Issuer       string  `json:"issuer"`
	SerialNumber string  `json:"serial_number"`
	Verify       bool    `json:"verify"`
	NotAfter     string  `json:"not_after"`
	NotBefore    string  `json:"not_before"`
	Error        string  `json:"error"`
	Score        float64 `json:"score"`
	Penalty      int64   `json:"penalty"`
	FirstTime    string  `json:"first_time"`
	LastTime     string  `json:"last_time"`
}

func mcpGetServerCertificateList(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	list := []mcpServerCertificateEnt{}
	datastore.ForEachCerts(func(c *datastore.CertEnt) bool {
		list = append(list, mcpServerCertificateEnt{
			Server:       c.Target,
			Port:         c.Port,
			Subject:      c.Subject,
			Issuer:       c.Issuer,
			SerialNumber: c.SerialNumber,
			Verify:       c.Verify,
			NotBefore:    time.Unix(c.NotBefore, 0).Format(time.RFC3339),
			NotAfter:     time.Unix(c.NotAfter, 0).Format(time.RFC3339),
			Error:        c.Error,
			Score:        c.Score,
			Penalty:      c.Penalty,
			FirstTime:    time.Unix(0, c.FirstTime).Format(time.RFC3339Nano),
			LastTime:     time.Unix(0, c.LastTime).Format(time.RFC3339Nano),
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

type mcpResourceMonitorEnt struct {
	Time   string `json:"time"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Swap   string `json:"swap"`
	Disk   string `json:"disk"`
	Load   string `json:"load"`
}

func mcpGetResourceMonitorList(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	list := []mcpResourceMonitorEnt{}
	skip := 30
	if len(datastore.MonitorDataes) < 120 {
		skip = 5
	}
	for i, m := range datastore.MonitorDataes {
		if i%skip != 0 {
			continue
		}
		list = append(list, mcpResourceMonitorEnt{
			Time:   time.Unix(m.At, 0).Format(time.RFC3339),
			CPU:    fmt.Sprintf("%.02f%%", m.CPU),
			Memory: fmt.Sprintf("%.02f%%", m.Mem),
			Swap:   fmt.Sprintf("%.02f%%", m.Swap),
			Disk:   fmt.Sprintf("%.02f%%", m.Disk),
			Load:   fmt.Sprintf("%.02f", m.Load),
		})
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
