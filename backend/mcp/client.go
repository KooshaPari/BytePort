package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ---------------------------------------------------------------------------
// BytePort MCP Client — talks to the BytePort REST API over HTTP.
// ---------------------------------------------------------------------------

// ClientConfig holds connection parameters for the BytePort API.
type ClientConfig struct {
	BaseURL    string
	AuthToken  string
	HTTPClient *http.Client
}

// DefaultConfig returns a ClientConfig pointing at localhost:8080.
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		BaseURL:   "http://localhost:8080",
		AuthToken: os.Getenv("BYTEPORT_API_TOKEN"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Client is a lightweight HTTP client for the BytePort REST API.
type Client struct {
	cfg *ClientConfig
}

// NewClient creates a new BytePort API client.
func NewClient(cfg *ClientConfig) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Client{cfg: cfg}
}

// doJSON performs an HTTP request and unmarshals the JSON response into dst.
func (c *Client) doJSON(method, path string, body, dst any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.cfg.BaseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.AuthToken)
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Tool handler helpers
// ---------------------------------------------------------------------------

// ToolResult is a generic JSON-serializable result for MCP tool responses.
type ToolResult map[string]any

// ErrResult creates a ToolResult with an error field.
func ErrResult(msg string) ToolResult {
	return ToolResult{"error": msg}
}

// OkResult creates a ToolResult with a success field and data.
func OkResult(data any) ToolResult {
	return ToolResult{"success": true, "data": data}
}
