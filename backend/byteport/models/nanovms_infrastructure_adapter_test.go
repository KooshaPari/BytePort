package models

import (
	"context"
	"testing"
)

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
	sandbox := NanoVMSSandbox{ID: "sandbox-1", Name: req.Name, Status: "running"}
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
		TargetID:  "local-nanovms",
		Provider:  "nanovms",
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
		PlannedResourceAction{DesiredResourceID: "service", Action: ReconcileCreate},
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
		Artifact: &ArtifactReference{Digest: "sha256:abc"},
	}
	result, err := adapter.Apply(
		context.Background(),
		PlannedResourceAction{DesiredResourceID: "service", Action: ReconcileCreate},
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
			DesiredResourceID: "service", RealizedResourceID: "r", Action: ReconcileDelete,
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
