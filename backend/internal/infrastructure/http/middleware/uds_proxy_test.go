package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startMockDataPlane starts a real net/http server on a Unix Domain Socket.
// Uses httptest.NewUnstartedServer with a custom UDS listener so the mock
// behaves identically to the production Rust data plane (HTTP/1.1 over UDS).
func startMockDataPlane(t *testing.T) (socketPath string, cleanup func()) {
	t.Helper()

	handler := http.NewServeMux()
	handler.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hello"}}]}`))
	})
	handler.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	ts := httptest.NewUnstartedServer(handler)

	dir, err := os.MkdirTemp("/tmp", "bp-uds-*")
	require.NoError(t, err)
	socketPath = filepath.Join(dir, "mock.sock")

	ln, err := net.Listen("unix", socketPath)
	require.NoError(t, err)

	ts.Listener = ln
	ts.Start()
	t.Cleanup(func() {
		ts.Close()
		os.RemoveAll(dir)
	})

	return socketPath, func() {
		ts.Close()
		os.RemoveAll(dir)
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestUDSProxy_NonChatRoutesFallThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())
	r.GET("/api/v1/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/healthz", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestUDSProxy_ForwardsChatToDataPlane(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/completions", nil)
	req.Header.Set("X-OmniRoute-Provider", "anthropic")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"content":"hello"`)
}

func TestUDSProxy_Returns502WhenDataPlaneUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", "/nonexistent/unreachable.sock")

	r := gin.New()
	r.Use(UDSProxy())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/completions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["details"], "dial")
}

func TestUDSProxy_ForwardsProviderHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/completions", nil)
	req.Header.Set("X-OmniRoute-Provider", "openai")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUDSProxy_ConnectionCleanedAfterRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/chat/completions", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestResolveSocket_RespectsEnv(t *testing.T) {
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", "/custom/path.sock")
	assert.Equal(t, "/custom/path.sock", resolveSocket())
}

func TestResolveSocket_FallsBackToXDG(t *testing.T) {
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", "")
	t.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")
	assert.Equal(t, "/run/user/1000/omniroute/routed.sock", resolveSocket())
}

func TestResolveSocket_FallsBackToTmp(t *testing.T) {
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", "")
	t.Setenv("XDG_RUNTIME_DIR", "")
	assert.Equal(t, "/tmp/omniroute/routed.sock", resolveSocket())
}

func TestUDSProxy_NonChatPathReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/unknown", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestResolveSocket_RespectsEnv_OverrideDefault(t *testing.T) {
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", "/custom/path.sock")
	assert.Equal(t, "/custom/path.sock", resolveSocket())
}

func TestResolveSocket_DefaultSocketPathConst(t *testing.T) {
	assert.Equal(t, "/tmp/omniroute/routed.sock", DefaultSocketPath)
}

func TestUDSProxy_EmptyBodyReturnsBadGateway(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", "/nonexistent/unreachable.sock")

	r := gin.New()
	r.Use(UDSProxy())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/completions", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestUDSProxy_ReusesTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/chat/completions", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestUDSProxy_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/chat/completions", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestUDSProxy_HealthzPathOverridesDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/chat/completions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUDSProxy_ForwardsProviderHeaderFromClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	socketPath, cleanup := startMockDataPlane(t)
	defer cleanup()
	t.Setenv("OMNIROUTE_DATA_PLANE_SOCKET", socketPath)

	r := gin.New()
	r.Use(UDSProxy())

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/chat/completions", nil)
	req.Header.Set("X-OmniRoute-Provider", "anthropic")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
