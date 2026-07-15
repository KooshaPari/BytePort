package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// startMockEngineDaemon starts an httptest server listening on a UDS that
// mimics the byteport-engine daemon's HTTP contract.
func startMockEngineDaemon(t *testing.T) (socketPath string, cleanup func()) {
	t.Helper()

	mux := http.NewServeMux()

	// POST /deploy
	mux.HandleFunc("POST /deploy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"deployment_id":"mock-deploy-1"}`))
	})

	// POST /deployments/{id}/stop
	mux.HandleFunc("POST /deployments/{id}/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"stopped"}`))
	})

	// POST /deploy — error case for tests that need it
	mux.HandleFunc("POST /deploy-error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid manifest","details":"missing services"}`))
	})

	ts := httptest.NewUnstartedServer(mux)

	dir, err := os.MkdirTemp("", "bp-engine-*")
	require.NoError(t, err)
	socketPath = filepath.Join(dir, "engine.sock")

	ln, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	ts.Listener = ln
	ts.Start()

	cleanup = func() {
		ts.Close()
		os.RemoveAll(dir)
	}
	t.Cleanup(cleanup)
	return
}

// newEngineClientForTest creates a client pointed at the given socket path
// and registers a t.Cleanup to restore the env after the test.
func newEngineClientForTest(t *testing.T, socketPath string) *EngineDaemonClient {
	t.Helper()
	t.Setenv("BYTEPORT_ENGINE_SOCKET", socketPath)
	return NewEngineDaemonClient()
}

// ---------------------------------------------------------------------------
// NewEngineDaemonClient / env resolution
// ---------------------------------------------------------------------------

func TestNewEngineDaemonClient_Defaults(t *testing.T) {
	t.Setenv("BYTEPORT_ENGINE_SOCKET", "")
	client := NewEngineDaemonClient()
	assert.Equal(t, DefaultEngineSocketPath, client.SocketPath())
}

func TestNewEngineDaemonClient_RespectsEnv(t *testing.T) {
	t.Setenv("BYTEPORT_ENGINE_SOCKET", "/custom/path.sock")
	client := NewEngineDaemonClient()
	assert.Equal(t, "/custom/path.sock", client.SocketPath())
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

func TestHealth_ReturnsTrueWhenSocketExists(t *testing.T) {
	socketPath, cleanup := startMockEngineDaemon(t)
	defer cleanup()

	client := newEngineClientForTest(t, socketPath)
	assert.True(t, client.Health(context.Background()))
}

func TestHealth_ReturnsFalseWhenSocketMissing(t *testing.T) {
	t.Setenv("BYTEPORT_ENGINE_SOCKET", "/nonexistent/missing.sock")
	client := NewEngineDaemonClient()
	assert.False(t, client.Health(context.Background()))
}

// ---------------------------------------------------------------------------
// Deploy
// ---------------------------------------------------------------------------

func TestDeploy_Success(t *testing.T) {
	socketPath, cleanup := startMockEngineDaemon(t)
	defer cleanup()

	client := newEngineClientForTest(t, socketPath)
	ctx := context.Background()

	result, err := client.Deploy(ctx, &DeployRequest{
		Name:       "my-app",
		User:       DeployUser{ID: "u1", Email: "u1@test.com"},
		Repository: "https://github.com/test/app",
		Services: []ServiceEntry{{
			Name: "web",
			Path: "nginx:latest",
			Port: 80,
		}},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "mock-deploy-1", result.DeploymentID)
}

func TestDeploy_ReturnsErrorOnBadRequest(t *testing.T) {
	socketPath, cleanup := startMockEngineDaemon(t)
	defer cleanup()

	client := newEngineClientForTest(t, socketPath)
	ctx := context.Background()

	// Reuse the mock's /deploy-error path by crafting the URL manually.
	// We test error parsing by pointing at the error handler directly.
	// The client normally POSTs to /deploy; we instead validate that a
	// non-201 with a JSON error body returns a well-formed error.
	body, _ := json.Marshal(&DeployRequest{Name: "broken"})
	req, err := http.NewRequestWithContext(ctx, "POST",
		"http://byteport-engine/deploy-error", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	err = parseDaemonError(resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid manifest")
	assert.Contains(t, err.Error(), "missing services")
}

func TestDeploy_ReturnsErrDaemonUnavailable(t *testing.T) {
	t.Setenv("BYTEPORT_ENGINE_SOCKET", "/nonexistent/missing.sock")
	client := NewEngineDaemonClient()
	ctx := context.Background()

	_, err := client.Deploy(ctx, &DeployRequest{Name: "test"})
	assert.ErrorIs(t, err, ErrDaemonUnavailable)
}

// ---------------------------------------------------------------------------
// Stop
// ---------------------------------------------------------------------------

func TestStop_Success(t *testing.T) {
	socketPath, cleanup := startMockEngineDaemon(t)
	defer cleanup()

	client := newEngineClientForTest(t, socketPath)
	ctx := context.Background()

	err := client.Stop(ctx, "deploy-42")
	assert.NoError(t, err)
}

func TestStop_ReturnsErrDaemonUnavailable(t *testing.T) {
	t.Setenv("BYTEPORT_ENGINE_SOCKET", "/nonexistent/missing.sock")
	client := NewEngineDaemonClient()
	ctx := context.Background()

	err := client.Stop(ctx, "deploy-42")
	assert.ErrorIs(t, err, ErrDaemonUnavailable)
}

// ---------------------------------------------------------------------------
// Health — env const
// ---------------------------------------------------------------------------

func TestDefaultEngineSocketPath(t *testing.T) {
	assert.Equal(t, "/tmp/byteport-engine.sock", DefaultEngineSocketPath)
}
