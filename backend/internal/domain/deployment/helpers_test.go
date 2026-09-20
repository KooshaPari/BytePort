package deployment

import (
	"testing"
	"time"
)

// newTestDeployment returns a fresh Deployment with name="test", owner="owner",
// and no project. Replaces the 11 inline `dep, _ := NewDeployment("test",
// "owner", nil)` calls across deployment_test.go.
func newTestDeployment(t *testing.T) *Deployment {
	t.Helper()
	dep, err := NewDeployment("test", "owner", nil)
	if err != nil {
		t.Fatalf("NewDeployment() unexpected error: %v", err)
	}
	return dep
}

// reconstructForValidate returns a Deployment reconstructed via
// ReconstructDeployment with sensible defaults for the timestamps and
// optional pointers, varying only uuid/name/owner/status. Used by the
// TestDeployment_Validate table cases.
func reconstructForValidate(uuid, name, owner string, status Status) *Deployment {
	now := time.Now().UTC()
	return ReconstructDeployment(
		uuid,
		name,
		owner,
		nil,
		status,
		now,
		now,
		nil,
		nil,
	)
}

// transitionTo walks the deployment through the status state machine from
// its current state to the given target. Mirrors the if/else if/else if
// chain that previously appeared in TestDeployment_StatusTransitions and
// TestDeployment_IsActive.
func transitionTo(dep *Deployment, target Status) {
	switch target {
	case StatusPending:
		// No-op: NewDeployment already starts in StatusPending.
	case StatusDetecting:
		_ = dep.SetStatus(StatusDetecting)
	case StatusProvisioning:
		_ = dep.SetStatus(StatusDetecting)
		_ = dep.SetStatus(StatusProvisioning)
	case StatusDeploying:
		_ = dep.SetStatus(StatusDetecting)
		_ = dep.SetStatus(StatusProvisioning)
		_ = dep.SetStatus(StatusDeploying)
	case StatusDeployed:
		_ = dep.SetStatus(StatusDetecting)
		_ = dep.SetStatus(StatusProvisioning)
		_ = dep.SetStatus(StatusDeploying)
		_ = dep.SetStatus(StatusDeployed)
	case StatusTerminated, StatusFailed:
		_ = dep.SetStatus(target)
	}
}

// assertErr fails the test when (err==nil) != wantErr.
func assertErr(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if wantErr && err == nil {
		t.Errorf("expected error, got nil")
	}
	if !wantErr && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
