package mcp

import (
	"encoding/json"
	"testing"
)

func TestGetTools(t *testing.T) {
	InitTools()
	tools := GetTools()
	if len(tools) == 0 {
		t.Fatal("expected at least one registered tool")
	}

	found := false
	for _, tool := range tools {
		if tool.Name == "byteport_health" {
			found = true
			if tool.Description == "" {
				t.Error("byteport_health tool has empty description")
			}
			if len(tool.InputSchema) == 0 {
				t.Error("byteport_health tool has empty inputSchema")
			}
			break
		}
	}
	if !found {
		t.Fatal("byteport_health tool not found in registry")
	}
}

func TestHandleTool(t *testing.T) {
	InitTools()

	result, err := HandleTool("byteport_health", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("HandleTool failed: %v", err)
	}

	data, ok := result["data"]
	if !ok {
		t.Fatal("result missing 'data' field")
	}

	dataMap, ok := data.(map[string]any)
	if !ok {
		t.Fatal("data is not a map")
	}

	if dataMap["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", dataMap["status"])
	}
	if dataMap["service"] != "BytePort" {
		t.Errorf("expected service 'BytePort', got %v", dataMap["service"])
	}
	if _, ok := dataMap["timestamp"]; !ok {
		t.Error("data missing 'timestamp' field")
	}
}

func TestHandleToolUnknown(t *testing.T) {
	InitTools()

	_, err := HandleTool("nonexistent_tool", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestJSONRPCInitialize(t *testing.T) {
	server := NewServer(nil)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
		Params:  json.RawMessage(`{}`),
	}

	resp := server.handleRequest(req)
	if resp.Error != nil {
		t.Fatalf("initialize returned error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("initialize returned nil result")
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocolVersion '2024-11-05', got %v", result["protocolVersion"])
	}
}

func TestJSONRPCToolsList(t *testing.T) {
	InitTools()
	server := NewServer(nil)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/list",
		Params:  json.RawMessage(`{}`),
	}

	resp := server.handleRequest(req)
	if resp.Error != nil {
		t.Fatalf("tools/list returned error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("tools/list returned nil result")
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatal("result['tools'] is not an array")
	}
	if len(tools) == 0 {
		t.Fatal("expected at least one tool")
	}
}

func TestJSONRPCToolsCall(t *testing.T) {
	InitTools()
	server := NewServer(nil)

	params := map[string]any{
		"name":      "byteport_health",
		"arguments": map[string]any{},
	}
	paramsRaw, _ := json.Marshal(params)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`3`),
		Method:  "tools/call",
		Params:  paramsRaw,
	}

	resp := server.handleRequest(req)
	if resp.Error != nil {
		t.Fatalf("tools/call returned error: %v", resp.Error)
	}
}

func TestJSONRPCMethodNotFound(t *testing.T) {
	server := NewServer(nil)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`99`),
		Method:  "nonexistent_method",
		Params:  json.RawMessage(`{}`),
	}

	resp := server.handleRequest(req)
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
}

func TestJSONRPCRequestIntegration(t *testing.T) {
	// Simulate the full conversation: initialize → tools/list → tools/call
	InitTools()
	server := NewServer(nil)

	// Step 1: Initialize
	initResp := server.handleRequest(jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
	})
	if initResp.Error != nil {
		t.Fatalf("initialize failed: %v", initResp.Error)
	}

	// Step 2: tools/list
	listResp := server.handleRequest(jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`2`),
		Method:  "tools/list",
	})
	if listResp.Error != nil {
		t.Fatalf("tools/list failed: %v", listResp.Error)
	}

	// Step 3: tools/call
	params, _ := json.Marshal(map[string]any{
		"name":      "byteport_health",
		"arguments": map[string]any{},
	})
	callResp := server.handleRequest(jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`3`),
		Method:  "tools/call",
		Params:  params,
	})
	if callResp.Error != nil {
		t.Fatalf("tools/call failed: %v", callResp.Error)
	}
}
