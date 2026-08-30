package routes

import (
	"byteport/lib"
	"byteport/models"
	bproutes "byteport/routes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupRouter creates a minimal Gin engine for testing.
func setupRouter() *gin.Engine {
	return gin.New()
}

// createTestToken generates a PASETO v4 local token with the given user-id
// claim, encrypted with the provided symmetric key. This mirrors the logic
// in lib.GenerateToken without requiring the OS keyring.
func createTestToken(key paseto.V4SymmetricKey, userID string) string {
	token := paseto.NewToken()
	token.SetAudience("test@example.com")
	token.SetExpiration(time.Now().Add(time.Hour))
	token.SetSubject("session")
	token.SetIssuer("BytePort")
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetString("user-id", userID)
	return token.V4Encrypt(key, nil)
}

// --- Route handler registration tests ---

// TestGetInstancesRequiresUser verifies that GetInstances returns 401 when
// no user is present in the gin context.
func TestGetInstancesRequiresUser(t *testing.T) {
	router := setupRouter()
	router.GET("/instances", bproutes.GetInstances)

	req := httptest.NewRequest(http.MethodGet, "/instances", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(w.Body.String(), "Unauthorized") {
		t.Errorf("body = %q, want to contain 'Unauthorized'", w.Body.String())
	}
}

// TestGetProjectsRequiresUser verifies that GetProjects returns 401 when
// no user is present in the gin context.
func TestGetProjectsRequiresUser(t *testing.T) {
	router := setupRouter()
	router.GET("/projects", bproutes.GetProjects)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(w.Body.String(), "Unauthorized") {
		t.Errorf("body = %q, want to contain 'Unauthorized'", w.Body.String())
	}
}

// TestGetInstancesInvalidUserContext verifies that GetInstances returns 500
// when the user context value is not of the expected type.
func TestGetInstancesInvalidUserContext(t *testing.T) {
	router := setupRouter()
	// Middleware that sets a wrong type in the "user" context key
	router.Use(func(c *gin.Context) {
		c.Set("user", "not-a-user-struct")
		c.Next()
	})
	router.GET("/instances", bproutes.GetInstances)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/instances", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(w.Body.String(), "Invalid user context") {
		t.Errorf("body = %q, want to contain 'Invalid user context'", w.Body.String())
	}
}

// TestGetProjectsInvalidUserContext verifies that GetProjects returns 500
// when the user context value is not of the expected type.
func TestGetProjectsInvalidUserContext(t *testing.T) {
	router := setupRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user", 12345) // wrong type
		c.Next()
	})
	router.GET("/projects", bproutes.GetProjects)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(w.Body.String(), "Invalid user context") {
		t.Errorf("body = %q, want to contain 'Invalid user context'", w.Body.String())
	}
}

// --- Auth middleware tests ---

// TestAuthMiddlewareBlocksMissingCookie verifies that the auth middleware
// returns 401 when no auth cookie is provided.
func TestAuthMiddlewareBlocksMissingCookie(t *testing.T) {
	router := setupRouter()
	router.Use(lib.AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestAuthMiddlewareBlocksEmptyCookie verifies that the auth middleware
// returns 401 when the auth cookie is empty.
func TestAuthMiddlewareBlocksEmptyCookie(t *testing.T) {
	router := setupRouter()
	router.Use(lib.AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "authToken", Value: ""})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestAuthMiddlewareBlocksInvalidToken verifies that the auth middleware
// returns 401 when the auth cookie contains a garbage token.
func TestAuthMiddlewareBlocksInvalidToken(t *testing.T) {
	router := setupRouter()
	router.Use(lib.AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "authToken", Value: "totally-invalid-token"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(w.Body.String(), "Invalid or expired token") {
		t.Errorf("body = %q, want to contain 'Invalid or expired token'", w.Body.String())
	}
}

// TestAuthMiddlewareWithMockValidToken verifies that when a valid PASETO
// token is set as the cookie, the middleware attempts decryption. When
// the keyring is unavailable (test env), it returns 401 from the token
// validation path rather than the "missing header" path.
func TestAuthMiddlewareWithMockValidToken(t *testing.T) {
	router := setupRouter()
	router.Use(lib.AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Set a PASETO-formatted token string as cookie — will fail at
	// keyring/token validation, but proves the middleware reached
	// the validation step (not the "missing header" short-circuit).
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "authToken",
		Value: "v4.local." + base64.StdEncoding.EncodeToString([]byte("fake-paseto-token")),
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	body := w.Body.String()
	if strings.Contains(body, "Authorization header missing") {
		t.Error("should not return 'Authorization header missing' when cookie is present")
	}
}

// --- Error response status code tests ---

// TestLoginReturnsBadRequestForInvalidJSON verifies that Login returns 400
// when the request body is not valid JSON.
func TestLoginReturnsBadRequestForInvalidJSON(t *testing.T) {
	router := setupRouter()
	router.POST("/login", bproutes.Login)

	body := strings.NewReader("not-json{{{")
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestLoginReturnsBadRequestForEmptyBody verifies that Login returns 400
// when the request body is empty.
func TestLoginReturnsBadRequestForEmptyBody(t *testing.T) {
	router := setupRouter()
	router.POST("/login", bproutes.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestSignupReturnsBadRequestForInvalidJSON verifies that Signup returns 400
// when the request body is not valid JSON.
func TestSignupReturnsBadRequestForInvalidJSON(t *testing.T) {
	router := setupRouter()
	router.POST("/signup", bproutes.Signup)

	body := strings.NewReader("{invalid")
	req := httptest.NewRequest(http.MethodPost, "/signup", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestSignupReturnsBadRequestForMalformedJSON verifies that Signup returns
// 400 when the JSON body is a valid JSON value but not an object.
func TestSignupReturnsBadRequestForMalformedJSON(t *testing.T) {
	router := setupRouter()
	router.POST("/signup", bproutes.Signup)

	// JSON array — valid JSON but wrong shape for struct binding
	body := strings.NewReader(`["name","email"]`)
	req := httptest.NewRequest(http.MethodPost, "/signup", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestAuthenticateReturnsUnauthorizedWithoutCookie verifies that the
// Authenticate handler returns 401 when no auth cookie is provided.
func TestAuthenticateReturnsUnauthorizedWithoutCookie(t *testing.T) {
	router := setupRouter()
	router.GET("/authenticate", bproutes.Authenticate)

	req := httptest.NewRequest(http.MethodGet, "/authenticate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestRouteRegistration verifies that all expected routes are registered
// on a router configured with the same structure as setupRouter in main.go.
func TestRouteRegistration(t *testing.T) {
	router := setupRouter()

	// Register routes matching main.go's protected group
	protected := router.Group("/")
	protected.Use(lib.AuthMiddleware())
	{
		protected.GET("/link", bproutes.LinkHandler)
		protected.GET("/authenticate", bproutes.Authenticate)
		protected.GET("/instances", bproutes.GetInstances)
		protected.GET("/projects", bproutes.GetProjects)
		protected.POST("/deploy", bproutes.DeployProject)
		protected.POST("/terminate", bproutes.TerminateInstance)
	}
	router.POST("/login", bproutes.Login)
	router.POST("/signup", bproutes.Signup)

	// Collect registered routes
	routes := router.Routes()
	routeMap := make(map[string]bool)
	for _, r := range routes {
		key := r.Method + " " + r.Path
		routeMap[key] = true
	}

	expected := []string{
		"GET /link",
		"GET /authenticate",
		"GET /instances",
		"GET /projects",
		"POST /deploy",
		"POST /terminate",
		"POST /login",
		"POST /signup",
	}

	for _, route := range expected {
		if !routeMap[route] {
			t.Errorf("expected route %q not found in registered routes", route)
		}
	}

	// Verify total count (at least the expected routes)
	if len(routes) < len(expected) {
		t.Errorf("registered %d routes, want at least %d", len(routes), len(expected))
	}
}

// TestRouteRegistrationMethodMismatch verifies that a POST request to a
// GET-only route returns 405 Method Not Allowed when HandleMethodNotAllowed
// is enabled on the router.
func TestRouteRegistrationMethodMismatch(t *testing.T) {
	router := setupRouter()
	router.HandleMethodNotAllowed = true
	router.GET("/instances", bproutes.GetInstances)

	req := httptest.NewRequest(http.MethodPost, "/instances", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d (Method Not Allowed)", w.Code, http.StatusMethodNotAllowed)
	}
}

// TestAuthMiddlewareResponseFormat verifies that the auth middleware returns
// a JSON response with the correct structure.
func TestAuthMiddlewareResponseFormat(t *testing.T) {
	router := setupRouter()
	router.Use(lib.AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// Verify response contains JSON with "error" field
	body := w.Body.String()
	if !strings.Contains(body, "error") {
		t.Errorf("response body %q should contain 'error' key", body)
	}

	// Verify Content-Type is JSON
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

// TestGetInstancesReturnsJSON verifies that the handler returns valid JSON
// with the correct Content-Type header even on error paths.
func TestGetInstancesReturnsJSON(t *testing.T) {
	router := setupRouter()
	router.GET("/instances", bproutes.GetInstances)

	req := httptest.NewRequest(http.MethodGet, "/instances", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, "error") {
		t.Errorf("response body %q should contain 'error' key", body)
	}
}

// TestUserContextPropagatedToHandler verifies that when the auth middleware
// sets a user in the context, the handler can retrieve it.
func TestUserContextPropagatedToHandler(t *testing.T) {
	router := setupRouter()

	// Simulate successful auth by setting user in context
	router.Use(func(c *gin.Context) {
		c.Set("user", models.User{
			UUID:  "test-uuid-123",
			Name:  "Test User",
			Email: "test@example.com",
		})
		c.Next()
	})

	var capturedUser models.User
	router.GET("/instances", func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			t.Error("user not in context")
			c.Status(http.StatusInternalServerError)
			return
		}
		user, ok := userVal.(models.User)
		if !ok {
			t.Error("user is not models.User type")
			c.Status(http.StatusInternalServerError)
			return
		}
		capturedUser = user
		c.JSON(http.StatusOK, gin.H{"uuid": user.UUID})
	})

	req := httptest.NewRequest(http.MethodGet, "/instances", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if capturedUser.UUID != "test-uuid-123" {
		t.Errorf("captured user UUID = %q, want %q", capturedUser.UUID, "test-uuid-123")
	}
	if capturedUser.Email != "test@example.com" {
		t.Errorf("captured user email = %q, want %q", capturedUser.Email, "test@example.com")
	}
}
