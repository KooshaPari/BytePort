package routes

import (
    "byteport/models"
    "net/http"
    "net/http/httptest"
    "testing"
)

// Evidence probe, not an acceptance oracle.
//
// A PASS here means the current implementation reproduced the unsafe
// side-effect window: the provider accepted a deployment, local persistence
// failed, and no compensating stop occurred. Once an accepted operation/
// reconciliation architecture exists, replace this observation probe with its
// desired-behavior oracle.
func TestRecoveryProbeRemoteDeploySurvivesLocalPersistenceFailure(t *testing.T) {
    db := testDB(t)
    stub := newNVMSStub(t)
    alice := models.User{UUID: "alice-uuid", Name: "alice"}

    sqlDB, err := db.DB()
    if err != nil {
        t.Fatalf("sql db: %v", err)
    }
    if err := sqlDB.Close(); err != nil {
        t.Fatalf("close db to inject persistence failure: %v", err)
    }

    w := httptest.NewRecorder()
    DeployProject(handlerContext(
        w,
        "/deploy",
        `{"name":"crash-window-app","repository":"owner/repo","repository_id":123}`,
        alice,
        true,
    ))

    if stub.deployCalls() != 1 {
        t.Fatalf("provider deploy calls = %d, want 1 to establish remote side effect", stub.deployCalls())
    }
    if w.Code < http.StatusInternalServerError {
        t.Fatalf("status = %d, want persistence failure surfaced as server error", w.Code)
    }
    if stub.stopCalls() != 0 {
        t.Fatalf("provider stop calls = %d, want 0 for this evidence probe; implementation now compensates and probe must be replaced", stub.stopCalls())
    }
}
