package routes

import (
    "byteport/models"
    "net/http"
    "net/http/httptest"
    "testing"

    "gorm.io/gorm"
)

// failProjectCreate installs a GORM create callback that fails only when the
// handler reaches local Project persistence. The database remains readable and
// usable before the provider call, so this injects failure at the intended
// remote-success/local-commit boundary rather than before deployment starts.
func failProjectCreate(t *testing.T, db *gorm.DB) {
    t.Helper()
    const callback = "recovery:fail-project-create"
    if err := db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
        if tx.Statement != nil && tx.Statement.Schema != nil && tx.Statement.Schema.Name == "Project" {
            tx.AddError(gorm.ErrInvalidTransaction)
        }
    }); err != nil {
        t.Fatalf("register create failure callback: %v", err)
    }
    t.Cleanup(func() { _ = db.Callback().Create().Remove(callback) })
}

// Evidence probe, not a desired-behavior acceptance oracle.
//
// PASS means the current implementation reproduced BP-F04's unsafe window:
// the provider accepted a deployment, local Project persistence then failed,
// and no compensating stop was issued. Once recovery architecture is accepted,
// replace this observation probe with the chosen journal/reconcile/adopt/
// compensate behavior.
func TestRecoveryProbeRemoteDeploySurvivesLocalPersistenceFailure(t *testing.T) {
    db := testDB(t)
    stub := newNVMSStub(t)
    alice := models.User{UUID: "alice-uuid", Name: "alice"}
    failProjectCreate(t, db)

    w := httptest.NewRecorder()
    DeployProject(handlerContext(
        w,
        "/deploy",
        `{"name":"crash-window-app","repository":{"id":123,"name":"repo"}}`,
        alice,
        true,
    ))

    if got := stub.deployCalls(); got != 1 {
        t.Fatalf("provider deploy calls = %d, want 1 to establish remote side effect", got)
    }
    if w.Code < http.StatusInternalServerError {
        t.Fatalf("status = %d, want local persistence failure surfaced as server error", w.Code)
    }
    if got := stub.stopCalls(); got != 0 {
        t.Fatalf("provider stop calls = %d, want 0 for evidence probe; implementation now compensates and probe must be replaced", got)
    }

    var projects int64
    if err := db.Model(&models.Project{}).Where("name = ?", "crash-window-app").Count(&projects).Error; err != nil {
        t.Fatalf("count local projects: %v", err)
    }
    if projects != 0 {
        t.Fatalf("local project rows = %d, want 0 after injected create failure", projects)
    }
}
