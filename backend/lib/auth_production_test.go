package lib

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProductionAuthMiddlewareRequiresWorkOSConfiguration(t *testing.T) {
	for _, key := range []string{"WORKOS_CLIENT_ID", "WORKOS_CLIENT_SECRET", "WORKOS_API_KEY"} {
		t.Setenv(key, "")
	}

	if _, err := ProductionAuthMiddleware(context.Background()); err == nil {
		t.Fatal("expected production auth initialization to fail without WorkOS configuration")
	}
}

func TestProductionAuthMiddlewareRejectsSyntheticToken(t *testing.T) {
	t.Setenv("WORKOS_CLIENT_ID", "client-id")
	t.Setenv("WORKOS_CLIENT_SECRET", "client-secret")
	t.Setenv("WORKOS_API_KEY", "api-key")

	middleware, err := ProductionAuthMiddleware(context.Background())
	if err != nil {
		t.Fatalf("initialize production auth: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer test-user")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected synthetic token to be rejected, got %d", resp.Code)
	}
}
