package middleware

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := newTestAuthService(t)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "creates middleware using WorkOS auth service",
			authHeader: "Bearer test-user123-test_at_example.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "handles invalid token",
			authHeader: "Bearer invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "handles missing authorization header",
			authHeader: noAuthHeader,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			middleware := AuthMiddleware(authService)
			w := runMiddlewareDirect(t, middleware, tc.authHeader)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestLegacyOptionalAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		wantStatus     int
		wantUserUUID   string
		wantUserExists bool
	}{
		{
			name:       "allows requests without authorization header",
			authHeader: noAuthHeader,
			wantStatus: http.StatusOK,
		},
		{
			name:           "handles valid authorization header",
			authHeader:     "Bearer any-token",
			wantStatus:     http.StatusOK,
			wantUserUUID:   "placeholder-user-uuid",
			wantUserExists: true,
		},
		{
			name:       "handles malformed authorization header gracefully",
			authHeader: "InvalidFormat token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "handles authorization header with only Bearer",
			authHeader: "Bearer",
			wantStatus: http.StatusOK,
		},
		{
			name:       "handles empty authorization header",
			authHeader: "",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runOptionalAuthSubtest(t, LegacyOptionalAuthMiddleware(), tc.authHeader, tc.wantStatus, tc.wantUserExists, tc.wantUserUUID)
		})
	}
}

func TestAuthMiddlewareWithFallback_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles nil auth service", func(t *testing.T) {
		handler := func(c *gin.Context) {
			if uuid, ok := c.Get("user_uuid"); ok {
				assert.Equal(t, "user", uuid.(string))
			} else {
				t.Errorf("expected user_uuid context key, got none")
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
		w := runMiddlewareViaRouter(t, AuthMiddlewareWithFallback(nil), handler, "Bearer test-user-test_at_example.com")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	tests := []struct {
		name       string
		authHeader string
		wantError  string
	}{
		{
			name:       "handles missing authorization header with nil auth service",
			authHeader: noAuthHeader,
			wantError:  "missing authorization header",
		},
		{
			name:       "handles invalid authorization header format with nil auth service",
			authHeader: "InvalidFormat token",
			wantError:  "invalid authorization header format",
		},
		{
			name:       "handles invalid token with nil auth service",
			authHeader: "Bearer invalid-token",
			wantError:  "invalid or expired token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := runMiddlewareDirect(t, AuthMiddlewareWithFallback(nil), tc.authHeader)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assertJSONError(t, w, tc.wantError, "UNAUTHORIZED")
		})
	}
}

func TestOptionalAuthMiddleware_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		authHeader   string
		wantStatus   int
		wantUserUUID string
		wantExists   bool
	}{
		{
			name:         "handles nil auth service",
			authHeader:   "Bearer test-user-test_at_example.com",
			wantStatus:   http.StatusOK,
			wantUserUUID: "user",
			wantExists:   true,
		},
		{
			name:       "handles malformed authorization header with nil auth service",
			authHeader: "InvalidFormat token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "handles authorization header with only Bearer with nil auth service",
			authHeader: "Bearer",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runOptionalAuthSubtest(t, OptionalAuthMiddleware(nil), tc.authHeader, tc.wantStatus, tc.wantExists, tc.wantUserUUID)
		})
	}
}
