package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/byteport/api/internal/application/meshworkload"
	"github.com/gin-gonic/gin"
)

// MeshWorkloadHandler accepts provider-neutral compute-mesh desired state.
type MeshWorkloadHandler struct {
	useCase *meshworkload.SubmitDesiredStateUseCase
}

// NewMeshWorkloadHandler constructs a mesh desired-state handler.
func NewMeshWorkloadHandler(useCase *meshworkload.SubmitDesiredStateUseCase) *MeshWorkloadHandler {
	return &MeshWorkloadHandler{useCase: useCase}
}

// RegisterRoutes registers the owner-scoped mesh workload endpoint.
func (h *MeshWorkloadHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/mesh/workloads", h.Submit)
}

// Submit validates and acknowledges desired state. The authenticated owner is
// read from middleware context; request bodies cannot impersonate another owner.
func (h *MeshWorkloadHandler) Submit(c *gin.Context) {
	var req meshworkload.DesiredStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body", Code: "INVALID_REQUEST"})
		return
	}
	owner := getUserUUID(c)
	response, err := h.useCase.Execute(c.Request.Context(), owner, req)
	if err != nil {
		var validationErr *meshworkload.ValidationError
		switch {
		case errors.As(err, &validationErr):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: validationErr.Error(), Code: "VALIDATION_ERROR"})
		case errors.Is(err, context.Canceled):
			c.JSON(http.StatusRequestTimeout, ErrorResponse{Error: "request canceled", Code: "REQUEST_CANCELED"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error", Code: "INTERNAL_ERROR"})
		}
		return
	}
	c.JSON(http.StatusAccepted, response)
}
