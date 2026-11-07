package routes

import (
    "byteport/models"
    "net/http/httptest"
    "testing"
)

// Evidence probe, not a desired-behavior oracle.
//
// It asks whether two indistinguishable deploy requests currently share a
// durable product operation/idempotency identity. A PASS means the current
// handler performs two provider creates and persists two project identities,
// which is evidence that product-level idempotency/reconciliation is absent on
// this path. Once an Operation contract is implemented, replace this with the
// accepted same-operation retry semantics.
func TestRecoveryProbeRepeatedDeployCreatesTwoRemoteOperations(t *testing.T) {
    db := testDB(t)
    stub := newNVMSStub(t)
    alice := models.User{UUID: "alice-uuid", Name: "alice"}

    body := `{"name":"retry-app","repository":{"id":123,"name":"repo"},"repository_id":"123"}`

    first := httptest.NewRecorder()
    DeployProject(handlerContext(first, "/deploy", body, alice, true))
    if first.Code != 200 {
        t.Fatalf("first deploy status = %d: %s", first.Code, first.Body.String())
    }

    second := httptest.NewRecorder()
    DeployProject(handlerContext(second, "/deploy", body, alice, true))
    if second.Code != 200 {
        t.Fatalf("second deploy status = %d: %s", second.Code, second.Body.String())
    }

    if got := stub.deployCalls(); got != 2 {
        t.Fatalf("provider deploy calls = %d, want 2 for current no-idempotency evidence probe", got)
    }

    var count int64
    if err := db.Model(&models.Project{}).
        Where("name = ? AND owner = ?", "retry-app", alice.UUID).
        Count(&count).Error; err != nil {
        t.Fatalf("count retry projects: %v", err)
    }
    if count != 2 {
        t.Fatalf("persisted retry-app projects = %d, want 2 for current behavior probe", count)
    }
}
