package sharedingress

import (
	"testing"
	"time"
)

func baseSnapshot() Snapshot {
	return Snapshot{
		ObservedAt: time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC),
		Source:     "cloudflare-read",
		Provider:   "cloudflare",
		AccountID:  "acct",
		Tunnel:     Tunnel{ID: "tunnel-1", Name: "shared"},
		Routes: []Route{
			{Hostname: "tracera.example.com", Service: "http://localhost:8080", Owner: "tracera", Audience: "machine", AccessAppID: "app-t"},
			{Hostname: "other.example.com", Service: "http://localhost:3000", Owner: "other", Audience: "human", AccessAppID: "app-o"},
			{Service: "http_status:404", Owner: "shared"},
		},
		DNS: []DNSRecord{
			{Name: "tracera.example.com", Type: "CNAME", Content: "tunnel.example", Proxied: true, Owner: "tracera"},
			{Name: "other.example.com", Type: "CNAME", Content: "tunnel.example", Proxied: true, Owner: "other"},
		},
		AccessApps: []AccessApp{
			{ID: "app-t", Domain: "tracera.example.com", Type: "self_hosted", Owner: "tracera", Audience: "machine"},
			{ID: "app-o", Domain: "other.example.com", Type: "self_hosted", Owner: "other", Audience: "human"},
		},
		Connectors: []Connector{{ID: "connector-1", Arch: "windows_amd64", Status: "healthy"}},
	}
}

func TestIdenticalSnapshotHasNoChanges(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()

	plan, err := BuildPlan(PlanInput{Live: live, Desired: desired, ManagedOwners: []string{"tracera"}})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if !plan.Safe() || len(plan.Changes) != 0 {
		t.Fatalf("expected safe no-op plan, got %#v", plan)
	}
}

func TestCannotSilentlyDeleteSiblingRoute(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	desired.Routes = desired.Routes[:1]

	plan, err := BuildPlan(PlanInput{Live: live, Desired: desired, ManagedOwners: []string{"tracera"}})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Safe() {
		t.Fatal("plan must be unsafe when desired state omits sibling-owned shared routes")
	}
	found := false
	for _, c := range plan.Conflicts {
		if c.ResourceType == "route" && c.LiveOwner == "other" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected conflict for sibling route, got %#v", plan.Conflicts)
	}
}

func TestManagedOwnerCanUpdateOwnRoute(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	desired.Routes[0].Service = "http://localhost:8081"

	plan, err := BuildPlan(PlanInput{Live: live, Desired: desired, ManagedOwners: []string{"tracera"}})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if !plan.Safe() {
		t.Fatalf("expected safe plan, got conflicts %#v", plan.Conflicts)
	}
	if len(plan.Changes) != 1 || plan.Changes[0].Kind != ChangeUpdate || plan.Changes[0].ResourceType != "route" {
		t.Fatalf("unexpected changes %#v", plan.Changes)
	}
}

func TestOwnershipTransferRequiresExplicitMigration(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	desired.DNS[0].Owner = "shared"

	plan, err := BuildPlan(PlanInput{Live: live, Desired: desired, ManagedOwners: []string{"tracera", "shared"}})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Safe() {
		t.Fatal("ownership transfer must be a conflict")
	}
	if got := plan.Conflicts[0].Reason; got != "ownership transfer requires explicit migration; planner will not infer it" {
		t.Fatalf("unexpected conflict reason %q", got)
	}
}

func TestTunnelReplacementIsAlwaysConflict(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	desired.Tunnel.ID = "different-tunnel"

	plan, err := BuildPlan(PlanInput{Live: live, Desired: desired, ManagedOwners: []string{"tracera"}})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Safe() {
		t.Fatal("tunnel replacement must not be planned implicitly")
	}
}

func TestSecretReferenceContainsNoCredentialValueField(t *testing.T) {
	s := baseSnapshot()
	s.AccessApps[0].MachineSecret = &SecretRef{Store: "github-actions", Key: "CF_ACCESS_CLIENT_SECRET"}

	if err := ValidateSnapshot(s); err != nil {
		t.Fatalf("ValidateSnapshot: %v", err)
	}
	if s.AccessApps[0].MachineSecret.Key != "CF_ACCESS_CLIENT_SECRET" {
		t.Fatal("secret reference changed unexpectedly")
	}
}

func TestDuplicateRouteKeyRejected(t *testing.T) {
	s := baseSnapshot()
	s.Routes = append(s.Routes, s.Routes[0])

	if err := ValidateSnapshot(s); err == nil {
		t.Fatal("expected duplicate route key to be rejected")
	}
}
