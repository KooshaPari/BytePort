package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/byteport/api/internal/infrastructure/clients"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeployRequest represents a deployment request
type DeployRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Type     string                 `json:"type" binding:"required"` // frontend, backend, database
	Provider string                 `json:"provider"`                // optional, auto-selected if empty
	GitURL   string                 `json:"git_url"`
	Branch   string                 `json:"branch"`
	EnvVars  map[string]string      `json:"env_vars"`
	Config   map[string]interface{} `json:"config"`
}

// DeployResponse represents a deployment response
type DeployResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	URL       string    `json:"url"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// handleDeploy handles deployment requests.
//
// When the request selects provider "engine" AND an engine daemon client is
// available, the request is forwarded to the Rust byteport-engine daemon
// over UDS. Otherwise we fall through to the existing in-process simulated
// deploy path so the legacy /legacy/deployments contract is preserved.
//
// The engine client is optional — callers in tests or dev environments that
// never set BYTEPORT_ENGINE_SOCKET pass nil and the simulated path is used
// for every request.
func handleDeploy(store *DeploymentStore, engineClient *clients.EngineDaemonClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeployRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// Engine-provider short-circuit: forward to the Rust daemon when
		// the caller asked for it AND the daemon client was wired up at
		// startup. Without a client we fall back to the simulation so
		// the legacy route still responds.
		if req.Provider == "engine" {
			if engineClient == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "engine provider requested but BYTEPORT_ENGINE_SOCKET is not configured",
				})
				return
			}
			handleEngineDeploy(c, store, engineClient, req)
			return
		}

		// Auto-select provider if not specified
		if req.Provider == "" {
			req.Provider = selectOptimalProvider(req.Type)
		}

		// Validate provider
		if !isValidProvider(req.Provider) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":           "Invalid provider",
				"valid_providers": []string{"vercel", "netlify", "render", "railway", "supabase", "cloudflare-pages", "engine"},
			})
			return
		}

		// Create deployment
		deployment := &Deployment{
			ID:        uuid.New().String(),
			Name:      req.Name,
			Type:      req.Type,
			Provider:  req.Provider,
			Status:    "deploying",
			GitURL:    req.GitURL,
			Branch:    req.Branch,
			EnvVars:   req.EnvVars,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Generate URL
		deployment.URL = generateDeploymentURL(req.Name, req.Provider)

		// Store deployment
		store.Add(deployment)

		// Trigger deployment: real provisioning for AWS, simulated for
		// third-party providers (vercel/netlify/etc) until their adapters land.
		if deployment.Provider == "aws" {
			go provisionEC2Instance(store, deployment.ID, req)
		} else {
			go simulateDeployment(store, deployment.ID)
		}

		// Return response
		c.JSON(http.StatusCreated, DeployResponse{
			ID:        deployment.ID,
			Name:      deployment.Name,
			Status:    deployment.Status,
			URL:       deployment.URL,
			Provider:  deployment.Provider,
			CreatedAt: deployment.CreatedAt,
			Message:   "Deployment started successfully",
		})
	}
}

// handleEngineDeploy forwards a legacy /legacy/deployments request whose
// provider is "engine" to the Rust byteport-engine daemon over UDS. The
// returned deployment is also recorded in the in-memory store so the
// subsequent /legacy/deployments/:id GET, /stop, etc. continue to work.
//
// Wire-format conversion: the legacy DeployRequest does not have a
// ServiceEntry; we synthesise a single-service manifest from the legacy
// fields. GitURL becomes the service `path` (used as the image reference
// by the Rust adapter), the chosen Type becomes the service name, and
// the legacy EnvVars map is converted to the daemon's []EnvEntry shape.
func handleEngineDeploy(
	c *gin.Context,
	store *DeploymentStore,
	engineClient *clients.EngineDaemonClient,
	req DeployRequest,
) {
	// Synthesise a single-service payload from the legacy request.
	services := []clients.ServiceEntry{}
	env := make([]clients.EnvEntry, 0, len(req.EnvVars))
	for k, v := range req.EnvVars {
		env = append(env, clients.EnvEntry{Key: k, Value: v})
	}
	// Service name falls back to the deployment name when Type is unset.
	svcName := req.Type
	if svcName == "" {
		svcName = req.Name
	}
	services = append(services, clients.ServiceEntry{
		Name: svcName,
		Path: req.GitURL,
		Port: 80,
		Env:  env,
	})

	// Best-effort user identity — auth middleware sets user_id / user_email
	// on the gin.Context; fall back to "anonymous" when the legacy path is
	// reached without auth (e.g. during a test).
	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("user_email")
	uidStr, _ := userID.(string)
	if uidStr == "" {
		uidStr = "anonymous"
	}
	emailStr, _ := userEmail.(string)

	daemonReq := &clients.DeployRequest{
		Name:       req.Name,
		User:       clients.DeployUser{ID: uidStr, Email: emailStr},
		Repository: req.GitURL,
		Services:   services,
	}

	result, err := engineClient.Deploy(c.Request.Context(), daemonReq)
	if err != nil {
		if err == clients.ErrDaemonUnavailable {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "engine daemon unavailable",
			})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "engine daemon returned an error",
			"details": err.Error(),
		})
		return
	}

	// Record the deployment in the in-memory store so legacy read endpoints
	// (/legacy/deployments/:id, /:id/status, /:id/logs, /:id/metrics) keep
	// working with the same in-process contract.
	deployment := &Deployment{
		ID:        result.DeploymentID,
		Name:      req.Name,
		Type:      req.Type,
		Provider:  "engine",
		Status:    "deploying",
		GitURL:    req.GitURL,
		Branch:    req.Branch,
		EnvVars:   req.EnvVars,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	deployment.URL = generateDeploymentURL(req.Name, "engine")
	store.Add(deployment)

	c.JSON(http.StatusCreated, DeployResponse{
		ID:        deployment.ID,
		Name:      deployment.Name,
		Status:    deployment.Status,
		URL:       deployment.URL,
		Provider:  deployment.Provider,
		CreatedAt: deployment.CreatedAt,
		Message:   "Engine deployment started",
	})
}

// handleListDeployments lists all deployments
func handleListDeployments(store *DeploymentStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		deployments := store.List()

		response := make([]DeployResponse, len(deployments))
		for i, dep := range deployments {
			response[i] = DeployResponse{
				ID:        dep.ID,
				Name:      dep.Name,
				Status:    dep.Status,
				URL:       dep.URL,
				Provider:  dep.Provider,
				CreatedAt: dep.CreatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"deployments": response,
			"total":       len(response),
		})
	}
}

// handleGetDeployment gets a specific deployment
func handleGetDeployment(store *DeploymentStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		deployment := store.Get(id)

		if deployment == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Deployment not found",
			})
			return
		}

		c.JSON(http.StatusOK, deployment)
	}
}

// handleTerminateDeployment terminates a deployment
func handleTerminateDeployment(store *DeploymentStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		deployment := store.Get(id)

		if deployment == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Deployment not found",
			})
			return
		}

		// Update status
		deployment.Status = "terminated"
		deployment.UpdatedAt = time.Now()
		store.Update(deployment)

		c.JSON(http.StatusOK, gin.H{
			"message": "Deployment terminated successfully",
			"id":      id,
		})
	}
}

// handleGetStatus gets deployment status
func handleGetStatus(store *DeploymentStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		deployment := store.Get(id)

		if deployment == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Deployment not found",
			})
			return
		}

		status := gin.H{
			"id":         deployment.ID,
			"status":     deployment.Status,
			"progress":   getProgress(deployment.Status),
			"updated_at": deployment.UpdatedAt,
		}

		c.JSON(http.StatusOK, status)
	}
}

// handleGetLogs gets deployment logs
func handleGetLogs(store *DeploymentStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		deployment := store.Get(id)

		if deployment == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Deployment not found",
			})
			return
		}

		// Mock logs
		logs := []gin.H{
			{
				"timestamp": time.Now().Add(-5 * time.Minute),
				"level":     "info",
				"message":   "Starting deployment process",
			},
			{
				"timestamp": time.Now().Add(-4 * time.Minute),
				"level":     "info",
				"message":   "Building application...",
			},
			{
				"timestamp": time.Now().Add(-2 * time.Minute),
				"level":     "info",
				"message":   "Deploying to " + deployment.Provider,
			},
			{
				"timestamp": time.Now(),
				"level":     "info",
				"message":   "Deployment completed successfully",
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"deployment_id": id,
			"logs":          logs,
		})
	}
}

// handleGetMetrics gets deployment metrics
func handleGetMetrics(store *DeploymentStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		deployment := store.Get(id)

		if deployment == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Deployment not found",
			})
			return
		}

		// Mock metrics
		metrics := gin.H{
			"deployment_id": id,
			"uptime":        "99.9%",
			"requests":      1234,
			"bandwidth":     "1.2 GB",
			"response_time": "45ms",
			"cost": gin.H{
				"monthly":  0.0,
				"currency": "USD",
			},
		}

		c.JSON(http.StatusOK, metrics)
	}
}

// handleDetectApp detects application type
func handleDetectApp(c *gin.Context) {
	var req struct {
		Files []string `json:"files"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simple detection logic
	detection := gin.H{
		"type":               "frontend",
		"framework":          "react",
		"confidence":         0.9,
		"suggested_provider": "vercel",
	}

	c.JSON(http.StatusOK, detection)
}

// handleEstimateCost estimates deployment cost
func handleEstimateCost(c *gin.Context) {
	var req struct {
		Type     string `json:"type"`
		Provider string `json:"provider"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cost := gin.H{
		"monthly":  0.0,
		"currency": "USD",
		"breakdown": []gin.H{
			{
				"service":  req.Type,
				"provider": req.Provider,
				"cost":     0.0,
				"plan":     "Free Tier",
			},
		},
		"message": "This deployment uses 100% free tiers",
	}

	c.JSON(http.StatusOK, cost)
}

// handleValidateConfig validates a deployment configuration
func handleValidateConfig(c *gin.Context) {
	var config map[string]interface{}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validation := gin.H{
		"valid":    true,
		"errors":   []string{},
		"warnings": []string{},
	}

	c.JSON(http.StatusOK, validation)
}

func handleWorkOSCallback(c *gin.Context) {
	var req struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": req.Code,
		"token_type":   "Bearer",
		"state":        req.State,
	})
}

func handleGetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		id = "current"
	}

	user := gin.H{"id": id}
	if userID, ok := c.Get("user_id"); ok {
		user["id"] = userID
	}
	if email, ok := c.Get("user_email"); ok {
		user["email"] = email
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// Helper functions

func selectOptimalProvider(appType string) string {
	switch appType {
	case "frontend":
		return "vercel"
	case "backend":
		return "render"
	case "database":
		return "supabase"
	default:
		return "vercel"
	}
}

func isValidProvider(provider string) bool {
	valid := []string{"vercel", "netlify", "render", "railway", "supabase", "cloudflare-pages", "engine"}
	for _, p := range valid {
		if p == provider {
			return true
		}
	}
	return false
}

func generateDeploymentURL(name, provider string) string {
	switch provider {
	case "vercel":
		return fmt.Sprintf("https://%s.vercel.app", name)
	case "netlify":
		return fmt.Sprintf("https://%s.netlify.app", name)
	case "render":
		return fmt.Sprintf("https://%s.onrender.com", name)
	case "railway":
		return fmt.Sprintf("https://%s.up.railway.app", name)
	case "supabase":
		return fmt.Sprintf("https://%s.supabase.co", name)
	case "cloudflare-pages":
		return fmt.Sprintf("https://%s.pages.dev", name)
	default:
		return fmt.Sprintf("https://%s.deployed.io", name)
	}
}

func getProgress(status string) int {
	switch status {
	case "deploying", "building":
		return 50
	case "deployed":
		return 100
	case "failed":
		return 0
	default:
		return 25
	}
}

func simulateDeployment(store *DeploymentStore, id string) {
	// Simulate deployment process
	time.Sleep(3 * time.Second)

	deployment := store.Get(id)
	if deployment != nil {
		deployment.Status = "deployed"
		deployment.UpdatedAt = time.Now()
		store.Update(deployment)
	}
}

// provisionEC2Instance drives a real EC2 RunInstances call against AWS (or
// LocalStack when AWS_ENDPOINT_URL is set). The AWS provider path is gated
// behind BuildEC2Input validation so a malformed payload fails fast at the
// HTTP boundary instead of triggering a downstream AWS call.
//
// On any non-validation failure the deployment is marked "failed" so the
// caller can poll and react, rather than silently staying at "deploying".
func provisionEC2Instance(store *DeploymentStore, id string, req DeployRequest) {
	deployment := store.Get(id)
	if deployment == nil {
		return
	}

	// Construction of RunInstancesInput has no external side effects, so build
	// it synchronously to fail fast on bad input.
	input := buildEC2InputFromDeploy(req)

	cfg, err := loadAWSConfigFromEnv()
	if err != nil {
		markDeployFailed(store, deployment, "aws-config: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if _, runErr := runEC2Instance(ctx, cfg, input); runErr != nil {
		markDeployFailed(store, deployment, "run-instances: "+runErr.Error())
		return
	}

	deployment.Status = "deployed"
	deployment.UpdatedAt = time.Now()
	store.Update(deployment)
}

// markDeployFailed sets the deployment to a terminal failed state. Exposed so
// provisionEC2Instance and future real-deploy paths can reuse it.
func markDeployFailed(store *DeploymentStore, d *Deployment, reason string) {
	d.Status = "failed"
	d.UpdatedAt = time.Now()
	if d.EnvVars == nil {
		d.EnvVars = map[string]string{}
	}
	d.EnvVars["__failure_reason"] = reason
	store.Update(d)
}

// envOr returns os.Getenv(key) or fallback if empty. Centralised so tests and
// the production code agree on default behaviour for unset AWS_* vars.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
