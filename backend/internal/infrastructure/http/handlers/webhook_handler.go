package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/byteport/api/internal/application/deployment"
	"github.com/gin-gonic/gin"
)

// GitHubWebhookHandler delivers a PoC CI/CD pipeline that triggers a deployment
// whenever GitHub posts a `push` event for a configured branch.
//
// Verification uses HMAC-SHA256 over the raw request body with a shared secret
// stored in $GITHUB_WEBHOOK_SECRET — the same contract GitHub uses. Setting
// $GITHUB_WEBHOOK_ALLOW_UNSIGNED=true disables signing for local dev only.
//
// Concurrency is bounded by a small repository→branch goroutine map; if you
// ship this, swap in a real queue (the engine adapter in byteport-engine has
// the dispatch primitives for it).
type GitHubWebhookHandler struct {
	createUseCase *deployment.CreateDeploymentUseCase
	secret        []byte
	allowUnsigned bool

	inFlight   map[string]bool
	inFlightMu sync.Mutex
}

// NewGitHubWebhookHandler constructs the webhook handler.
func NewGitHubWebhookHandler(
	createUseCase *deployment.CreateDeploymentUseCase,
) *GitHubWebhookHandler {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	allowUnsigned := strings.EqualFold(os.Getenv("GITHUB_WEBHOOK_ALLOW_UNSIGNED"), "true")
	if secret == "" && !allowUnsigned {
		// We log via root logger later — keep silent in tests when unset.
		secret = ""
	}
	return &GitHubWebhookHandler{
		createUseCase: createUseCase,
		secret:        []byte(secret),
		allowUnsigned: allowUnsigned,
		inFlight:      make(map[string]bool),
	}
}

// RegisterRoutes mounts the webhook handler at POST /webhooks/github. This is
// intentionally NOT behind the AuthMiddleware — GitHub signs the body itself.
func (h *GitHubWebhookHandler) RegisterRoutes(router *gin.RouterGroup) {
	wh := router.Group("/webhooks")
	{
		wh.POST("/github", h.Handle)
	}
}

// GitHubPushPayload is the subset of GitHub's `push` event we care about.
type GitHubPushPayload struct {
	Ref        string `json:"ref"`   // refs/heads/main
	After      string `json:"after"` // commit SHA
	Repository struct {
		FullName string `json:"full_name"` // owner/repo
		HTMLURL  string `json:"html_url"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
	Pusher struct {
		Name string `json:"name"`
	} `json:"pusher"`
	HeadCommit struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	} `json:"head_commit"`
}

// Handle is the entry point GitHub POSTs to.
func (h *GitHubWebhookHandler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "could not read body",
			Code:  "READ_BODY_FAILED",
		})
		return
	}

	if !h.allowUnsigned {
		sigHeader := c.GetHeader("X-Hub-Signature-256")
		if !h.verifySignature(sigHeader, body) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "invalid signature",
				Code:  "INVALID_SIGNATURE",
			})
			return
		}
	}

	event := c.GetHeader("X-GitHub-Event")
	if event != "push" {
		// GitHub sends pings; just 202.
		c.JSON(http.StatusAccepted, gin.H{"ignored": event})
		return
	}

	var payload GitHubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid json",
			Code:  "INVALID_JSON",
		})
		return
	}

	if !strings.HasPrefix(payload.Ref, "refs/heads/") {
		c.JSON(http.StatusAccepted, gin.H{"ignored": "non-branch ref", "ref": payload.Ref})
		return
	}
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Only deploy the configured branch (default: main).
	targetBranch := os.Getenv("GITHUB_DEPLOY_BRANCH")
	if targetBranch == "" {
		targetBranch = "main"
	}
	if branch != targetBranch {
		c.JSON(http.StatusAccepted, gin.H{"ignored": "branch", "branch": branch})
		return
	}

	deployKey := payload.Repository.FullName + "@" + branch
	if !h.tryAcquire(deployKey) {
		c.JSON(http.StatusAccepted, gin.H{"ignored": "in-flight", "key": deployKey})
		return
	}
	defer h.release(deployKey)

	// Build a deployment request from the webhook payload.
	name := sanitizeRepoName(payload.Repository.FullName)
	if name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "could not derive deployment name from repository",
			Code:  "INVALID_REPO",
		})
		return
	}

	userUUID, err := h.resolveSystemUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: err.Error(),
			Code:  "WEBHOOK_AUTH_FAILED",
		})
		return
	}

	gitURL := payload.Repository.CloneURL
	if gitURL == "" {
		gitURL = payload.Repository.HTMLURL
	}

	req := deployment.CreateDeploymentRequest{
		Name:  name,
		Owner: userUUID,
		Config: map[string]interface{}{
			"git_url":      gitURL,
			"branch":       branch,
			"commit_sha":   payload.After,
			"trigger":      "github-webhook",
			"pusher":       payload.Pusher.Name,
			"webhook_time": time.Now().UTC().Format(time.RFC3339),
		},
	}

	resp, err := h.createUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		handleApplicationError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// verifySignature checks the X-Hub-Signature-256 hex digest against HMAC-SHA256
// of body. Empty secrets reject all signed requests.
func (h *GitHubWebhookHandler) verifySignature(header string, body []byte) bool {
	if len(h.secret) == 0 {
		return false
	}
	if !strings.HasPrefix(header, "sha256=") {
		return false
	}
	provided, err := hex.DecodeString(strings.TrimPrefix(header, "sha256="))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, h.secret)
	mac.Write(body)
	expected := mac.Sum(nil)
	return hmac.Equal(provided, expected)
}

// tryAcquire prevents the same (repo, branch) from triggering twice concurrently.
func (h *GitHubWebhookHandler) tryAcquire(key string) bool {
	h.inFlightMu.Lock()
	defer h.inFlightMu.Unlock()
	if h.inFlight[key] {
		return false
	}
	h.inFlight[key] = true
	return true
}

func (h *GitHubWebhookHandler) release(key string) {
	h.inFlightMu.Lock()
	defer h.inFlightMu.Unlock()
	delete(h.inFlight, key)
}

// resolveSystemUser maps a webhook delivery to a system user. In production
// lookup the github-login → user-uuid mapping; in dev accept the system user.
func (h *GitHubWebhookHandler) resolveSystemUser(c *gin.Context) (string, error) {
	if v := os.Getenv("GITHUB_SYSTEM_USER_UUID"); v != "" {
		return v, nil
	}
	if u := getUserUUID(c); u != "" {
		return u, nil
	}
	return "", errors.New("no system user mapping configured for webhooks")
}

// sanitizeRepoName maps `owner/repo` → `owner-repo` to fit deployment names.
func sanitizeRepoName(fullName string) string {
	out := make([]byte, 0, len(fullName))
	for i := 0; i < len(fullName); i++ {
		c := fullName[i]
		switch {
		case c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == '-' || c == '_':
			out = append(out, c)
		default:
			out = append(out, '-')
		}
	}
	if len(out) > 63 {
		out = out[:63]
	}
	if len(out) == 0 {
		return ""
	}
	return fmt.Sprintf("gh-%s", string(out))
}
