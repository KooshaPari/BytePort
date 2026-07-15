//go:build liveengine

package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEngineDaemon_LiveBinary exercises EngineDaemonClient against the real
// Rust byteport-engine daemon binary (compiled by cargo as
// target/debug/byteport-engine-daemon).
//
// This is the strongest end-to-end check we have that the Go HTTP wire
// contract actually matches the Rust daemon's axum router.
//
// Run with:
//
//	go test -tags liveengine -run TestEngineDaemon_LiveBinary \
//	  -v ./internal/infrastructure/clients/
//
// Skipped automatically when the binary is not present (CI).
func TestEngineDaemon_LiveBinary(t *testing.T) {
	daemonPath := findDaemonBinary(t)
	if daemonPath == "" {
		t.Skip("byteport-engine-daemon binary not found — build with `cargo build -p byteport-engine --bin byteport-engine-daemon`")
		return
	}

	// 1. Allocate a per-test UDS path.
	dir, err := os.MkdirTemp("", "bp-live-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	socketPath := filepath.Join(dir, "engine.sock")

	// 2. Spawn the daemon pointed at our socket.
	cmd := exec.Command(daemonPath)
	cmd.Env = append(os.Environ(), "BYTEPORT_ENGINE_SOCKET="+socketPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())
	defer func() {
		cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() { _ = cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			cmd.Process.Kill()
		}
	}()

	// 3. Wait for the socket to appear.
	require.Eventually(t, func() bool {
		conn, dialErr := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
		if dialErr != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, 5*time.Second, 100*time.Millisecond, "daemon never opened UDS socket")

	// 4. Create the client and exercise the full pipeline.
	t.Setenv("BYTEPORT_ENGINE_SOCKET", socketPath)
	client := NewEngineDaemonClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 4a. Health check.
	assert.True(t, client.Health(ctx), "health should succeed with live daemon")

	// 4b. Deploy — this hits the actual Rust handler that selects an engine,
	//     invokes the registry, and returns a deployment_id.
	result, err := client.Deploy(ctx, &DeployRequest{
		Name:       "live-binary-test",
		User:       DeployUser{ID: "live-user", Email: "live@test.local"},
		Repository: "https://github.com/test/live",
		Services: []ServiceEntry{{
			Name: "api",
			Path: "nginx:alpine",
			Port: 8080,
			Env: []EnvEntry{
				{Key: "NODE_ENV", Value: "test"},
			},
		}},
	})
	require.NoError(t, err, "deploy against live daemon should succeed")
	require.NotNil(t, result, "deploy result should be non-nil")
	assert.NotEmpty(t, result.DeploymentID, "live daemon should return a deployment_id")

	// 4c. Stop the deployment — also goes to live daemon.
	err = client.Stop(ctx, result.DeploymentID)
	require.NoError(t, err, "stop against live daemon should succeed")

	// 4d. Health is still true after traffic.
	assert.True(t, client.Health(ctx), "health should remain true after stop")

	t.Logf("Live daemon deploy/stop roundtrip OK with deployment_id=%s", result.DeploymentID)
}

// findDaemonBinary searches known workspace locations for the compiled
// daemon and returns the first match (or "" if not found).
func findDaemonBinary(t *testing.T) string {
	t.Helper()

	candidates := []string{
		// repo-relative default
		"../../../target/debug/byteport-engine-daemon",
		// absolute under the byteport repo
		"/Users/kooshapari/CodeProjects/Phenotype/repos/BytePort/target/debug/byteport-engine-daemon",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	// Fallback: walk PATH-style.
	if p, err := exec.LookPath("byteport-engine-daemon"); err == nil {
		return p
	}
	return ""
}

// _ = bytes.NewReader — keep encoding/json & bytes imports referenced even
// when the test short-circuits on Skip.
var (
	_ = bytes.NewReader
	_ = json.Marshal
	_ = http.MethodPost
)
