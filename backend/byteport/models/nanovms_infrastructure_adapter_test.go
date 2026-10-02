package models

import (
	"context"
	"errors"
	"testing"
)

type fakeArtifactResolver struct {
	artifacts map[BuildArtifactID]BuildArtifact
}

func (f fakeArtifactResolver) ResolveBuildArtifact(
	_ context.Context,
	id BuildArtifactID,
) (BuildArtifact, bool, error) {
	artifact, ok := f.artifacts[id]
	return artifact, ok, nil
}

type fakeNanoVMSTransport struct {
	deploys []NanoVMSDeployRequest
	stops   []string
	seen    map[string]NanoVMSSandbox
}

func (f *fakeNanoVMSTransport) Deploy(
	_ context.Context,
	req NanoVMSDeployRequest,
) (NanoVMSSandbox, error) {
	f.deploys = append(f.deploys, req)
	sandbox := NanoVMSSandbox{
		ID: "sandbox-1", Name: req.Name, Status: "running", ConfigDigest: req.ConfigDigest,
	}
	if f.seen == nil {
		f.seen = map[string]NanoVMSSandbox{}
	}
	f.seen[sandbox.ID] = sandbox
	return sandbox, nil
}

func (f *fakeNanoVMSTransport) Stop(_ context.Context, id string) error {
	f.stops = append(f.stops, id)
	return nil
}

func (f *fakeNanoVMSTransport) Observe(
	_ context.Context,
	id string,
) (NanoVMSSandbox, bool, error) {
	value, ok := f.seen[id]
	return value, ok, nil
}

func nanoAdapter(transport NanoVMSTransport) NanoVMSInfrastructureAdapter {
	return NanoVMSInfrastructureAdapter{
		Transport: transport,
		Artifacts: fakeArtifactResolver{artifacts: map[BuildArtifactID]BuildArtifact{
			"artifact-1": {
				ID: "artifact-1",
				ImmutableRef: "sha256:abc",
				MediaKind: "oci-image",
			},
		}},
		TargetID: "local-nanovms",
		Provider: "nanovms",
	}
}

func TestNanoVMSAdapterCreateRequiresImmutableArtifact(t *testing.T) {
	transport := &fakeNanoVMSTransport{}
	adapter := nanoAdapter(transport)
	desired := DesiredResource{
		ID: "service", Kind: DesiredResourceService,
		ConfigDigest: "cfg", Target: "local-nanovms", Lifecycle: LifecycleManage,
	}
	_, err := adapter.Apply(
		context.Background(),
		PlannedResourceAction{OperationID: "op-create", DesiredResourceID: "service", Action: ReconcileCreate},
		&desired,
		nil,
	)
	if err == nil {
		t.Fatal("CREATE accepted without immutable artifact")
	}
	if len(transport.deploys) != 0 {
		t.Fatalf("artifact rejection happened after %d provider mutations", len(transport.deploys))
	}
}

func TestNanoVMSAdapterCreateUsesArtifactDigestAndPreservesProviderIdentity(t *testing.T) {
	transport := &fakeNanoVMSTransport{}
	adapter := nanoAdapter(transport)
	desired := DesiredResource{
		ID: "service", Kind: DesiredResourceService,
		ConfigDigest: "cfg", Target: "local-nanovms", Lifecycle: LifecycleManage,
		Artifact: func() *BuildArtifactID { id := BuildArtifactID("artifact-1"); return &id }(),
	}
	result, err := adapter.Apply(
		context.Background(),
		PlannedResourceAction{OperationID: "op-create", DesiredResourceID: "service", Action: ReconcileCreate},
		&desired,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(transport.deploys) != 1 || transport.deploys[0].Image != "sha256:abc" {
		t.Fatalf("deploys=%v", transport.deploys)
	}
	if result.Realized == nil ||
		result.Realized.ExternalID != "sandbox-1" ||
		result.Realized.TargetID != "local-nanovms" ||
		result.Realized.Provider != "nanovms" {
		t.Fatalf("realized=%v", result.Realized)
	}
}

func TestNanoVMSAdapterDeleteRequiresExactProviderIdentityBeforeStop(t *testing.T) {
	transport := &fakeNanoVMSTransport{}
	adapter := nanoAdapter(transport)
	wrong := RealizedResource{
		ID: "r", DesiredResourceID: "service",
		TargetID: "local-nanovms", Provider: "other", ExternalID: "sandbox-1",
	}
	_, err := adapter.Apply(
		context.Background(),
		PlannedResourceAction{
			OperationID: "op-delete", DesiredResourceID: "service",
			RealizedResourceID: "r", Action: ReconcileDelete,
		},
		nil,
		&wrong,
	)
	if err == nil {
		t.Fatal("wrong-provider delete was accepted")
	}
	if len(transport.stops) != 0 {
		t.Fatalf("wrong-provider delete caused %d stops", len(transport.stops))
	}
}

func TestNanoVMSAdapterDoesNotClaimUpdateOrReplace(t *testing.T) {
	transport := &fakeNanoVMSTransport{}
	adapter := nanoAdapter(transport)
	caps, err := adapter.Capabilities(context.Background(), "local-nanovms")
	if err != nil {
		t.Fatal(err)
	}
	if caps.SupportsUpdate || caps.SupportsReplace {
		t.Fatalf("unsupported NanoVMS mutation capability was claimed: %+v", caps)
	}
}

func TestNanoVMSAdapterMissingObservationIsExplicitlyNotFresh(t *testing.T) {
	transport := &fakeNanoVMSTransport{seen: map[string]NanoVMSSandbox{}}
	adapter := nanoAdapter(transport)
	observation, err := adapter.Observe(context.Background(), RealizedResource{
		ID: "r", DesiredResourceID: "service", TargetID: "local-nanovms",
		Provider: "nanovms", ExternalID: "missing",
	})
	if err != nil {
		t.Fatal(err)
	}
	if observation.Fresh || observation.State != "unknown" {
		t.Fatalf("missing provider state was treated as fresh: %+v", observation)
	}
}


type delayedNanoVMSTransport struct {
	deploys       int
	observations  int
	sandbox       NanoVMSSandbox
	visibleAfter  int
}

func (d *delayedNanoVMSTransport) Deploy(
	_ context.Context,
	_ NanoVMSDeployRequest,
) (NanoVMSSandbox, error) {
	d.deploys++
	if d.sandbox.ID == "" {
		d.sandbox = NanoVMSSandbox{
			ID: "sandbox-delayed", Name: "service", Status: "running", ConfigDigest: "cfg",
		}
	}
	return d.sandbox, nil
}

func (d *delayedNanoVMSTransport) Stop(_ context.Context, _ string) error { return nil }

func (d *delayedNanoVMSTransport) Observe(
	_ context.Context,
	id string,
) (NanoVMSSandbox, bool, error) {
	d.observations++
	if id != d.sandbox.ID || d.observations <= d.visibleAfter {
		return NanoVMSSandbox{}, false, nil
	}
	return d.sandbox, true, nil
}

func TestNanoVMSLostResponseReconciliationObservesBeforeAnySecondCreate(t *testing.T) {
	transport := &delayedNanoVMSTransport{visibleAfter: 1}
	adapter := nanoAdapter(transport)
	artifactID := BuildArtifactID("artifact-1")
	desired := DesiredResource{
		ID: "service", Kind: DesiredResourceService,
		ConfigDigest: "cfg", Target: "local-nanovms",
		Lifecycle: LifecycleManage, Artifact: &artifactID,
	}

	// Provider mutation happened, but imagine the caller lost the response
	// after persisting the external sandbox identity into its operation journal.
	create, err := adapter.Apply(
		context.Background(),
		PlannedResourceAction{OperationID: "op-create", DesiredResourceID: "service", Action: ReconcileCreate},
		&desired,
		nil,
	)
	if err != nil { t.Fatal(err) }
	if create.Realized == nil { t.Fatal("create returned no realized identity") }
	if transport.deploys != 1 { t.Fatalf("deploys=%d", transport.deploys) }

	// First provider lookup is visibility-uncertain. It must remain non-fresh
	// rather than authorizing another CREATE.
	first, err := adapter.Observe(context.Background(), *create.Realized)
	if err != nil { t.Fatal(err) }
	if first.Fresh || first.State != "unknown" {
		t.Fatalf("first observation=%+v", first)
	}
	if transport.deploys != 1 {
		t.Fatalf("observation caused duplicate deploy: %d", transport.deploys)
	}

	// Once provider visibility converges, observation resolves the same exact
	// sandbox identity; no second mutation is needed.
	second, err := adapter.Observe(context.Background(), *create.Realized)
	if err != nil { t.Fatal(err) }
	if !second.Fresh || second.State != "running" {
		t.Fatalf("second observation=%+v", second)
	}
	if transport.deploys != 1 {
		t.Fatalf("delayed visibility caused duplicate deploy: %d", transport.deploys)
	}
}


type ambiguousNanoVMSTransport struct {
	deploys int
}

func (a *ambiguousNanoVMSTransport) Deploy(
	_ context.Context,
	_ NanoVMSDeployRequest,
) (NanoVMSSandbox, error) {
	a.deploys++
	return NanoVMSSandbox{}, &NanoVMSOutcomeUnknownError{
		Operation: "deploy",
		Cause: errors.New("connection reset after request body sent"),
	}
}

func (a *ambiguousNanoVMSTransport) Stop(_ context.Context, _ string) error {
	return nil
}

func (a *ambiguousNanoVMSTransport) Observe(
	_ context.Context,
	_ string,
) (NanoVMSSandbox, bool, error) {
	return NanoVMSSandbox{}, false, nil
}

func TestNanoVMSAmbiguousCreateReturnsUnknownInsteadOfRetryableGenericError(t *testing.T) {
	transport := &ambiguousNanoVMSTransport{}
	adapter := nanoAdapter(transport)
	artifactID := BuildArtifactID("artifact-1")
	desired := DesiredResource{
		ID: "service", Kind: DesiredResourceService,
		ConfigDigest: "cfg", Target: "local-nanovms",
		Lifecycle: LifecycleManage, Artifact: &artifactID,
	}
	result, err := adapter.Apply(
		context.Background(),
		PlannedResourceAction{OperationID: "op-create", DesiredResourceID: "service", Action: ReconcileCreate},
		&desired,
		nil,
	)
	if err != nil {
		t.Fatalf("ambiguous provider outcome escaped as generic retryable error: %v", err)
	}
	if result.Outcome != InfrastructureApplyUnknown {
		t.Fatalf("outcome=%q want UNKNOWN", result.Outcome)
	}
	if result.Realized != nil {
		t.Fatalf("ambiguous create fabricated realized identity: %+v", result.Realized)
	}
	if result.ExternalOperation == nil ||
		result.ExternalOperation.Provider != "nanovms" ||
		result.ExternalOperation.TargetID != "local-nanovms" {
		t.Fatalf("missing reconciliation identity: %+v", result.ExternalOperation)
	}
	if transport.deploys != 1 {
		t.Fatalf("deploys=%d want 1", transport.deploys)
	}
}
