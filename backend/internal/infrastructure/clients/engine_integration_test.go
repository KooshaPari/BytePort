//go:build integration

package clients

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEngineDaemon_Integration exercises the EngineDaemonClient over a real
// UDS connection against a mock server that speaks the same HTTP contract as
// the Rust byteport-engine daemon.
//
// This test is gated by the "integration" build tag and is not run during
// normal `go test ./...`. To run:
//
//	go test -tags=integration -run TestEngineDaemon_Integration ./internal/infrastructure/clients/
func TestEngineDaemon_Integration(t *testing.T) {
	// -----------------------------------------------------------------------
	// 1. Start a mock daemon on a UDS socket
	// -----------------------------------------------------------------------
	mux := http.NewServeMux()

	// Deploy: POST /deploy -> 201 + deployment_id
	mux.HandleFunc("POST /deploy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"deployment_id":"int-deploy-1"}`))
	})

	// Stop: POST /deployments/{id}/stop -> 200
	mux.HandleFunc("POST /deployments/{id}/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"stopped"}`))
	})

	ts := httptest.NewUnstartedServer(mux)
	dir, err := os.MkdirTemp("", "bp-int-engine-*")
	require.NoError(t, err)
	socketPath := filepath.Join(dir, "engine.sock")

	ln, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	ts.Listener = ln
	ts.Start()
	defer ts.Close()
	defer os.RemoveAll(dir)

	// -----------------------------------------------------------------------
	// 2. Create the client pointed at the mock socket
	// -----------------------------------------------------------------------
	t.Setenv("BYTEPORT_ENGINE_SOCKET", socketPath)
	client := NewEngineDaemonClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// -----------------------------------------------------------------------
	// 3. Health check — daemon should be reachable
	// -----------------------------------------------------------------------
	t.Log("Step 1: health check")
	assert.True(t, client.Health(ctx), "engine daemon should be reachable")

	// -----------------------------------------------------------------------
	// 4. Deploy — send a deploy request, expect a deployment ID back
	// -----------------------------------------------------------------------
	t.Log("Step 2: deploy")
	result, err := client.Deploy(ctx, &DeployRequest{
		Name:       "integration-test",
		User:       DeployUser{ID: "int-user", Email: "int@test.com"},
		Repository: "https://github.com/test/int",
		Services: []ServiceEntry{{
			Name: "api",
			Path: "nginx:alpine",
			Port: 8080,
			Env: []EnvEntry{
				{Key: "NODE_ENV", Value: "production"},
			},
		}},
	})
	require.NoError(t, err, "deploy should succeed")
	require.NotNil(t, result, "deploy result should not be nil")
	assert.Equal(t, "int-deploy-1", result.DeploymentID,
		"deployment ID should match mock response")

	// -----------------------------------------------------------------------
	// 5. Stop — stop the deployment we just created
	// -----------------------------------------------------------------------
	t.Log("Step 3: stop")
	err = client.Stop(ctx, result.DeploymentID)
	require.NoError(t, err, "stop should succeed")

	// -----------------------------------------------------------------------
	// 6. Health after use — daemon should still be reachable
	// -----------------------------------------------------------------------
	t.Log("Step 4: health check after traffic")
	assert.True(t, client.Health(ctx), "engine daemon should still be reachable")

	// -----------------------------------------------------------------------
	// 7. Unavailable — close the mock, health should flip
	// -----------------------------------------------------------------------
	t.Log("Step 5: health after daemon shutdown")
	ts.Close()
	os.RemoveAll(dir)

	// Give the OS a moment to release the socket file.
	time.Sleep(100 * time.Millisecond)
	assert.False(t, client.Health(ctx),
		"health should return false when daemon is gone")
}
