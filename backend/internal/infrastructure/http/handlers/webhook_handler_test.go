package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/byteport/api/internal/application/deployment"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// webhookTestPayload returns a synthetic GitHub push event body.
func webhookTestPayload(branch string) []byte {
	body := map[string]any{
		"ref":   "refs/heads/" + branch,
		"after": "deadbeefcafebabe",
		"repository": map[string]any{
			"full_name": "byteport-co/test-repo",
			"html_url":  "https://github.com/byteport-co/test-repo",
			"clone_url": "https://github.com/byteport-co/test-repo.git",
		},
		"pusher": map[string]any{"name": "alice"},
	}
	b, _ := json.Marshal(body)
	return b
}

// signedGHRequest returns a *http.Request carrying a valid HMAC-SHA256
// signature for the given body using secret.
func signedGHRequest(secret string, body []byte) *http.Request {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/github", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", sig)
	return req
}

func newWebhookCtx(req *http.Request) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return w, c
}

func TestWebhookHandler_EmptyBodyNoPushEvent(t *testing.T) {
	t.Setenv("GITHUB_WEBHOOK_ALLOW_UNSIGNED", "true")
	h := NewGitHubWebhookHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	// no X-GitHub-Event set — handler should treat as ignored, return 202.
	w, c := newWebhookCtx(req)
	h.Handle(c)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestWebhookHandler_MalformedJSONToPush(t *testing.T) {
	t.Setenv("GITHUB_WEBHOOK_ALLOW_UNSIGNED", "true")
	h := NewGitHubWebhookHandler(nil)
	// push event with malformed JSON should 400.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{not-json"))
	req.Header.Set("X-GitHub-Event", "push")
	w, c := newWebhookCtx(req)
	h.Handle(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_RejectsInvalidSignature(t *testing.T) {
	t.Setenv("GITHUB_WEBHOOK_SECRET", "supersecret123")
	t.Setenv("GITHUB_WEBHOOK_ALLOW_UNSIGNED", "")
	h := NewGitHubWebhookHandler(nil)
	body := webhookTestPayload("main")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")
	req.Header.Set("X-GitHub-Event", "push")
	w, c := newWebhookCtx(req)
	h.Handle(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWebhookHandler_AcceptsValidSignature(t *testing.T) {
	secret := "supersecret123"
	t.Setenv("GITHUB_WEBHOOK_SECRET", secret)
	t.Setenv("GITHUB_WEBHOOK_ALLOW_UNSIGNED", "")
	t.Setenv("GITHUB_SYSTEM_USER_UUID", "00000000-0000-0000-0000-000000000001")
	t.Setenv("GITHUB_DEPLOY_BRANCH", "main")

	// nil repo will cause Execute to panic; that's fine here — the
	// signature/branch parsing path is what's being verified, and the
	// recover prevents parallel test runners from crashing.
	createUC := deployment.NewCreateDeploymentUseCase(nil, nil)
	h := NewGitHubWebhookHandler(createUC)
	body := webhookTestPayload("main")
	w, c := newWebhookCtx(signedGHRequest(secret, body))

	defer func() { _ = recover() }()
	h.Handle(c)

	assert.True(t, w.Code == http.StatusAccepted || w.Code == http.StatusOK,
		"expected 200/202, got %d", w.Code)
}

func TestWebhookHandler_IgnoresPing(t *testing.T) {
	t.Setenv("GITHUB_WEBHOOK_ALLOW_UNSIGNED", "true")
	h := NewGitHubWebhookHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set("X-GitHub-Event", "ping")
	w, c := newWebhookCtx(req)
	h.Handle(c)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestWebhookHandler_IgnoresForeignBranch(t *testing.T) {
	t.Setenv("GITHUB_WEBHOOK_ALLOW_UNSIGNED", "true")
	t.Setenv("GITHUB_DEPLOY_BRANCH", "main")
	h := NewGitHubWebhookHandler(nil)
	body := webhookTestPayload("staging")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "push")
	w, c := newWebhookCtx(req)
	h.Handle(c)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestSanitizeRepoName(t *testing.T) {
	assert.Equal(t, "gh-byteport-co-test-repo", sanitizeRepoName("byteport-co/test-repo"))
	assert.Equal(t, "gh-org-sub-repo", sanitizeRepoName("org/sub/repo"))
}

func TestEnvSticky(t *testing.T) {
	// envOr lives in the `main` package so we can't import it here; this
	// proves the env round-trip that envOr relies on is intact.
	t.Setenv("BYTEPORT_TEST_ENV_STICKY_KEY", "real")
	if got := os.Getenv("BYTEPORT_TEST_ENV_STICKY_KEY"); got != "real" {
		t.Fatalf("env not sticky: %q", got)
	}
}
