// FR: BP-WP-B08. In-memory lifecycle composition, not a durable process restart
// or a live NanoVMS provider. Stop and resource deletion are distinct operations.
package models

import (
	"context"
	"fmt"
	"testing"
)

type lifecycleRefResolver struct{ commit string }

func (r lifecycleRefResolver) ResolveRef(context.Context, string, string) (string, error) {
	return r.commit, nil
}

type lifecycleArtifactResolver struct {
	artifacts map[BuildArtifactID]BuildArtifact
}

func (r lifecycleArtifactResolver) ResolveBuildArtifact(_ context.Context, id BuildArtifactID) (BuildArtifact, bool, error) {
	artifact, ok := r.artifacts[id]
	return artifact, ok, nil
}

type lifecycleNanoTransport struct {
	deploys          []NanoVMSDeployRequest
	stops            []string
	deletes          []string
	deleteOperations []string
	seen             map[string]NanoVMSSandbox
}

func (t *lifecycleNanoTransport) Deploy(_ context.Context, req NanoVMSDeployRequest) (NanoVMSSandbox, error) {
	t.deploys = append(t.deploys, req)
	sandbox := NanoVMSSandbox{ID: "sandbox-e2e", Name: req.Name, Status: "running", ConfigDigest: req.ConfigDigest}
	if t.seen == nil {
		t.seen = map[string]NanoVMSSandbox{}
	}
	t.seen[sandbox.ID] = sandbox
	return sandbox, nil
}
func (t *lifecycleNanoTransport) Stop(_ context.Context, id string) error {
	t.stops = append(t.stops, id)
	value, ok := t.seen[id]
	if !ok {
		return fmt.Errorf("unknown sandbox %q", id)
	}
	value.Status = "stopped"
	t.seen[id] = value
	return nil
}

// Delete exists only on this explicit fixture provider, not on the real HTTP transport.
func (t *lifecycleNanoTransport) Delete(_ context.Context, id, operation string) error {
	if _, ok := t.seen[id]; !ok || operation == "" {
		return fmt.Errorf("invalid fixture deletion identity")
	}
	t.deletes = append(t.deletes, id)
	t.deleteOperations = append(t.deleteOperations, operation)
	delete(t.seen, id)
	return nil
}
func (t *lifecycleNanoTransport) Observe(_ context.Context, id string) (NanoVMSSandbox, bool, error) {
	sandbox, ok := t.seen[id]
	return sandbox, ok, nil
}

func TestDisposableLifecycleSnapshotToExactDestroyWithoutDuplicateCreate(t *testing.T) {
	ctx := context.Background()
	snapshot, err := ResolveGitSourceSnapshot(ctx, lifecycleRefResolver{commit: "1111111111111111111111111111111111111111"}, "repo", "main")
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte("NAME: disposable\nSERVICES:\n  - NAME: api\n    PATH: .\n    PORT: 8080\n    RUNTIME: container\n")
	revision, manifest, err := ParseRecoveryManifest(raw, snapshot.ID, "byteport.yaml")
	if err != nil {
		t.Fatal(err)
	}
	graph, err := RecoveryManifestToDesiredGraph(revision, manifest, "local-nanovms")
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Resources) != 1 {
		t.Fatalf("resources=%v", graph.Resources)
	}
	artifactID := BuildArtifactID("artifact-e2e")
	graph.Resources[0].Artifact = &artifactID
	artifact := BuildArtifact{ID: artifactID, ImmutableRef: "sha256:e2e", MediaKind: "oci-image", SourceSnapshot: snapshot.ID, ManifestRevision: revision.ID, BuildOperation: "build-e2e", Engine: "fixture", EngineVersion: "1"}
	artifacts := lifecycleArtifactResolver{artifacts: map[BuildArtifactID]BuildArtifact{artifactID: artifact}}
	if err := ValidateDesiredGraphArtifactLineage(ctx, graph, artifacts); err != nil {
		t.Fatal(err)
	}
	transport := &lifecycleNanoTransport{}
	adapter := NanoVMSInfrastructureAdapter{Transport: transport, Artifacts: artifacts, TargetID: "local-nanovms", Provider: "nanovms"}
	root := RuntimeOperationID("runtime-create-e2e")
	created, err := ReconcileOnceWithTargetAdapterOperation(ctx, root, graph, nil, map[string]RealizedResource{}, ReconciliationPolicy{}, nil, "local-nanovms", adapter)
	if err != nil {
		t.Fatal(err)
	}
	expectedCreate := deriveActionOperationID(root, PlannedResourceAction{DesiredResourceID: graph.Resources[0].ID, Action: ReconcileCreate})
	if len(created.Executions) != 1 {
		t.Fatalf("create executions=%+v", created.Executions)
	}
	execution := created.Executions[0]
	if execution.Action.Action != ReconcileCreate || execution.Action.OperationID != expectedCreate || !execution.Applied || execution.ApplyResult == nil || execution.ApplyResult.Outcome != InfrastructureApplyRealized || execution.ApplyResult.Realized == nil {
		t.Fatalf("create=%+v", execution)
	}
	if len(transport.deploys) != 1 || transport.deploys[0].OperationID != string(expectedCreate) || transport.deploys[0].Image != "sha256:e2e" {
		t.Fatalf("deploys=%+v", transport.deploys)
	}
	realized := *execution.ApplyResult.Realized
	observation, err := adapter.Observe(ctx, realized)
	if err != nil {
		t.Fatal(err)
	}
	if !observation.Fresh || observation.RealizedResourceID != realized.ID || observation.ConfigDigest != graph.Resources[0].ConfigDigest {
		t.Fatalf("observation=%+v", observation)
	}
	observed := []ObservedResourceState{{DesiredResourceID: graph.Resources[0].ID, RealizedResourceID: realized.ID, TargetID: observation.TargetID, Provider: observation.Provider, ConfigDigest: observation.ConfigDigest, Fresh: observation.Fresh, Lifecycle: graph.Resources[0].Lifecycle}}
	realizedByID := map[string]RealizedResource{realized.ID: realized}
	// Reconstruct the adapter from in-memory identities. This does not test DB persistence.
	resumed := NanoVMSInfrastructureAdapter{Transport: transport, Artifacts: artifacts, TargetID: "local-nanovms", Provider: "nanovms"}
	replay, err := ReconcileOnceWithTargetAdapterOperation(ctx, "runtime-reconcile-e2e", graph, observed, realizedByID, ReconciliationPolicy{}, nil, "local-nanovms", resumed)
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.Plan.Actions) != 1 || replay.Plan.Actions[0].Action != ReconcileNoop || len(transport.deploys) != 1 {
		t.Fatalf("replay=%+v deploys=%d", replay, len(transport.deploys))
	}
	observed[0].Lifecycle = LifecycleDestroyOnExplicitIntent
	intent := ReconciliationPolicy{DestructionIntents: []DestructionIntent{{ID: "destroy-e2e", DesiredResourceID: graph.Resources[0].ID, RealizedResourceID: realized.ID, AuthorizedBy: "disposable-fixture", Reason: "end disposable lifecycle"}}}
	deleteRoot := RuntimeOperationID("runtime-delete-e2e")
	deleted, err := ReconcileOnceWithTargetAdapterOperation(ctx, deleteRoot, DesiredResourceGraph{ID: graph.ID + ":removed", Manifest: graph.Manifest}, observed, realizedByID, intent, nil, "local-nanovms", resumed)
	if err != nil {
		t.Fatal(err)
	}
	expectedDelete := deriveActionOperationID(deleteRoot, PlannedResourceAction{DesiredResourceID: graph.Resources[0].ID, RealizedResourceID: realized.ID, Action: ReconcileDelete})
	if len(deleted.Executions) != 1 || deleted.Executions[0].Action.OperationID != expectedDelete || !deleted.Executions[0].Applied || deleted.Executions[0].ApplyResult == nil || deleted.Executions[0].ApplyResult.Observation == nil || deleted.Executions[0].ApplyResult.Observation.State != "absent" {
		t.Fatalf("delete=%+v", deleted)
	}
	if len(transport.stops) != 0 || len(transport.deletes) != 1 || transport.deletes[0] != realized.ExternalID || transport.deleteOperations[0] != string(expectedDelete) {
		t.Fatalf("stops=%v deletes=%v ops=%v", transport.stops, transport.deletes, transport.deleteOperations)
	}
	if _, found, err := transport.Observe(ctx, realized.ExternalID); err != nil || found {
		t.Fatalf("removed resource still exists: %v %v", found, err)
	}
	if len(transport.deploys) != 1 {
		t.Fatal("lifecycle duplicated CREATE")
	}
}
