package middleware

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// DefaultSocketPath exposes the default UDS path so tests can reference it.
const DefaultSocketPath = "/tmp/omniroute/routed.sock"

// UDSClient is an http.Client that speaks HTTP/1.1 over a Unix Domain Socket.
// Package-level singleton so the transport is reused (connection pooling).
var UDSClient = &http.Client{
	Timeout: 60 * time.Second,
}

// UDSProxy forwards /v1/chat/completions requests to the Rust data plane via
// Unix Domain Socket.  The proxy dials the UDS once per request, writes the
// raw HTTP/1.1 request, and reads the response — no client-side pool needed
// because the data plane reuses connections natively.
//
// Socket path resolution order:
//  1. OMNIROUTE_DATA_PLANE_SOCKET env var (explicit override).
//  2. $XDG_RUNTIME_DIR/omniroute/routed.sock (XDG portable default).
//  3. /tmp/omniroute/routed.sock (POSIX fallback when XDG_RUNTIME_DIR is unset).
//
// Headers forwarded on every proxied request:
//   - X-OmniRoute-Provider: provider id (e.g. "openai", "anthropic").
//   - Authorization: Bearer token from the client (Rust data plane does NOT read
//     env vars or files; it expects a credential per request via this header per
//     `omniroute-core::credentials::Context`).
//
// Response is passed through (status + body + Content-Type) verbatim so the
// downstream client sees an identical contract to the direct Rust client.
func UDSProxy() gin.HandlerFunc {
	socketPath := resolveSocket()

	// Build a transport that always dials the UDS socket.
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, 2*time.Second)
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	return func(c *gin.Context) {
		// Proxy any path that ends in /chat/completions — the Gin server
		// mounts at /api/v1/* and the Rust data plane lives at /v1/*; the
		// target URL is always /v1/chat/completions regardless of prefix.
		if !strings.HasSuffix(c.Request.URL.Path, "/chat/completions") {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to read request body",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		req, err := http.NewRequestWithContext(
			c.Request.Context(),
			"POST",
			"http://routed/v1/chat/completions",
			bytes.NewReader(body),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to create upstream request",
				"details": err.Error(),
			})
			c.Abort()
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-OmniRoute-Provider", c.GetHeader("X-OmniRoute-Provider"))
		req.Header.Set("Authorization", c.GetHeader("Authorization"))

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream request failed",
				"details": err.Error(),
			})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to read upstream response",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		c.Data(resp.StatusCode, "application/json", respBody)
	}
}

// resolveSocket returns the UDS path for the Rust data plane. Override order:
// env > XDG_RUNTIME_DIR > /tmp.
func resolveSocket() string {
	if path := os.Getenv("OMNIROUTE_DATA_PLANE_SOCKET"); path != "" {
		return path
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = "/tmp"
	}
	return runtimeDir + "/omniroute/routed.sock"
}
