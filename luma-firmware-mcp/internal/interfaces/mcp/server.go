package mcp

import "context"

// Server is the transport boundary for the future MCP implementation.
type Server interface {
 Start(context.Context) error
 Shutdown(context.Context) error
}
