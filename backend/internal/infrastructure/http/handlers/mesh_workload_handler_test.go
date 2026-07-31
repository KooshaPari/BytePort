package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/byteport/api/internal/application/meshworkload"
	"github.com/gin-gonic/gin"
)

func TestMeshWorkloadSubmitUsesAuthenticatedOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewMeshWorkloadHandler(meshworkload.NewSubmitDesiredStateUseCase())
	router.POST("/mesh/workloads", func(c *gin.Context) { c.Set("user_uuid", "auth-user"); handler.Submit(c) })

	body := `{"owner":"attacker","name":"demo","composition_digest":"sha256:` + strings.Repeat("a", 64) + `","artifact_ref":"oci://registry/demo","execution_backend":"podman"}`
	req := httptest.NewRequest("POST", "/mesh/workloads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 202 {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"owner":"auth-user"`) {
		t.Fatalf("response did not use authenticated owner: %s", rec.Body.String())
	}
}

type conflictStore struct{}

func (conflictStore) Save(_ context.Context, _ string, _ meshworkload.DesiredStateRequest) error {
	return &meshworkload.ConflictError{Message: "workload already exists with a different composition digest"}
}

func TestMeshWorkloadSubmitMapsIdempotencyConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewMeshWorkloadHandler(meshworkload.NewSubmitDesiredStateUseCase(conflictStore{}))
	router.POST("/mesh/workloads", func(c *gin.Context) { c.Set("user_uuid", "auth-user"); handler.Submit(c) })

	body := `{"name":"demo","composition_digest":"sha256:` + strings.Repeat("a", 64) + `","artifact_ref":"oci://registry/demo","execution_backend":"podman"}`
	req := httptest.NewRequest("POST", "/mesh/workloads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 409 {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"CONFLICT"`) {
		t.Fatalf("response did not expose conflict code: %s", rec.Body.String())
	}
}
