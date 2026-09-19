package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"byteport/models"

	"github.com/gin-gonic/gin"
)

// authTestRouter returns a gin engine with a /login, /signup, /authenticate,
// and /healthz route. We deliberately keep setup minimal: the auth helpers
// need a gin context with optional request bodies but they all branch on
// validation before any database call.
func authTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/signup", Signup)
	r.POST("/login", Login)
	r.GET("/authenticate", Authenticate)
	r.GET("/current", func(c *gin.Context) {
		if _, ok := currentUser(c); !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

// TestCurrentUserMissing verifies that currentUser returns false and writes a
// 401 JSON body when no "user" key is in the gin context.
func TestCurrentUserMissing(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/current", nil)

	user, ok := currentUser(c)
	if ok {
		t.Fatalf("ok = true, want false")
	}
	if user.UUID != "" { // zero-value user returned
		t.Errorf("user = %+v, want zero value", user)
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Unauthorized") {
		t.Errorf("body = %q, want to contain Unauthorized", w.Body.String())
	}
}

// TestCurrentUserWrongType verifies that currentUser returns false and writes
// a 500 JSON body when the "user" value is not a models.User.
func TestCurrentUserWrongType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/current", nil)
	c.Set("user", "not-a-user-struct")

	if _, ok := currentUser(c); ok {
		t.Fatal("ok = true, want false on wrong type")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Invalid user context") {
		t.Errorf("body = %q, want to contain 'Invalid user context'", w.Body.String())
	}
}

// TestCurrentUserHappyPath verifies the success branch: when the context
// carries a models.User, currentUser returns true and writes nothing.
func TestCurrentUserHappyPath(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/current", nil)
	want := models.User{UUID: "u-1", Email: "u@example.com"}
	c.Set("user", want)

	got, ok := currentUser(c)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if got.UUID != want.UUID || got.Email != want.Email {
		t.Errorf("user = %+v, want %+v", got, want)
	}
	if w.Body.Len() != 0 {
		t.Errorf("body = %q, want empty (no error response written)", w.Body.String())
	}
}

// TestSetAuthCookieWritesCookie verifies that setAuthCookie sets the authToken
// cookie with the requested value, the documented SameSite=Lax mode, and the
// secure/httpOnly flags.
func TestSetAuthCookieWritesCookie(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Host = "example.test"

	setAuthCookie(c, "token-abc-123")

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}
	got := cookies[0]
	if got.Name != "authToken" {
		t.Errorf("name = %q, want authToken", got.Name)
	}
	if got.Value != "token-abc-123" {
		t.Errorf("value = %q, want token-abc-123", got.Value)
	}
	if got.Path != "/" {
		t.Errorf("path = %q, want /", got.Path)
	}
	if !got.Secure {
		t.Error("secure = false, want true")
	}
	if !got.HttpOnly {
		t.Error("httpOnly = false, want true")
	}
	if got.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %d, want %d (Lax)", got.SameSite, http.SameSiteLaxMode)
	}
}

// TestAuthenticateMissingCookie verifies the cookie-missing branch of the
// Authenticate handler (it returns 401 before any DB call).
func TestAuthenticateMissingCookie(t *testing.T) {
	r := authTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/authenticate", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Unauthorized") {
		t.Errorf("body = %q, want to contain Unauthorized", w.Body.String())
	}
}

// TestSignupRejectsNonJSONBody verifies the body-binding error branch.
func TestSignupRejectsNonJSONBody(t *testing.T) {
	r := authTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestLoginRejectsNonJSONBody verifies the body-binding error branch.
func TestLoginRejectsNonJSONBody(t *testing.T) {
	r := authTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
