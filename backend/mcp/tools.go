package mcp

import (
	"encoding/json"
	"fmt"
)

// ---------------------------------------------------------------------------
// MCP tool definitions for BytePort.
//
// Each tool wraps one or more REST API calls behind a named tool that an MCP
// host (Claude Desktop, Codex CLI, etc.) can invoke via JSON-RPC over stdio.
// ---------------------------------------------------------------------------

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolHandler is a function that processes a tool invocation.
type ToolHandler func(args json.RawMessage) (ToolResult, error)

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

// toolRegistry maps tool names to their definition and handler.
var toolRegistry []struct {
	def     Tool
	handler ToolHandler
}

// RegisterTool adds a tool to the registry.
func RegisterTool(name, desc string, schema any, handler ToolHandler) {
	schemaRaw, err := json.Marshal(schema)
	if err != nil {
		panic(fmt.Sprintf("mcp: failed to marshal schema for tool %q: %v", name, err))
	}
	toolRegistry = append(toolRegistry, struct {
		def     Tool
		handler ToolHandler
	}{
		def: Tool{
			Name:        name,
			Description: desc,
			InputSchema: schemaRaw,
		},
		handler: handler,
	})
}

// GetTools returns the list of registered tool definitions.
func GetTools() []Tool {
	tools := make([]Tool, len(toolRegistry))
	for i, t := range toolRegistry {
		tools[i] = t.def
	}
	return tools
}

// HandleTool dispatches a tool call to the registered handler.
func HandleTool(name string, args json.RawMessage) (ToolResult, error) {
	for _, t := range toolRegistry {
		if t.def.Name == name {
			return t.handler(args)
		}
	}
	return nil, fmt.Errorf("unknown tool: %s", name)
}

// ---------------------------------------------------------------------------
// Tool input schemas (JSON Schema)
// ---------------------------------------------------------------------------

type noArgsSchema struct{}

var NoArgs = noArgsSchema{}

// HealthArgs is the input schema for the health tool.
var HealthArgs = struct {
}{}

// ---------------------------------------------------------------------------
// Tool initialisation — called during server startup.
// ---------------------------------------------------------------------------

// InitTools registers all BytePort MCP tools.
func InitTools() {
	// Health check (no API call needed)
	RegisterTool(
		"byteport_health",
		"Check BytePort service health and readiness.",
		struct {
			Type       string `json:"type"`
			Properties struct{} `json:"properties"`
		}{
			Type:       "object",
			Properties: struct{}{},
		},
		func(_ json.RawMessage) (ToolResult, error) {
			return OkResult(map[string]any{
				"status":    "ok",
				"service":   "BytePort",
				"timestamp": timeNow(),
			}), nil
		},
	)

	// Register additional tools here as the REST API grows.
	// Example:
	//   byteport_create_package
	//   byteport_list_packages
	//   byteport_get_package
	//   byteport_upload_file
	//   byteport_list_snapshots
}
