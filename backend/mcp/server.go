package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// ---------------------------------------------------------------------------
// MCP JSON-RPC server over stdio transport.
//
// Implements the Model Context Protocol (MCP) specification for tool
// discovery and invocation. Communicates via stdin/stdout using newline-
// delimited JSON-RPC messages.
// ---------------------------------------------------------------------------

func timeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ---------------------------------------------------------------------------
// JSON-RPC types
// ---------------------------------------------------------------------------

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

// Server represents the MCP stdio server.
type Server struct {
	client *Client
}

// NewServer creates a new MCP server backed by the BytePort REST client.
func NewServer(client *Client) *Server {
	if client == nil {
		client = NewClient(nil)
	}
	return &Server{client: client}
}

// Run starts the MCP server, reading JSON-RPC requests from stdin and
// writing responses to stdout.
func (s *Server) Run() error {
	// Ensure tools are registered.
	if len(toolRegistry) == 0 {
		InitTools()
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB buffer

	logger := log.New(os.Stderr, "[mcp] ", log.LstdFlags)
	logger.Println("MCP server started (stdio transport)")

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			logger.Printf("invalid JSON-RPC request: %v", err)
			continue
		}

		response := s.handleRequest(req)
		respBytes, err := json.Marshal(response)
		if err != nil {
			logger.Printf("failed to marshal response: %v", err)
			continue
		}
		fmt.Println(string(respBytes))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stdin scanner error: %w", err)
	}
	return nil
}

func (s *Server) handleRequest(req jsonrpcRequest) jsonrpcResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "notifications/initialized":
		// No response expected for notifications.
		return jsonrpcResponse{}
	default:
		return s.errorResponse(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(req jsonrpcRequest) jsonrpcResponse {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "byteport-mcp",
			"version": "0.1.0",
		},
	}
	return s.resultResponse(req.ID, result)
}

func (s *Server) handleToolsList(req jsonrpcRequest) jsonrpcResponse {
	tools := GetTools()
	result := map[string]any{
		"tools": tools,
	}
	return s.resultResponse(req.ID, result)
}

func (s *Server) handleToolsCall(req jsonrpcRequest) jsonrpcResponse {
	var params struct {
		Name string          `json:"name"`
		Args json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.errorResponse(req.ID, -32602, fmt.Sprintf("Invalid params: %v", err))
	}

	toolResult, err := HandleTool(params.Name, params.Args)
	if err != nil {
		return s.errorResponse(req.ID, -32000, err.Error())
	}

	return s.resultResponse(req.ID, toolResult)
}

func (s *Server) resultResponse(id json.RawMessage, result any) jsonrpcResponse {
	raw, _ := json.Marshal(result)
	return jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  raw,
	}
}

func (s *Server) errorResponse(id json.RawMessage, code int, message string) jsonrpcResponse {
	return jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &jsonrpcError{
			Code:    code,
			Message: message,
		},
	}
}
