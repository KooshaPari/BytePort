package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/byteport/api/internal/infrastructure/auth"
	"github.com/byteport/api/internal/infrastructure/secrets"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noAuthHeader is the sentinel value passed to runMiddleware* helpers to
// indicate that no Authorization header should be set on the request.
const noAuthHeader = ""

// runMiddlewareViaRouter executes mw inside a fresh gin engine, with the
// given Authorization header (or no header when hdr == noAuthHeader), and
// dispatches a single GET /test through handler. Returns the response
// recorder so callers can assert on w.Code and w.Body.
func runMiddlewareViaRouter(t *testing.T, mw, handler gin.HandlerFunc, hdr string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)
	router.Use(mw)
	router.GET("/test", handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if hdr != noAuthHeader {
		req.Header.Set("Authorization", hdr)
	}
	router.ServeHTTP(w, req)
	return w
}

// runMiddlewareDirect executes mw against a fresh gin.Context, with the
// given Authorization header (or none). Used for middlewares that abort
// before reaching a downstream handler.
func runMiddlewareDirect(t *testing.T, mw gin.HandlerFunc, hdr string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if hdr != noAuthHeader {
		req.Header.Set("Authorization", hdr)
	}
	c.Request = req
	mw(c)
	return w
}

// assertJSONError unmarshals w.Body and asserts the JSON "error" and "code"
// fields. Used to confirm middleware-emitted 401 payloads.
func assertJSONError(t *testing.T, w *httptest.ResponseRecorder, wantError, wantCode string) {
	t.Helper()
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, wantError, body["error"])
	assert.Equal(t, wantCode, body["code"])
}

// newTestAuthService returns an initialized WorkOSAuthService backed by the
// mock secrets manager below.
func newTestAuthService(t *testing.T) *auth.WorkOSAuthService {
	t.Helper()
	manager := newMockSecretsManager()
	svc := auth.NewWorkOSAuthService(manager)
	require.NoError(t, svc.Initialize(context.Background()))
	return svc
}

// newMockSecretsManager returns a secrets.Manager with the WorkOS client
// credentials registered under a single "mock" provider.
func newMockSecretsManager() *secrets.Manager {
	manager := secrets.New(secrets.Config{CacheTTL: time.Minute})
	mock := &mockProvider{
		secrets: map[string]string{
			secrets.SecretWorkOSClientID:     "test-client-id",
			secrets.SecretWorkOSClientSecret: "test-client-secret",
			secrets.SecretWorkOSAPIKey:       "test-api-key",
		},
	}
	manager.RegisterProvider("mock", mock)
	return manager
}

// runOptionalAuthSubtest executes an OptionalAuth-style middleware in a
// fresh gin engine and asserts on the response status and any user_uuid
// context value the handler observes. Used by both the Legacy and the
// modern Optional auth test loops to avoid copy-paste of the capture-
// handler-assertions block.
func runOptionalAuthSubtest(t *testing.T, mw gin.HandlerFunc, authHeader string, wantStatus int, wantExists bool, wantUUID string) {
	t.Helper()
	var capturedUUID string
	var capturedExists bool
	handler := func(c *gin.Context) {
		if uuid, ok := c.Get("user_uuid"); ok {
			capturedExists = true
			capturedUUID = uuid.(string)
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
	w := runMiddlewareViaRouter(t, mw, handler, authHeader)
	assert.Equal(t, wantStatus, w.Code)
	if wantExists {
		assert.True(t, capturedExists)
		assert.Equal(t, wantUUID, capturedUUID)
	}
}

// mockProvider implements the secrets.Provider interface for testing.
type mockProvider struct {
	secrets map[string]string
}

func (m *mockProvider) GetSecret(_ context.Context, key string) (string, error) {
	if secret, exists := m.secrets[key]; exists {
		return secret, nil
	}
	return "", fmt.Errorf("secret not found: %s", key)
}

func (m *mockProvider) SetSecret(_ context.Context, key, value string) error {
	if m.secrets == nil {
		m.secrets = make(map[string]string)
	}
	m.secrets[key] = value
	return nil
}

func (m *mockProvider) DeleteSecret(_ context.Context, key string) error {
	delete(m.secrets, key)
	return nil
}

func (m *mockProvider) ListSecrets(_ context.Context) ([]string, error) {
	keys := make([]string, 0, len(m.secrets))
	for key := range m.secrets {
		keys = append(keys, key)
	}
	return keys, nil
}
