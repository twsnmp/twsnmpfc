package webapi

import (
	"context"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/mark3labs/mcp-go/server"
)

var mcpSSEServer *server.SSEServer
var mcpStreamableHTTPServer *server.StreamableHTTPServer
var mcpAllow sync.Map

func startMCPServer(e *echo.Echo, p *WebAPI) {
	log.Printf("start mcp server mode=%s,form=%s", p.MCPMode, p.MCPFrom)
	mcpAllow.Store("127.0.0.1", true)
	for _, ip := range strings.Split(p.MCPFrom, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			mcpAllow.Store(ip, true)
		}
	}
	// Create MCP Server
	s := server.NewMCPServer(
		"TWSNMP MCP Server",
		p.Version,
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)
	// Add tools to MCP server
	// map
	addGetNodeListTool(s)
	addGetNetworkListTool(s)
	addGetPollingListTool(s)
	addGetPollingLogTool(s)
	addDoPingtTool(s)
	addGetMIBTreeTool(s)
	addSNMPWalkTool(s)
	addAddNodeTool(s)
	addUpdateNodeTool(s)
	// log
	addSearchEventLogTool(s)
	addAddEventLogTool(s)
	addSearchSyslogTool(s)
	addGetSyslogSummaryTool(s)
	addSearchSNMPTrapLogTool(s)
	// report
	addGetSensorListTool(s)
	addGetMACAddressListTool(s)
	addGetIPAddressListTool(s)
	addGetWifiAPListTool(s)
	addGetBluetoothDeviceListTool(s)
	addGetServerCertificateListTool(s)
	addGetResourceMonitorListTool(s)
	switch p.MCPMode {
	case "sse":
		mcpSSEServer = server.NewSSEServer(s)
		e.Any("/sse", func(c echo.Context) error {
			if !mcpCheckFromAddress(c) {
				return echo.ErrUnauthorized
			}
			mcpSSEServer.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/message", func(c echo.Context) error {
			if !mcpCheckFromAddress(c) {
				return echo.ErrUnauthorized
			}
			mcpSSEServer.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
	case "auth":
		mcpStreamableHTTPServer = server.NewStreamableHTTPServer(s)
		e.Any("/mcp", func(c echo.Context) error {
			if !mcpCheckFromAddress(c) {
				return echo.ErrUnauthorized
			}
			mcpStreamableHTTPServer.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		}, echojwt.JWT([]byte(p.Password)))
	default:
		mcpStreamableHTTPServer = server.NewStreamableHTTPServer(s)
		e.Any("/mcp", func(c echo.Context) error {
			if !mcpCheckFromAddress(c) {
				return echo.ErrUnauthorized
			}
			mcpStreamableHTTPServer.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
	}
}

func stopMCPServer(ctx context.Context) {
	if mcpSSEServer != nil {
		mcpSSEServer.Shutdown(ctx)
	}
	if mcpStreamableHTTPServer != nil {
		mcpStreamableHTTPServer.Shutdown(ctx)
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
