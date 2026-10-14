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
			{
				Hostname:    "tracera.example.com",
				Service:     "http://localhost:8080",
				Owner:       "tracera",
				Audience:    "machine",
				AccessAppID: "app-t",
			},
			{
				Hostname:    "other.example.com",
				Service:     "http://localhost:3000",
				Owner:       "other",
				Audience:    "human",
				AccessAppID: "app-o",
			},
			{Service: "http_status:404", Owner: "shared"},
		},
		DNS: []DNSRecord{
			{
				ID: "dns-tracera", Name: "tracera.example.com", Type: "CNAME",
				Content: "tunnel.example", Proxied: true, Owner: "tracera",
			},
			{
				ID: "dns-other", Name: "other.example.com", Type: "CNAME",
				Content: "tunnel.example", Proxied: true, Owner: "other",
			},
		},
		AccessApps: []AccessApp{
			{
				ID: "app-t", Domain: "tracera.example.com", Type: "self_hosted",
				Owner: "tracera", Audience: "machine",
				PolicyRefs: []PolicyRef{
					{
						ID: "service-auth", Decision: "non_identity", Precedence: 1,
						IncludeKinds: []string{"service_token"},
					},
					{
						ID: "human-allow", Decision: "allow", Precedence: 2,
						IncludeKinds: []string{"email"},
					},
				},
			},
			{
				ID: "app-o", Domain: "other.example.com", Type: "self_hosted",
				Owner: "other", Audience: "human",
				PolicyRefs: []PolicyRef{
					{
						ID: "other-human", Decision: "allow", Precedence: 1,
						IncludeKinds: []string{"email"},
					},
				},
			},
		},
		Connectors: []Connector{
			{ID: "connector-1", Arch: "windows_amd64", Status: "healthy"},
		},
	}
}

func TestIdenticalSnapshotHasNoChanges(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()

	plan, err := BuildPlan(PlanInput{
		Live: live, Desired: desired, ManagedOwners: []string{"tracera"},
	})
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
	desired.Routes = []Route{desired.Routes[0], desired.Routes[2]}

	plan, err := BuildPlan(PlanInput{
		Live: live, Desired: desired, ManagedOwners: []string{"tracera"},
	})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Safe() {
		t.Fatal("plan must be unsafe when desired state omits sibling-owned shared routes")
	}

	found := false
	for _, conflict := range plan.Conflicts {
		if conflict.ResourceType == "route" && conflict.LiveOwner == "other" {
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

	plan, err := BuildPlan(PlanInput{
		Live: live, Desired: desired, ManagedOwners: []string{"tracera"},
	})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if !plan.Safe() {
		t.Fatalf("expected safe plan, got conflicts %#v", plan.Conflicts)
	}
	if len(plan.Changes) != 1 ||
		plan.Changes[0].Kind != ChangeUpdate ||
		plan.Changes[0].ResourceType != "route" {
		t.Fatalf("unexpected changes %#v", plan.Changes)
	}
}

func TestRouteReorderIsDriftAndProtectsSiblingOwner(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	desired.Routes[0], desired.Routes[1] = desired.Routes[1], desired.Routes[0]

	plan, err := BuildPlan(PlanInput{
		Live: live, Desired: desired, ManagedOwners: []string{"tracera"},
	})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Safe() {
		t.Fatal("route reorder touching an unmanaged sibling must be unsafe")
	}

	var ownUpdate, siblingConflict bool
	for _, change := range plan.Changes {
		if change.ResourceType == "route" && change.Owner == "tracera" {
			ownUpdate = true
		}
	}
	for _, conflict := range plan.Conflicts {
		if conflict.ResourceType == "route" && conflict.LiveOwner == "other" {
			siblingConflict = true
		}
	}
	if !ownUpdate || !siblingConflict {
		t.Fatalf("expected own update + sibling conflict, got plan %#v", plan)
	}
}

func TestCloudflareCatchAllMustBeLast(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.Routes[1], snapshot.Routes[2] = snapshot.Routes[2], snapshot.Routes[1]

	if err := ValidateSnapshot(snapshot); err == nil {
		t.Fatal("expected non-terminal Cloudflare catch-all to be rejected")
	}
}

func TestAccessPolicyOrderIsAuthorizationDrift(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	policies := desired.AccessApps[0].PolicyRefs
	policies[0], policies[1] = policies[1], policies[0]
	desired.AccessApps[0].PolicyRefs = policies

	plan, err := BuildPlan(PlanInput{
		Live: live, Desired: desired, ManagedOwners: []string{"tracera"},
	})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if !plan.Safe() {
		t.Fatalf("managed Access policy reorder should be plan-able, got %#v", plan.Conflicts)
	}
	if len(plan.Changes) != 1 ||
		plan.Changes[0].ResourceType != "access_app" ||
		plan.Changes[0].Kind != ChangeUpdate {
		t.Fatalf("expected one access_app update, got %#v", plan.Changes)
	}
}

func TestDuplicateAccessPolicyPrecedenceRejected(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.AccessApps[0].PolicyRefs[1].Precedence = 1

	if err := ValidateSnapshot(snapshot); err == nil {
		t.Fatal("expected duplicate Access policy precedence to be rejected")
	}
}

func TestOwnershipTransferRequiresExplicitMigration(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	desired.DNS[0].Owner = "shared"

	plan, err := BuildPlan(PlanInput{
		Live: live, Desired: desired, ManagedOwners: []string{"tracera", "shared"},
	})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Safe() {
		t.Fatal("ownership transfer must be a conflict")
	}
	if got := plan.Conflicts[0].Reason; got !=
		"ownership transfer requires explicit migration; planner will not infer it" {
		t.Fatalf("unexpected conflict reason %q", got)
	}
}

func TestDesiredOnlyUnmanagedConflictHasNoLiveOwner(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	desired.DNS = append(desired.DNS, DNSRecord{
		ID: "dns-new", Name: "new.example.com", Type: "A",
		Content: "192.0.2.10", Owner: "other",
	})

	plan, err := BuildPlan(PlanInput{
		Live: live, Desired: desired, ManagedOwners: []string{"tracera"},
	})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}

	for _, conflict := range plan.Conflicts {
		if conflict.ResourceType == "dns" && conflict.Key == "id:dns-new" {
			if conflict.LiveOwner != "" || conflict.DesiredOwner != "other" {
				t.Fatalf("unexpected owner audit fields %#v", conflict)
			}
			return
		}
	}
	t.Fatalf("expected desired-only DNS conflict, got %#v", plan.Conflicts)
}

func TestCloudflareAllowsSameNameAndTypeWithDistinctRecordIDs(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.DNS = append(snapshot.DNS,
		DNSRecord{
			ID: "dns-a-1", Name: "multi.example.com", Type: "A",
			Content: "192.0.2.11", Owner: "tracera",
		},
		DNSRecord{
			ID: "dns-a-2", Name: "multi.example.com", Type: "A",
			Content: "192.0.2.12", Owner: "tracera",
		},
	)

	if err := ValidateSnapshot(snapshot); err != nil {
		t.Fatalf("same-name/type Cloudflare records with distinct IDs must be valid: %v", err)
	}
}

func TestCloudflareDNSRequiresProviderRecordID(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.DNS[0].ID = ""

	if err := ValidateSnapshot(snapshot); err == nil {
		t.Fatal("expected Cloudflare DNS record without provider ID to be rejected")
	}
}

func TestTunnelReplacementIsAlwaysConflict(t *testing.T) {
	live := baseSnapshot()
	desired := baseSnapshot()
	desired.Tunnel.ID = "different-tunnel"

	plan, err := BuildPlan(PlanInput{
		Live: live, Desired: desired, ManagedOwners: []string{"tracera"},
	})
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Safe() {
		t.Fatal("tunnel replacement must not be planned implicitly")
	}
}

func TestSecretReferenceContainsNoCredentialValueField(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.AccessApps[0].MachineSecret = &SecretRef{
		Store: "github-actions",
		Key:   "CF_ACCESS_CLIENT_SECRET",
	}

	if err := ValidateSnapshot(snapshot); err != nil {
		t.Fatalf("ValidateSnapshot: %v", err)
	}
	if snapshot.AccessApps[0].MachineSecret.Key != "CF_ACCESS_CLIENT_SECRET" {
		t.Fatal("secret reference changed unexpectedly")
	}
}

func TestDuplicateRouteKeyRejected(t *testing.T) {
	snapshot := baseSnapshot()
	snapshot.Routes = append(
		[]Route{snapshot.Routes[0], snapshot.Routes[0]},
		snapshot.Routes[1:]...,
	)

	if err := ValidateSnapshot(snapshot); err == nil {
		t.Fatal("expected duplicate route key to be rejected")
	}
}
