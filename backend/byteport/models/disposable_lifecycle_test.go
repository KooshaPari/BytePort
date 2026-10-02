package models

import (
	"context"
	"testing"
)

type lifecycleRefResolver struct {
	commit string
}

func (r lifecycleRefResolver) ResolveRef(
	_ context.Context,
	_ string,
	_ string,
) (string, error) {
	return r.commit, nil
}

type lifecycleArtifactResolver struct {
	artifacts map[BuildArtifactID]BuildArtifact
}

func (r lifecycleArtifactResolver) ResolveBuildArtifact(
	_ context.Context,
	id BuildArtifactID,
) (BuildArtifact, bool, error) {
	artifact, ok := r.artifacts[id]
	return artifact, ok, nil
}

type lifecycleNanoTransport struct {
	deploys []NanoVMSDeployRequest
	stops   []string
	seen    map[string]NanoVMSSandbox
}

func (t *lifecycleNanoTransport) Deploy(
	_ context.Context,
	req NanoVMSDeployRequest,
) (NanoVMSSandbox, error) {
	t.deploys = append(t.deploys, req)
	sandbox := NanoVMSSandbox{
		ID:           "sandbox-e2e",
		Name:         req.Name,
		Status:       "running",
		ConfigDigest: req.ConfigDigest,
	}
	if t.seen == nil {
		t.seen = map[string]NanoVMSSandbox{}
	}
	t.seen[sandbox.ID] = sandbox
	return sandbox, nil
}

func (t *lifecycleNanoTransport) Stop(_ context.Context, id string) error {
	t.stops = append(t.stops, id)
	delete(t.seen, id)
	return nil
}

func (t *lifecycleNanoTransport) Observe(
	_ context.Context,
	id string,
) (NanoVMSSandbox, bool, error) {
	sandbox, ok := t.seen[id]
	return sandbox, ok, nil
}

func TestDisposableLifecycleSnapshotToExactDestroyWithoutDuplicateCreate(t *testing.T) {
	ctx := context.Background()

	snapshot, err := ResolveGitSourceSnapshot(
		ctx,
		lifecycleRefResolver{
			commit: "1111111111111111111111111111111111111111",
		},
		"repo",
		"main",
	)
	if err != nil {
		t.Fatal(err)
	}

	rawManifest := []byte(`NAME: disposable
SERVICES:
  - NAME: api
    PATH: .
    PORT: 8080
    RUNTIME: container
`)
	revision, manifest, err := ParseRecoveryManifest(
		rawManifest,
		snapshot.ID,
		"byteport.yaml",
	)
	if err != nil {
		t.Fatal(err)
	}

	graph, err := RecoveryManifestToDesiredGraph(
		revision,
		manifest,
		"local-nanovms",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Resources) != 1 {
		t.Fatalf("resources=%v", graph.Resources)
	}

	artifactID := BuildArtifactID("artifact-e2e")
	graph.Resources[0].Artifact = &artifactID
	artifact := BuildArtifact{
		ID:               artifactID,
		ImmutableRef:     "sha256:e2e",
		MediaKind:        "oci-image",
		SourceSnapshot:   snapshot.ID,
		ManifestRevision: revision.ID,
		BuildOperation:   "build-e2e",
		Engine:           "fixture",
		EngineVersion:    "1",
	}
	artifacts := lifecycleArtifactResolver{
		artifacts: map[BuildArtifactID]BuildArtifact{
			artifactID: artifact,
		},
	}

	if err := ValidateDesiredGraphArtifactLineage(ctx, graph, artifacts); err != nil {
		t.Fatal(err)
	}

	transport := &lifecycleNanoTransport{}
	adapter := NanoVMSInfrastructureAdapter{
		Transport: transport,
		Artifacts: artifacts,
		TargetID:  "local-nanovms",
		Provider:  "nanovms",
	}

	createRoot := RuntimeOperationID("runtime-create-e2e")
	createReceipt, err := ReconcileOnceWithTargetAdapterOperation(
		ctx,
		createRoot,
		graph,
		nil,
		map[string]RealizedResource{},
		ReconciliationPolicy{},
		nil,
		"local-nanovms",
		adapter,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(createReceipt.Executions) != 1 {
		t.Fatalf("create executions=%v", createReceipt.Executions)
	}
	createExecution := createReceipt.Executions[0]
	expectedCreateOperation := deriveActionOperationID(
		createRoot,
		PlannedResourceAction{
			DesiredResourceID: graph.Resources[0].ID,
			Action:            ReconcileCreate,
		},
	)
	if createExecution.Action.Action != ReconcileCreate ||
		createExecution.Action.OperationID != expectedCreateOperation ||
		createExecution.ApplyResult == nil ||
		createExecution.ApplyResult.Outcome != InfrastructureApplyRealized ||
		createExecution.ApplyResult.Realized == nil {
		t.Fatalf("create execution=%+v", createExecution)
	}
	if len(transport.deploys) != 1 ||
		transport.deploys[0].OperationID != string(expectedCreateOperation) ||
		transport.deploys[0].Image != "sha256:e2e" {
		t.Fatalf("deploy receipt=%+v", transport.deploys)
	}

	realized := *createExecution.ApplyResult.Realized
	observation, err := adapter.Observe(ctx, realized)
	if err != nil {
		t.Fatal(err)
	}
	if !observation.Fresh ||
		observation.RealizedResourceID != realized.ID ||
		observation.ConfigDigest != graph.Resources[0].ConfigDigest {
		t.Fatalf("observation=%+v", observation)
	}

	// Reconstruct only persisted identities and provider observation, as a
	// restart would. The second reconciliation must NOOP, never CREATE again.
	observed := []ObservedResourceState{{
		DesiredResourceID:  graph.Resources[0].ID,
		RealizedResourceID: realized.ID,
		TargetID:           observation.TargetID,
		Provider:           observation.Provider,
		ConfigDigest:       observation.ConfigDigest,
		Fresh:              observation.Fresh,
		Lifecycle:          graph.Resources[0].Lifecycle,
	}}
	realizedByID := map[string]RealizedResource{realized.ID: realized}

	restartedAdapter := NanoVMSInfrastructureAdapter{
		Transport: transport,
		Artifacts: artifacts,
		TargetID:  "local-nanovms",
		Provider:  "nanovms",
	}
	restartReceipt, err := ReconcileOnceWithTargetAdapterOperation(
		ctx,
		"runtime-reconcile-e2e",
		graph,
		observed,
		realizedByID,
		ReconciliationPolicy{},
		nil,
		"local-nanovms",
		restartedAdapter,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(restartReceipt.Plan.Actions) != 1 ||
		restartReceipt.Plan.Actions[0].Action != ReconcileNoop {
		t.Fatalf("restart plan=%+v", restartReceipt.Plan.Actions)
	}
	if len(transport.deploys) != 1 {
		t.Fatalf("restart duplicated CREATE: deploys=%d", len(transport.deploys))
	}

	// Removing the resource from desired state does not authorize destruction
	// by itself. Exact lifecycle state + exact intent + exact realized identity
	// are required for the one destructive mutation.
	deleteObserved := []ObservedResourceState{{
		DesiredResourceID:  graph.Resources[0].ID,
		RealizedResourceID: realized.ID,
		TargetID:           realized.TargetID,
		Provider:           realized.Provider,
		ConfigDigest:       graph.Resources[0].ConfigDigest,
		Fresh:              true,
		Lifecycle:          LifecycleDestroyOnExplicitIntent,
	}}
	deletePolicy := ReconciliationPolicy{
		DestructionIntents: []DestructionIntent{{
			ID:                 "destroy-e2e",
			DesiredResourceID:  graph.Resources[0].ID,
			RealizedResourceID: realized.ID,
			AuthorizedBy:       "disposable-fixture",
			Reason:             "end disposable lifecycle",
		}},
	}

	deleteRoot := RuntimeOperationID("runtime-delete-e2e")
	deleteReceipt, err := ReconcileOnceWithTargetAdapterOperation(
		ctx,
		deleteRoot,
		DesiredResourceGraph{
			ID:       graph.ID + ":removed",
			Manifest: graph.Manifest,
		},
		deleteObserved,
		realizedByID,
		deletePolicy,
		nil,
		"local-nanovms",
		restartedAdapter,
	)
	if err != nil {
		t.Fatal(err)
	}
	expectedDeleteOperation := deriveActionOperationID(
		deleteRoot,
		PlannedResourceAction{
			DesiredResourceID:  graph.Resources[0].ID,
			RealizedResourceID: realized.ID,
			Action:             ReconcileDelete,
		},
	)
	if len(deleteReceipt.Plan.Actions) != 1 ||
		deleteReceipt.Plan.Actions[0].Action != ReconcileDelete ||
		deleteReceipt.Plan.Actions[0].OperationID != expectedDeleteOperation {
		t.Fatalf("delete plan=%+v", deleteReceipt.Plan.Actions)
	}
	if len(transport.stops) != 1 || transport.stops[0] != realized.ExternalID {
		t.Fatalf("stop receipt=%v", transport.stops)
	}
	if len(transport.deploys) != 1 {
		t.Fatalf("lifecycle performed duplicate CREATE: deploys=%d", len(transport.deploys))
	}
}
