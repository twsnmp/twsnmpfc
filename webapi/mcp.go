package webapi

import (
	"context"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/mark3labs/mcp-go/server"
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
	addSearchSyslogTool(s)
	addGetSensorListTool(s)
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

// makeRegexFilter
func makeRegexFilter(s string) *regexp.Regexp {
	if s != "" {
		if f, err := regexp.Compile(s); err == nil && f != nil {
			return f
		}
	}
	return nil
}
