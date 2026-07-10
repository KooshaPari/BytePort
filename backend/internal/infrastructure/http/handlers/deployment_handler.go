package handlers

import (
	"net/http"
	"time"

	"github.com/byteport/api/internal/application/deployment"
	"github.com/byteport/api/internal/infrastructure/clients"
	"github.com/gin-gonic/gin"
)

// DeploymentHandler handles HTTP requests for deployments.
//
// engineClient is OPTIONAL (may be nil). When set AND the request selects
// provider="engine", CreateDeployment forwards the request to the Rust
// byteport-engine daemon over UDS instead of the use-case path. nil clients
// fall through to the use case, preserving the original contract.
type DeploymentHandler struct {
	createUseCase       *deployment.CreateDeploymentUseCase
	getUseCase          *deployment.GetDeploymentUseCase
	listUseCase         *deployment.ListDeploymentsUseCase
	terminateUseCase    *deployment.TerminateDeploymentUseCase
	updateStatusUseCase *deployment.UpdateStatusUseCase
	engineClient        *clients.EngineDaemonClient
}

// NewDeploymentHandler creates a new handler. Pass nil for engineClient when
// the engine daemon is not configured (BYTEPORT_ENGINE_SOCKET unset).
func NewDeploymentHandler(
	createUseCase *deployment.CreateDeploymentUseCase,
	getUseCase *deployment.GetDeploymentUseCase,
	listUseCase *deployment.ListDeploymentsUseCase,
	terminateUseCase *deployment.TerminateDeploymentUseCase,
	updateStatusUseCase *deployment.UpdateStatusUseCase,
	engineClient *clients.EngineDaemonClient,
) *DeploymentHandler {
	return &DeploymentHandler{
		createUseCase:       createUseCase,
		getUseCase:          getUseCase,
		listUseCase:         listUseCase,
		terminateUseCase:    terminateUseCase,
		updateStatusUseCase: updateStatusUseCase,
		engineClient:        engineClient,
	}
}

// RegisterRoutes registers all deployment routes
func (h *DeploymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	deployments := router.Group("/deployments")
	{
		deployments.POST("", h.CreateDeployment)
		deployments.GET("", h.ListDeployments)
		deployments.GET("/:uuid", h.GetDeployment)
		deployments.DELETE("/:uuid", h.TerminateDeployment)
		deployments.PATCH("/:uuid/status", h.UpdateStatus)
	}
}

// CreateDeployment creates a new deployment
// @Summary Create deployment
// @Tags deployments
// @Accept json
// @Produce json
// @Param deployment body deployment.CreateDeploymentRequest true "Deployment data"
// @Success 201 {object} deployment.CreateDeploymentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /deployments [post]
func (h *DeploymentHandler) CreateDeployment(c *gin.Context) {
	var req deployment.CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Engine-provider short-circuit. When the hexagonal handler is wired
	// to an engine daemon (BYTEPORT_ENGINE_SOCKET set) AND the request
	// asks for the engine provider, forward to the Rust daemon rather
	// than running the in-process use case. The use-case path still
	// serves every other provider.
	if req.Provider == "engine" && h.engineClient != nil {
		h.createViaEngine(c, &req)
		return
	}

	response, err := h.createUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		handleApplicationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// createViaEngine forwards a hexagonal /api/v1/deployments request whose
// provider is "engine" to the Rust byteport-engine daemon over UDS. The
// returned deployment_id is wrapped in the standard CreateDeploymentResponse
// so callers do not need to special-case the engine contract.
func (h *DeploymentHandler) createViaEngine(c *gin.Context, req *deployment.CreateDeploymentRequest) {
	svcName := req.Name
	if svcName == "" {
		svcName = req.Owner
	}
	env := make([]clients.EnvEntry, 0, len(req.EnvVars))
	for k, v := range req.EnvVars {
		env = append(env, clients.EnvEntry{Key: k, Value: v})
	}
	repoURL, _ := req.Config["git_url"].(string)

	daemonReq := &clients.DeployRequest{
		Name:       req.Name,
		User:       clients.DeployUser{ID: req.Owner, Email: ""},
		Repository: repoURL,
		Services: []clients.ServiceEntry{{
			Name: svcName,
			Path: repoURL,
			Port: 80,
			Env:  env,
		}},
	}

	result, err := h.engineClient.Deploy(c.Request.Context(), daemonReq)
	if err != nil {
		if err == clients.ErrDaemonUnavailable {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Error: "engine daemon unavailable",
				Code:  "ENGINE_UNAVAILABLE",
			})
			return
		}
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Error: "engine daemon error: " + err.Error(),
			Code:  "ENGINE_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, deployment.CreateDeploymentResponse{
		UUID:      result.DeploymentID,
		Name:      req.Name,
		Owner:     req.Owner,
		Status:    "deploying",
		CreatedAt: nowUTC(),
		Message:   "Engine deployment started",
	})
}

// GetDeployment retrieves a deployment by UUID
// @Summary Get deployment
// @Tags deployments
// @Produce json
// @Param uuid path string true "Deployment UUID"
// @Success 200 {object} deployment.GetDeploymentResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /deployments/{uuid} [get]
func (h *DeploymentHandler) GetDeployment(c *gin.Context) {
	uuid := c.Param("uuid")
	userUUID := getUserUUID(c) // Helper to extract user from context

	response, err := h.getUseCase.Execute(c.Request.Context(), uuid, userUUID)
	if err != nil {
		handleApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListDeployments lists deployments with optional filtering
// @Summary List deployments
// @Tags deployments
// @Produce json
// @Param owner query string false "Filter by owner"
// @Param status query string false "Filter by status"
// @Param offset query int false "Pagination offset"
// @Param limit query int false "Pagination limit"
// @Success 200 {object} deployment.ListDeploymentsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /deployments [get]
func (h *DeploymentHandler) ListDeployments(c *gin.Context) {
	var req deployment.ListDeploymentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid query parameters",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	response, err := h.listUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		handleApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// TerminateDeployment terminates a deployment
// @Summary Terminate deployment
// @Tags deployments
// @Produce json
// @Param uuid path string true "Deployment UUID"
// @Success 200 {object} deployment.TerminateDeploymentResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /deployments/{uuid} [delete]
func (h *DeploymentHandler) TerminateDeployment(c *gin.Context) {
	uuid := c.Param("uuid")
	userUUID := getUserUUID(c)

	response, err := h.terminateUseCase.Execute(c.Request.Context(), uuid, userUUID)
	if err != nil {
		handleApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateStatus updates deployment status
// @Summary Update deployment status
// @Tags deployments
// @Accept json
// @Produce json
// @Param uuid path string true "Deployment UUID"
// @Param status body deployment.UpdateStatusRequest true "Status update"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /deployments/{uuid}/status [patch]
func (h *DeploymentHandler) UpdateStatus(c *gin.Context) {
	uuid := c.Param("uuid")
	userUUID := getUserUUID(c)

	var req deployment.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	err := h.updateStatusUseCase.Execute(c.Request.Context(), uuid, req, userUUID)
	if err != nil {
		handleApplicationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// handleApplicationError converts application errors to HTTP responses
func handleApplicationError(c *gin.Context, err error) {
	if appErr, ok := err.(*deployment.ApplicationError); ok {
		c.JSON(appErr.StatusCode, ErrorResponse{
			Error: appErr.Message,
			Code:  appErr.Code,
		})
		return
	}

	// Unknown error
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "internal server error",
		Code:  "INTERNAL_ERROR",
	})
}

// getUserUUID extracts user UUID from context (set by auth middleware)
func getUserUUID(c *gin.Context) string {
	// This would be set by authentication middleware
	if userUUID, exists := c.Get("user_uuid"); exists {
		if uuid, ok := userUUID.(string); ok {
			return uuid
		}
	}
	return ""
}

// nowUTC is overridable so tests can pin the time without monkey-patching.
// Centralised so the engine and use-case response timestamps stay aligned.
var nowUTC = func() time.Time { return time.Now().UTC() }
