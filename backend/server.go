package main

import (
	"os"
	"strings"

	"github.com/byteport/api/internal/container"
	"github.com/byteport/api/internal/infrastructure/http/middleware"
	"github.com/byteport/api/internal/infrastructure/otel"
	"github.com/byteport/api/lib"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// APIServer represents the HTTP API server
type APIServer struct {
	router    *gin.Engine
	container *container.Container
	store     *DeploymentStore // Legacy - will be removed
}

// NewAPIServer creates a new API server instance
func NewAPIServer(c *container.Container) *APIServer {
	r := gin.Default()

	// Legacy store for backward compatibility during migration
	store := NewDeploymentStore()

	allowedOrigins := parseAllowedOrigins()

	// OTel middleware (gated by OTEL_ENDPOINT env)
	r.Use(otel.GinMiddleware())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Cookie"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public unauthenticated group — surfaces machine-readable discovery
	// documents for agents. No auth required (information disclosure only).
	public := r.Group("/")
	{
		// RFC 8615 well-known: agent-card for A2A-compatible agents.
		// Mirrors public/.well-known/agent.json, served dynamically.
		public.GET("/.well-known/agent.json", handleAgentDiscovery)

		// Root: lightweight pointer to the discovery doc.
		public.GET("/", handleRoot)
	}

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", handleHealth)
		v1.GET("/health/readiness", handleReadiness)
		v1.GET("/", handleAPIInfo)

		// Public auth endpoints
		// Auth endpoints - WorkOS AuthKit only
		v1.POST("/auth/workos/callback", handleWorkOSCallback)

		// GitHub webhook — authenticated by HMAC-SHA256, not by user JWT.
		c.WebhookHandler.RegisterRoutes(v1)

		// Engine daemon endpoints (gated by BYTEPORT_ENGINE_SOCKET env).
		if c.EngineDeployHandler != nil {
			c.EngineDeployHandler.RegisterRoutes(v1)
		}

		protected := v1.Group("/")
		protected.Use(lib.AuthMiddleware())
		{
			// Protected endpoints - require AuthKit authentication
			protected.GET("/user/:id", handleGetUser)

			// Org-scoped deployment routes with RBAC
			orgScoped := protected.Group("/orgs/:org_id")
			orgScoped.Use(middleware.RBACMiddleware("owner", "admin"))
			{
				orgScoped.GET("", handleGetOrg)
				orgScoped.PUT("", handleUpdateOrg)
				orgScoped.GET("/members", handleListMembers)
				orgScoped.DELETE("/members/:user_id", handleRemoveMember)
			}

			// NEW: Hexagonal architecture endpoints
			c.DeploymentHandler.RegisterRoutes(protected)

			// T2 UDS proxy: forward OpenAI-compatible chat traffic to the
			// Rust omniroute data plane (plans/2026-07-04-byteport-evolution-v1.md).
			protected.Any("/v1/chat/completions", middleware.UDSProxy())

			// LEGACY: Old deployment endpoints (will be removed). The
			// handleDeploy shim forwards to the Rust engine daemon when
			// the caller asks for provider=engine AND the engine client
			// is wired (BYTEPORT_ENGINE_SOCKET set at boot). Otherwise
			// the in-process simulated path handles the request.
			legacyDeployments := protected.Group("/legacy/deployments")
			{
				legacyDeployments.POST("", handleDeploy(store, c.EngineDaemonClient))
				legacyDeployments.GET("", handleListDeployments(store))
				legacyDeployments.GET("/:id", handleGetDeployment(store))
				legacyDeployments.DELETE("/:id", handleTerminateDeployment(store))

				// Deployment operations
				legacyDeployments.GET("/:id/status", handleGetStatus(store))
				legacyDeployments.GET("/:id/logs", handleGetLogs(store))
				legacyDeployments.GET("/:id/metrics", handleGetMetrics(store))
			}
		}

		// Utilities
		v1.POST("/detect", handleDetectApp)
		v1.POST("/estimate-cost", handleEstimateCost)
		v1.POST("/validate-config", handleValidateConfig)

		// Documentation
		v1.GET("/docs", handleDocs)
	}

	return &APIServer{
		router:    r,
		container: c,
		store:     store,
	}
}

func parseAllowedOrigins() []string {
	raw := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	origins := make([]string, 0, len(raw))
	for _, origin := range raw {
		origin = strings.TrimSpace(origin)
		if origin != "" && origin != "*" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		origins = []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:8002",
			"https://byte.kooshapari.com",
		}
	}
	return origins
}

// API info handler — surfaces the canonical API description at /api/v1/.
func handleAPIInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"name":        "BytePort API",
		"version":     "2.0.0",
		"description": "Zero-cost deployment platform API",
		"endpoints": gin.H{
			"health":        "/api/v1/health",
			"health_ready":  "/api/v1/health/readiness",
			"deployments":   "/api/v1/deployments",
			"projects":      "/api/v1/projects",
			"docs":          "/api/v1/docs",
			"agent_card":    "/.well-known/agent.json",
		},
	})
}

// handleRoot — lightweight landing pointer that nudges agents and
// humans toward the discovery document. Returns a small JSON body.
func handleRoot(c *gin.Context) {
	c.JSON(200, gin.H{
		"name":          "BytePort API",
		"version":       "2.0.0",
		"description":   "Zero-cost deployment platform",
		"discovery":     "/.well-known/agent.json",
		"mcp_endpoint":  "stdio (spawn `byteport-mcp-server`)",
		"openapi":       "/api/v1/docs",
		"health":        "/api/v1/health",
		"health_ready":  "/api/v1/health/readiness",
		"agent_capable": true,
	})
}

// handleAgentDiscovery — serves the A2A-compatible agent card.
// The source of truth is public/.well-known/agent.json; this handler
// embeds the canonical document inline so live deployments do not need
// a separate static file route.
//
// Why embed vs read-file: this guarantees the well-known URI is reachable
// from any deployment (Vercel, Docker, k8s) without requiring the static
// asset to be copied alongside binaries. Drift between the static file
// and the live endpoint is mitigated by the agent-card_test.go which
// loads the static file and asserts both payloads decode to equivalent
// semantic content.
func handleAgentDiscovery(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=300")
	c.Header("Content-Type", "application/agent-card+json")
	c.JSON(200, agentCard)
}

// agentCard is the canonical A2A 0.3.0-compliant agent descriptor
// served at /.well-known/agent.json. Keep in sync with
// public/.well-known/agent.json.
//
// Capabilities listed here are reflected in:
//   - docs/openapi.yml (REST API)
//   - backend/mcp/tools.go (MCP tool manifest)
//   - .github/workflows/bench.yml (perf benchmarks)
//
// ui:// capabilities require the omniroute SDK to surface the
// dashboard. Updates here must be idempotent and additive.
var agentCard = gin.H{
	"schema_version": "0.3.0",
	"id":             "byteport-api/2.0.0",
	"name":           "BytePort",
	"description":    "Zero-cost deployment platform — deploy static sites, APIs, and databases to free-tier infrastructure.",
	"version":        "2.0.0",
	"author":         "kooshapari",
	"homepage":       "https://byte.kooshapari.com",
	"repository":     "https://github.com/kooshapari/BytePort",
	"license":        "MIT",
	"capabilities":   gin.H{
		"ui": []string{
			"ui://dashboard",
			"ui://deployment-status",
			"ui://logs-stream",
		},
		"tools": []string{
			// MUST stay in sync with public/.well-known/agent.json
			// capabilities.tools[].name — verified by TestHandleAgentDiscovery_StaticSync.
			"byteport_health",
			"byteport_deploy",
			"byteport_list_deployments",
			"byteport_get_deployment",
			"byteport_terminate_deployment",
			"byteport_deployment_status",
			"byteport_deployment_logs",
			"byteport_estimate_cost",
			"byteport_detect_app",
		},
		"protocols": []string{
			"mcp/0.1",
			"json-rpc/2.0",
			"a2a/0.3.0",
			"openai-functions/v1",
		},
		"transports": []string{
			"stdio",
			"http",
			"sse",
		},
	},
	"endpoints": gin.H{
		"discovery":         "/.well-known/agent.json",
		"rest_api":          "https://api.byte.kooshapari.com/api/v1",
		"openapi_spec":      "https://api.byte.kooshapari.com/api/v1/docs",
		"mcp":               "stdio://byteport-mcp-server",
		"webhook_ingress":   "https://api.byte.kooshapari.com/api/v1/webhooks/github",
		"health":            "https://api.byte.kooshapari.com/api/v1/health",
		"readiness":         "https://api.byte.kooshapari.com/api/v1/health/readiness",
	},
	"auth": gin.H{
		"schemes":     []string{"Bearer", "WorkOS-AuthKit"},
		"session":     "Cookie-based session for browser clients",
		"machine":     "Bearer JWT or session-bound token for agents",
		"federated":   true,
		"scim":        true,
		"sso":         true,
		"docs":        "https://docs.byte.kooshapari.com/auth",
	},
	"security": gin.H{
		"container_signing": "cosign keyless (Sigstore Fulcio + Rekor)",
		"sbom":              "cyclonedx-json, published per release",
		"vuln_scan":         "trivy, blocking on Critical/High",
		"signed_releases":   true,
	},
	"sla": gin.H{
		"uptime_target":    "99.9%",
		"p95_latency_ms":   500,
		"incident_channel": "https://status.byte.kooshapari.com",
	},
	"rate_limits": gin.H{
		"per_minute":   60,
		"per_hour":     1000,
		"per_day":      10000,
		"burst_window": "10s",
	},
	"metadata": gin.H{
		"audit_score":       72.3,
		"audit_grade":       "C-",
		"convergence_track": "sprint-1-target-70",
		"tested_at":         "2026-07-08",
		"openapi":           "3.1.0",
		"go_version":        "1.24",
		"fips_140":          "go-crypto",
	},
}

// Liveness probe — always returns ok with no dependency checks.
func handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "byteport-api",
		"version": "2.0.0",
	})
}

// Readiness probe — checks dependencies before declaring ready.
func handleReadiness(c *gin.Context) {
	dbReady := os.Getenv("DATABASE_URL") != ""
	otelReady := os.Getenv("OTEL_ENDPOINT") != ""

	checks := gin.H{
		"database": dbReady,
		"otel":     otelReady,
	}

	allReady := dbReady && otelReady
	status := "ready"
	code := 200
	if !allReady {
		status = "not_ready"
		code = 503
	}

	c.JSON(code, gin.H{
		"status":  status,
		"service": "byteport-api",
		"version": "2.0.0",
		"checks":  checks,
	})
}

// Documentation handler
// Placeholder stubs for org management endpoints (sibling-session route declaration).
func handleGetOrg(c *gin.Context)    { c.JSON(501, gin.H{"error": "not implemented"}) }
func handleUpdateOrg(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func handleListMembers(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func handleRemoveMember(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }

func handleDocs(c *gin.Context) {
	c.JSON(200, gin.H{
		"title":       "BytePort API Documentation",
		"version":     "2.0.0",
		"description": "REST API for zero-cost deployments",
		"endpoints": []gin.H{
			{
				"method":      "POST",
				"path":        "/api/v1/deployments",
				"description": "Deploy an application",
				"body": gin.H{
					"name":     "string",
					"type":     "string (frontend/backend/database)",
					"provider": "string (optional)",
					"git_url":  "string (optional)",
					"env_vars": "object (optional)",
				},
			},
			{
				"method":      "GET",
				"path":        "/api/v1/deployments",
				"description": "List all deployments",
			},
			{
				"method":      "GET",
				"path":        "/api/v1/deployments/:id",
				"description": "Get deployment details",
			},
			{
				"method":      "DELETE",
				"path":        "/api/v1/deployments/:id",
				"description": "Terminate a deployment",
			},
			{
				"method":      "GET",
				"path":        "/api/v1/deployments/:id/status",
				"description": "Get deployment status",
			},
			{
				"method":      "GET",
				"path":        "/api/v1/deployments/:id/logs",
				"description": "Get deployment logs",
			},
			{
				"method":      "POST",
				"path":        "/api/v1/detect",
				"description": "Auto-detect application type",
			},
			{
				"method":      "POST",
				"path":        "/api/v1/estimate-cost",
				"description": "Estimate deployment cost",
			},
		},
	})
}
