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

	"github.com/byteport/api/internal/infrastructure/clients"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// startMockDaemonForHandler starts an httptest server on a UDS that
// responds to engine daemon endpoints used by the handler tests.
func startMockDaemonForHandler(t *testing.T) (socketPath string, cleanup func()) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /deploy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"deployment_id":"handler-deploy-1"}`))
	})
	mux.HandleFunc("POST /deployments/{id}/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"stopped"}`))
	})

	ts := httptest.NewUnstartedServer(mux)
	dir, err := os.MkdirTemp("", "bp-handler-*")
	require.NoError(t, err)
	socketPath = filepath.Join(dir, "handler.sock")

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

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestEngineDeploymentHandler_New(t *testing.T) {
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)
	assert.NotNil(t, handler)
	assert.Same(t, client, handler.client)
}

func TestEngineDeploymentHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)

	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	routes := r.Routes()
	found := map[string]bool{
		"POST /api/v1/engine/deploy":                  false,
		"POST /api/v1/engine/deployments/:id/stop":    false,
		"GET /api/v1/engine/health":                   false,
	}

	for _, rt := range routes {
		key := rt.Method + " " + rt.Path
		if _, ok := found[key]; ok {
			found[key] = true
		}
	}

	for key, v := range found {
		assert.True(t, v, "route %s not registered", key)
	}
}

func TestEngineDeploymentHandler_Deploy_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDaemonForHandler(t)
	defer cleanup()

	t.Setenv("BYTEPORT_ENGINE_SOCKET", socketPath)
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)

	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	body := `{
		"name": "test-app",
		"user": {"id": "u1", "email": "u1@test.com"},
		"repository": "https://github.com/test/app",
		"services": [{"name": "web", "path": "nginx:latest", "port": 80, "env": []}]
	}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/engine/deploy",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp clients.DeployResult
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "handler-deploy-1", resp.DeploymentID)
}

func TestEngineDeploymentHandler_Deploy_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDaemonForHandler(t)
	defer cleanup()

	t.Setenv("BYTEPORT_ENGINE_SOCKET", socketPath)
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)

	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	// Invalid JSON
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/engine/deploy",
		strings.NewReader(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
}

func TestEngineDeploymentHandler_Deploy_DaemonUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("BYTEPORT_ENGINE_SOCKET", "/nonexistent/missing.sock")
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)

	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	body := `{
		"name": "test",
		"user": {"id": "u1", "email": "u1@test.com"},
		"repository": "https://github.com/test/app",
		"services": [{"name": "web", "path": "nginx:latest", "port": 80, "env": []}]
	}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/engine/deploy",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "engine daemon unavailable")
}

func TestEngineDeploymentHandler_Stop_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDaemonForHandler(t)
	defer cleanup()

	t.Setenv("BYTEPORT_ENGINE_SOCKET", socketPath)
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)

	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/engine/deployments/deploy-99/stop", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"stopped"`)
}

func TestEngineDeploymentHandler_Stop_DaemonUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("BYTEPORT_ENGINE_SOCKET", "/nonexistent/missing.sock")
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)

	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/engine/deployments/deploy-99/stop", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "engine daemon unavailable")
}

func TestEngineDeploymentHandler_Health_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDaemonForHandler(t)
	defer cleanup()

	t.Setenv("BYTEPORT_ENGINE_SOCKET", socketPath)
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)

	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/engine/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"healthy"`)
	assert.Contains(t, w.Body.String(), `"engine":"connected"`)
}

func TestEngineDeploymentHandler_Health_Unhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("BYTEPORT_ENGINE_SOCKET", "/nonexistent/missing.sock")
	client := clients.NewEngineDaemonClient()
	handler := NewEngineDeploymentHandler(client)

	r := gin.New()
	grp := r.Group("/api/v1")
	handler.RegisterRoutes(grp)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/engine/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"unhealthy"`)
	assert.Contains(t, w.Body.String(), `"engine":"unreachable"`)
}
