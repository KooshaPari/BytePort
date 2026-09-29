package routes

import (
    "byteport/models"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"
)

// Mature-recovery adversarial control: product identity and provider runtime
// identity are deliberately different.  A test that uses equal IDs can hide a
// wrong-target termination bug.
func TestTerminateInstanceUsesPersistedProviderSandboxID(t *testing.T) {
    db := testDB(t)
    alice := models.User{UUID: "alice-uuid"}

    project := models.Project{
        UUID: "project-123",
        ID: "project-123",
        Owner: alice.UUID,
        Name: "app",
    }
    project.SetDeploy(map[string]models.Instance{
        "default": {
            UUID: "provider-sandbox-777",
            Owner: alice.UUID,
            Name: "app-runtime",
            Status: "running",
            ResUUID: "project-123",
        },
    })
    if err := db.Create(&project).Error; err != nil {
        t.Fatalf("seed project: %v", err)
    }

    var mu sync.Mutex
    var stoppedID string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/v1/stop" {
            t.Fatalf("unexpected path %s", r.URL.Path)
        }
        mu.Lock()
        stoppedID = r.URL.Query().Get("id")
        mu.Unlock()
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
    }))
    t.Cleanup(server.Close)
    t.Setenv("NVMS_URL", server.URL)

    w := httptest.NewRecorder()
    TerminateInstance(handlerContext(
        w,
        "/terminate",
        `{"uuid":"project-123"}`,
        alice,
        true,
    ))

    if w.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
    }

    mu.Lock()
    got := stoppedID
    mu.Unlock()
    if got != "provider-sandbox-777" {
        t.Fatalf(
            "wrong runtime targeted: stop id = %q, want persisted provider id %q (project id is %q)",
            got,
            "provider-sandbox-777",
            "project-123",
        )
    }
}
