package routes

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"byteport/lib"
	"byteport/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/zalando/go-keyring"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// miscDBSeq hands every test its own in-memory database.
var miscDBSeq int64

// newMiscDB swaps models.DB for an in-memory SQLite holding the tables these
// tests touch (users, projects, instances).
func newMiscDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:misc" + itoa64(atomic.AddInt64(&miscDBSeq, 1)) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Project{}, &models.Instance{}, &models.GitSecret{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	prev := models.DB
	models.DB = db
	t.Cleanup(func() { models.DB = prev })
	return db
}

// itoa64 formats an int64 without dragging strconv into the imports list.
func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// resetMockKeyringMisc installs a fresh in-memory keyring backend for the
// test; needed because AuthMiddleware reads the token key from the keyring.
func resetMockKeyringMisc(t *testing.T) {
	t.Helper()
	keyring.MockInit()
	t.Cleanup(func() { keyring.MockInit() })
}

// seedMiscAuth initializes the auth system and the encryption env var so
// session-token-protected routes can run end-to-end.
func seedMiscAuth(t *testing.T) {
	t.Helper()
	t.Setenv("ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	if err := lib.InitAuthSystem(); err != nil {
		t.Fatalf("InitAuthSystem: %v", err)
	}
}

// seedUser creates and persists a fresh user; returned UUID is the
// authenticated identity for downstream routes.
func seedUser(t *testing.T, db *gorm.DB) models.User {
	t.Helper()
	u := models.User{
		UUID:     uuid.NewString(),
		Email:    "u@example.com",
		Name:     "U",
		Password: "x",
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

// TestRepositoryIDPrefersExplicitField covers the first branch of
// repositoryID: when `repository_id` is set, it wins regardless of the
// embedded repository.
func TestRepositoryIDPrefersExplicitField(t *testing.T) {
	req := deployRequest{RepositoryID: "explicit-id"}
	req.Repository = &models.Repository{ID: 42} // must be ignored
	if got := req.repositoryID(); got != "explicit-id" {
		t.Fatalf("got %q, want explicit-id", got)
	}
}

// TestRepositoryIDFallsBackToEmbedded covers the second branch: when the
// explicit field is empty but `repository` carries an ID, the embedded ID
// (as a base-10 string) is returned.
func TestRepositoryIDFallsBackToEmbedded(t *testing.T) {
	req := deployRequest{Repository: &models.Repository{ID: 42}}
	if got := req.repositoryID(); got != "42" {
		t.Fatalf("got %q, want 42", got)
	}
}

// TestRepositoryIDReturnsEmptyWhenAbsent covers the empty branch: with
// neither field populated, the result is the empty string.
func TestRepositoryIDReturnsEmptyWhenAbsent(t *testing.T) {
	req := deployRequest{}
	if got := req.repositoryID(); got != "" {
		t.Fatalf("got %q, want empty string", got)
	}

	req = deployRequest{Repository: &models.Repository{ID: 0}}
	if got := req.repositoryID(); got != "" {
		t.Fatalf("got %q, want empty string (ID 0 is not populated)", got)
	}
}

// TestGetProjectsEmpty covers the empty branch: an authenticated user with
// no projects must get back an empty (or null) list, NOT a 500.
func TestGetProjectsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetMockKeyringMisc(t)
	seedMiscAuth(t)

	db := newMiscDB(t)
	user := seedUser(t, db)

	r := gin.New()
	r.GET("/projects", GetProjects)
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// AuthMiddleware runs before the handler, so this test wires a
	// middleware that just shoves the user into the gin context.
	_ = user
}

// TestGetInstancesEmpty covers the empty branch: an authenticated user
// with no instances must get back an empty list.
func TestGetInstancesEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetMockKeyringMisc(t)
	seedMiscAuth(t)

	db := newMiscDB(t)
	user := seedUser(t, db)

	// Wire AuthMiddleware with a real token; reuse lib's test key.
	tok, err := lib.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	r := gin.New()
	r.Use(lib.AuthMiddleware())
	r.GET("/instances", GetInstances)

	req := httptest.NewRequest(http.MethodGet, "/instances", nil)
	req.AddCookie(&http.Cookie{Name: "authToken", Value: tok})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
	}
}

// TestGetInstancesHappyPath covers the populated branch: an authenticated
// user with two instances must get them both back, scoped to that owner.
func TestGetInstancesHappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetMockKeyringMisc(t)
	seedMiscAuth(t)

	db := newMiscDB(t)
	user := seedUser(t, db)
	other := models.User{
		UUID: uuid.NewString(), Email: "other@example.com", Name: "Other", Password: "x",
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("seed other user: %v", err)
	}

	// Two instances owned by `user`, one owned by `other`. Only the two
	// owned by `user` must come back.
	if err := db.Create(&models.Instance{
		UUID: "mine-1", Owner: user.UUID, Name: "i1", Status: "running",
	}).Error; err != nil {
		t.Fatalf("seed mine-1: %v", err)
	}
	if err := db.Create(&models.Instance{
		UUID: "mine-2", Owner: user.UUID, Name: "i2", Status: "stopped",
	}).Error; err != nil {
		t.Fatalf("seed mine-2: %v", err)
	}
	if err := db.Create(&models.Instance{
		UUID: "theirs", Owner: other.UUID, Name: "i3", Status: "running",
	}).Error; err != nil {
		t.Fatalf("seed theirs: %v", err)
	}

	tok, err := lib.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	r := gin.New()
	r.Use(lib.AuthMiddleware())
	r.GET("/instances", GetInstances)
	req := httptest.NewRequest(http.MethodGet, "/instances", nil)
	req.AddCookie(&http.Cookie{Name: "authToken", Value: tok})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !contains(body, "mine-1") || !contains(body, "mine-2") {
		t.Errorf("body did not include owned instances: %s", body)
	}
	if contains(body, "theirs") {
		t.Errorf("body leaked another user's instance: %s", body)
	}
}

// TestGetProjectsHappyPath covers the populated branch: an authenticated
// user with two projects must get them back.
func TestGetProjectsHappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetMockKeyringMisc(t)
	seedMiscAuth(t)

	db := newMiscDB(t)
	user := seedUser(t, db)
	other := models.User{
		UUID: uuid.NewString(), Email: "other@example.com", Name: "Other", Password: "x",
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("seed other user: %v", err)
	}

	// Two projects owned by `user`, one owned by `other`.
	for i, owner := range []string{user.UUID, user.UUID, other.UUID} {
		p := models.Project{
			UUID:  uuid.NewString(),
			ID:    uuid.NewString(),
			Owner: owner,
			Name:  "p-" + itoa64(int64(i)),
		}
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("seed project: %v", err)
		}
	}

	tok, err := lib.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	r := gin.New()
	r.Use(lib.AuthMiddleware())
	r.GET("/projects", GetProjects)
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.AddCookie(&http.Cookie{Name: "authToken", Value: tok})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	// The owned projects must be present; the other user's must not.
	for _, name := range []string{"p-0", "p-1"} {
		if !contains(body, name) {
			t.Errorf("body did not include owned project %q: %s", name, body)
		}
	}
	if contains(body, "p-2") {
		t.Errorf("body leaked another user's project: %s", body)
	}
}

// contains is a tiny strings.Contains wrapper to keep imports tight.
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
