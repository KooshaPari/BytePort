//go:build integration

package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/byteport/api/internal/infrastructure/clients"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEngineDeploymentHandler_Integration exercises the full HTTP path
// through engine_deployment_handler.go::Deploy against a real UDS-bound
// engine daemon (or a Go mock speaking the same HTTP contract).
//
// Gated by the "integration" build tag. To run:
//
//	go test -tags=integration -run TestEngineDeploymentHandler_Integration \
//	    ./internal/infrastructure/http/handlers/
//
// This complements engine_integration_test.go in clients/ which only
// exercises the EngineDaemonClient directly. The path here proves that
// the Gin handler + the client correctly compose end-to-end and that
// a deployment_id flows back through the HTTP layer.
func TestEngineDeploymentHandler_Integration(t *testing.T) {
	// -----------------------------------------------------------------------
	// 1. Start a mock engine daemon on a UDS socket.
	// -----------------------------------------------------------------------
	gin.SetMode(gin.TestMode)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /deploy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"deployment_id":"handler-int-deploy-1","status":"deploying"}`))
	})
	mux.HandleFunc("POST /deployments/{id}/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"stopped"}`))
	})

	ts := httptest.NewUnstartedServer(mux)
	dir, err := os.MkdirTemp("", "bp-handler-int-*")
	require.NoError(t, err)
	socketPath := filepath.Join(dir, "engine.sock")
	defer os.RemoveAll(dir)

	ln, err := net.Listen("unix", socketPath)
	require.NoError(t, err)
	ts.Listener = ln
	ts.Start()
	defer ts.Close()

	// -----------------------------------------------------------------------
	// 2. Wire the handler to the UDS-backed daemon.
	// -----------------------------------------------------------------------
	t.Setenv("BYTEPORT_ENGINE_SOCKET", socketPath)
	client := clients.NewEngineDaemonClient()

	handler := NewEngineDeploymentHandler(client)
	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	// -----------------------------------------------------------------------
	// 3. POST a deploy through the full HTTP path.
	// -----------------------------------------------------------------------
	body := `{
		"name": "integration-handler-app",
		"user": {"id": "int-handler-user", "email": "int-handler@test.com"},
		"repository": "https://github.com/test/handler-int",
		"services": [{
			"name": "api",
			"path": "nginx:alpine",
			"port": 8080,
			"env": [{"key":"NODE_ENV","value":"production"}]
		}]
	}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/engine/deploy",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// -----------------------------------------------------------------------
	// 4. Assert the deployment_id flowed back through the handler.
	// -----------------------------------------------------------------------
	require.Equal(t, http.StatusCreated, w.Code,
		"handler should respond 201 when daemon accepts the deploy")

	var resp clients.DeployResult
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp),
		"response body should be parseable as DeployResult")
	require.NotEmpty(t, resp.DeploymentID,
		"deployment_id must be returned")
	assert.Equal(t, "handler-int-deploy-1", resp.DeploymentID,
		"deployment_id should match the mock daemon's response")

	// -----------------------------------------------------------------------
	// 5. Stop the deployment we just created, end-to-end.
	// -----------------------------------------------------------------------
	t.Log("Stopping deployment through handler")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST",
		"/api/v1/engine/deployments/"+resp.DeploymentID+"/stop", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code,
		"stop should succeed against mock daemon")

	// -----------------------------------------------------------------------
	// 6. Health probe through the handler to confirm connectivity.
	// -----------------------------------------------------------------------
	t.Log("Probing health through handler")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/engine/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code,
		"handler health probe should report healthy")
	assert.Contains(t, w.Body.String(), `"engine":"connected"`,
		"health body should report engine connected")

	// -----------------------------------------------------------------------
	// 7. Stop the mock; health should flip to unhealthy.
	// -----------------------------------------------------------------------
	ts.Close()
	os.RemoveAll(dir)
	// Give the OS a moment to release the socket file.
	time.Sleep(100 * time.Millisecond)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/engine/health", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code,
		"after daemon shutdown, health should be unhealthy")
}