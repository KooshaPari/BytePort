package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// ErrDaemonUnavailable is returned when the engine daemon UDS cannot be dialled.
var ErrDaemonUnavailable = errors.New("engine daemon unavailable")

// DefaultEngineSocketPath is the default UDS path for the byteport-engine daemon.
const DefaultEngineSocketPath = "/tmp/byteport-engine.sock"

// ---------------------------------------------------------------------------
// Wire types (mirror the daemon's JSON contract)
// ---------------------------------------------------------------------------

// DeployRequest is the payload for creating a deployment.
type DeployRequest struct {
	Name       string         `json:"name"`
	User       DeployUser     `json:"user"`
	Repository string         `json:"repository"`
	Services   []ServiceEntry `json:"services"`
}

// DeployUser identifies the user who owns the deployment.
type DeployUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// ServiceEntry describes one service within a deployment.
type ServiceEntry struct {
	Name string     `json:"name"`
	Path string     `json:"path"`
	Port uint16     `json:"port"`
	Env  []EnvEntry `json:"env"`
}

// EnvEntry is a single environment variable.
type EnvEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DeployResult is returned by the daemon on successful deployment.
type DeployResult struct {
	DeploymentID string `json:"deployment_id"`
}

// daemonError mirrors the daemon's JSON error body.
type daemonError struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// ---------------------------------------------------------------------------
// EngineDaemonClient
// ---------------------------------------------------------------------------

// EngineDaemonClient speaks HTTP/1.1 over a Unix Domain Socket to the Rust
// byteport-engine daemon.
type EngineDaemonClient struct {
	socketPath string
	httpClient *http.Client
}

// NewEngineDaemonClient creates a new client.
//
// Socket path resolution order:
//  1. BYTEPORT_ENGINE_SOCKET env var
//  2. DefaultEngineSocketPath (/tmp/byteport-engine.sock)
func NewEngineDaemonClient() *EngineDaemonClient {
	socketPath := os.Getenv("BYTEPORT_ENGINE_SOCKET")
	if socketPath == "" {
		socketPath = DefaultEngineSocketPath
	}
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, 2*time.Second)
		},
	}
	return &EngineDaemonClient{
		socketPath: socketPath,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
	}
}

// SocketPath returns the configured UDS path.
func (c *EngineDaemonClient) SocketPath() string { return c.socketPath }

// Deploy sends a deployment request to the engine daemon.
// Returns ErrDaemonUnavailable when the daemon is unreachable.
func (c *EngineDaemonClient) Deploy(ctx context.Context, req *DeployRequest) (*DeployResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal deploy request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://byteport-engine/deploy", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create deploy request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if isDialError(err) {
			return nil, ErrDaemonUnavailable
		}
		return nil, fmt.Errorf("deploy request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, parseDaemonError(resp)
	}

	var result DeployResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode deploy response: %w", err)
	}
	return &result, nil
}

// Stop sends a stop/destroy request for a deployment.
// Returns ErrDaemonUnavailable when the daemon is unreachable.
func (c *EngineDaemonClient) Stop(ctx context.Context, id string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"http://byteport-engine/deployments/"+id+"/stop", nil)
	if err != nil {
		return fmt.Errorf("create stop request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if isDialError(err) {
			return ErrDaemonUnavailable
		}
		return fmt.Errorf("stop request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseDaemonError(resp)
	}
	return nil
}

// Health checks whether the engine daemon is reachable by dialling the UDS
// socket.  A successful dial means the listener is accepting connections.
func (c *EngineDaemonClient) Health(ctx context.Context) bool {
	dialer := net.Dialer{Timeout: 1 * time.Second}
	conn, err := dialer.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// isDialError returns true when the error indicates the UDS socket is
// unreachable (connection refused or no such file/directory).
func isDialError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "no such file") ||
		strings.Contains(s, "connect: connection refused")
}

// parseDaemonError reads the response body and builds a Go error from it.
// When the body is valid JSON with an "error" field, that message (and
// optional "details") are surfaced. Fallback includes the status code and
// raw body.
func parseDaemonError(resp *http.Response) error {
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("engine daemon returned status %d", resp.StatusCode)
	}
	var de daemonError
	if json.Unmarshal(body, &de) == nil && de.Error != "" {
		if de.Details != "" {
			return fmt.Errorf("engine daemon error (%d): %s — %s",
				resp.StatusCode, de.Error, de.Details)
		}
		return fmt.Errorf("engine daemon error (%d): %s",
			resp.StatusCode, de.Error)
	}
	return fmt.Errorf("engine daemon returned status %d: %s",
		resp.StatusCode, string(body))
}
