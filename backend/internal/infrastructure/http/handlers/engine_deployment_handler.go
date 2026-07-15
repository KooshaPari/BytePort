package handlers

import (
	"net/http"

	"github.com/byteport/api/internal/infrastructure/clients"
	"github.com/gin-gonic/gin"
)

// EngineDeploymentHandler exposes the engine daemon behind HTTP endpoints.
// It wraps an EngineDaemonClient and translates UDS daemon responses into
// HTTP-level JSON responses.
type EngineDeploymentHandler struct {
	client *clients.EngineDaemonClient
}

// NewEngineDeploymentHandler creates a handler backed by the given client.
func NewEngineDeploymentHandler(client *clients.EngineDaemonClient) *EngineDeploymentHandler {
	return &EngineDeploymentHandler{client: client}
}

// RegisterRoutes registers engine-related routes on a Gin router group.
func (h *EngineDeploymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	engine := router.Group("/engine")
	{
		engine.POST("/deploy", h.Deploy)
		engine.POST("/deployments/:id/stop", h.Stop)
		engine.GET("/health", h.Health)
	}
}

// Deploy handles POST /api/v1/engine/deploy.
func (h *EngineDeploymentHandler) Deploy(c *gin.Context) {
	var req clients.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.client.Deploy(c.Request.Context(), &req)
	if err != nil {
		if err == clients.ErrDaemonUnavailable {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "engine daemon unavailable",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Stop handles POST /api/v1/engine/deployments/:id/stop.
func (h *EngineDeploymentHandler) Stop(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "deployment id is required",
		})
		return
	}

	err := h.client.Stop(c.Request.Context(), id)
	if err != nil {
		if err == clients.ErrDaemonUnavailable {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "engine daemon unavailable",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "stopped",
	})
}

// Health handles GET /api/v1/engine/health.
func (h *EngineDeploymentHandler) Health(c *gin.Context) {
	healthy := h.client.Health(c.Request.Context())
	if !healthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"engine": "unreachable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"engine": "connected",
	})
}
