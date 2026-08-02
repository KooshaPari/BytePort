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
	code, body := submitMeshWorkload(t, "user_uuid", "a", "attacker", nil)
	assertMeshResponse(t, code, body, 202, `"owner":"auth-user"`, "authenticated owner")
}

func TestMeshWorkloadSubmitAcceptsLegacyUserIDIdentityAlias(t *testing.T) {
	code, body := submitMeshWorkload(t, "user_id", "b", "", nil)
	assertMeshResponse(t, code, body, 202, `"owner":"auth-user"`, "user_id identity alias")
}

type conflictStore struct{}

func (conflictStore) Save(_ context.Context, _ string, _ meshworkload.DesiredStateRequest) error {
	return &meshworkload.ConflictError{Message: "workload already exists with a different composition digest"}
}

func TestMeshWorkloadSubmitMapsIdempotencyConflict(t *testing.T) {
	code, body := submitMeshWorkload(t, "user_uuid", "a", "", conflictStore{})
	assertMeshResponse(t, code, body, 409, `"code":"CONFLICT"`, "conflict code")
}

func submitMeshWorkload(t *testing.T, identityKey, digest, owner string, store meshworkload.DesiredStateSaver) (int, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	useCase := meshworkload.NewSubmitDesiredStateUseCase()
	if store != nil {
		useCase = meshworkload.NewSubmitDesiredStateUseCase(store)
	}
	handler := NewMeshWorkloadHandler(useCase)
	router.POST("/mesh/workloads", func(c *gin.Context) { c.Set(identityKey, "auth-user"); handler.Submit(c) })

	body := `{"owner":"` + owner + `","name":"demo","composition_digest":"sha256:` + strings.Repeat(digest, 64) + `","artifact_ref":"oci://registry/demo","execution_backend":"podman"}`
	req := httptest.NewRequest("POST", "/mesh/workloads", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func assertMeshResponse(t *testing.T, code int, body string, expectedCode int, expectedBody, description string) {
	t.Helper()
	if code != expectedCode {
		t.Fatalf("expected %d, got %d: %s", expectedCode, code, body)
	}
	if !strings.Contains(body, expectedBody) {
		t.Fatalf("response did not expose %s: %s", description, body)
	}
}
