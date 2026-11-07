// FR: BP-WP-B08 destructive capability and identity are not inferred from Stop.
package models

import (
	"context"
	"errors"
	"testing"
)

type deletionSafetyTransport struct {
	stops, creates, observations int
	exists                       bool
	observeErr                   error
	createID                     string
}

func (f *deletionSafetyTransport) Deploy(_ context.Context, req NanoVMSDeployRequest) (NanoVMSSandbox, error) {
	f.creates++
	f.exists = true
	return NanoVMSSandbox{ID: f.createID, Name: req.Name, Status: "running"}, nil
}
func (f *deletionSafetyTransport) Stop(context.Context, string) error { f.stops++; return nil }
func (f *deletionSafetyTransport) Observe(_ context.Context, id string) (NanoVMSSandbox, bool, error) {
	f.observations++
	if f.observeErr != nil {
		return NanoVMSSandbox{}, false, f.observeErr
	}
	return NanoVMSSandbox{ID: id, Status: "stopped"}, f.exists, nil
}

type deletionSafetyDeleter struct {
	*deletionSafetyTransport
	deletes               int
	keep                  bool
	deleteErr             error
	seenID, seenOperation string
}

func (f *deletionSafetyDeleter) Delete(_ context.Context, id, operation string) error {
	f.deletes++
	f.seenID = id
	f.seenOperation = operation
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.exists = f.keep
	return nil
}

type deletionSafetyArtifacts struct{}

func (deletionSafetyArtifacts) ResolveBuildArtifact(_ context.Context, id BuildArtifactID) (BuildArtifact, bool, error) {
	return BuildArtifact{ID: id, ImmutableRef: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}, true, nil
}
func deletionSafetyAdapter(transport NanoVMSTransport) NanoVMSInfrastructureAdapter {
	return NanoVMSInfrastructureAdapter{Transport: transport, Artifacts: deletionSafetyArtifacts{}, TargetID: "target", Provider: "nanovms"}
}
func deletionSafetyResource() RealizedResource {
	return RealizedResource{ID: "real", DesiredResourceID: "service", TargetID: "target", Provider: "nanovms", ExternalID: "sandbox"}
}
func deletionSafetyAction() PlannedResourceAction {
	return PlannedResourceAction{OperationID: "delete-op", DesiredResourceID: "service", RealizedResourceID: "real", Action: ReconcileDelete}
}
func deletionSafetyDesired() DesiredResource {
	id := BuildArtifactID("artifact")
	return DesiredResource{ID: "service", Kind: DesiredResourceService, ConfigDigest: "config", Artifact: &id, Target: "target", Lifecycle: LifecycleManage}
}

func TestDeletionSafetyStopOnlyDoesNotAdvertiseDelete(t *testing.T) {
	f := &deletionSafetyTransport{exists: true}
	a := deletionSafetyAdapter(f)
	c, err := a.Capabilities(context.Background(), "target")
	if err != nil {
		t.Fatal(err)
	}
	if c.SupportsDelete {
		t.Fatal("Stop-only transport claimed destructive capability")
	}
}
func TestDeletionSafetyStopOnlyCannotExecuteDelete(t *testing.T) {
	f := &deletionSafetyTransport{exists: true}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	result, err := a.Apply(context.Background(), deletionSafetyAction(), nil, &r)
	if err == nil {
		t.Fatalf("DELETE accepted on Stop-only transport: %+v", result)
	}
	if f.stops != 0 {
		t.Fatalf("DELETE silently degraded to %d Stop calls", f.stops)
	}
	if !f.exists {
		t.Fatal("unsupported deletion changed resource")
	}
}
func TestDeletionSafetyConfirmedDeletionCallsDeleteNotStop(t *testing.T) {
	f := &deletionSafetyDeleter{deletionSafetyTransport: &deletionSafetyTransport{exists: true}}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	result, err := a.Apply(context.Background(), deletionSafetyAction(), nil, &r)
	if err != nil {
		t.Fatal(err)
	}
	if f.deletes != 1 || f.stops != 0 || f.exists {
		t.Fatalf("delete=%d stop=%d exists=%v", f.deletes, f.stops, f.exists)
	}
	if f.seenID != "sandbox" || f.seenOperation != "delete-op" {
		t.Fatalf("lost mutation identity: %+v", f)
	}
	if result.Outcome != InfrastructureApplyRealized || result.Realized != nil || result.Observation == nil || result.Observation.State != "absent" {
		t.Fatalf("unproven deletion receipt: %+v", result)
	}
}
func TestDeletionSafetyAcknowledgementWithRetainedResourceRemainsUnknown(t *testing.T) {
	f := &deletionSafetyDeleter{deletionSafetyTransport: &deletionSafetyTransport{exists: true}, keep: true}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	result, err := a.Apply(context.Background(), deletionSafetyAction(), nil, &r)
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != InfrastructureApplyUnknown || result.Realized != nil {
		t.Fatalf("retained resource became deleted: %+v", result)
	}
	if f.deletes != 1 || f.stops != 0 || !f.exists {
		t.Fatalf("incorrect destructive call: %+v", f)
	}
}
func TestDeletionSafetyObservationFailureDoesNotEstablishDeletion(t *testing.T) {
	f := &deletionSafetyDeleter{deletionSafetyTransport: &deletionSafetyTransport{exists: true, observeErr: errors.New("observer unavailable")}}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	result, err := a.Apply(context.Background(), deletionSafetyAction(), nil, &r)
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != InfrastructureApplyUnknown || result.ExternalOperation == nil || result.ExternalOperation.ExternalID != "sandbox" {
		t.Fatalf("lost uncertainty/resource identity: %+v", result)
	}
}
func TestDeletionSafetyAmbiguousDeleteRetainsIdentityWithoutAnotherMutation(t *testing.T) {
	f := &deletionSafetyDeleter{deletionSafetyTransport: &deletionSafetyTransport{exists: true}, deleteErr: &NanoVMSOutcomeUnknownError{Operation: "delete-op", Cause: errors.New("lost response")}}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	result, err := a.Apply(context.Background(), deletionSafetyAction(), nil, &r)
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != InfrastructureApplyUnknown || result.ExternalOperation == nil || result.ExternalOperation.ExternalID != "sandbox" || f.deletes != 1 || f.stops != 0 {
		t.Fatalf("ambiguity collapsed: result=%+v transport=%+v", result, f)
	}
}
func TestDeletionSafetyWrongRealizedActionIdentityCannotMutate(t *testing.T) {
	f := &deletionSafetyDeleter{deletionSafetyTransport: &deletionSafetyTransport{exists: true}}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	action := deletionSafetyAction()
	action.RealizedResourceID = "other"
	_, err := a.Apply(context.Background(), action, nil, &r)
	if err == nil || f.deletes+f.stops != 0 {
		t.Fatalf("wrong realized identity reached mutation: %v %+v", err, f)
	}
}
func TestDeletionSafetyWrongDesiredActionIdentityCannotMutate(t *testing.T) {
	f := &deletionSafetyDeleter{deletionSafetyTransport: &deletionSafetyTransport{exists: true}}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	action := deletionSafetyAction()
	action.DesiredResourceID = "other"
	_, err := a.Apply(context.Background(), action, nil, &r)
	if err == nil || f.deletes+f.stops != 0 {
		t.Fatalf("wrong desired identity reached mutation: %v %+v", err, f)
	}
}
func TestDeletionSafetyWrongProviderCannotMutate(t *testing.T) {
	f := &deletionSafetyDeleter{deletionSafetyTransport: &deletionSafetyTransport{exists: true}}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	r.Provider = "other"
	_, err := a.Apply(context.Background(), deletionSafetyAction(), nil, &r)
	if err == nil || f.deletes+f.stops != 0 {
		t.Fatal("wrong provider accepted")
	}
}
func TestDeletionSafetyMissingProviderObservationIsNotAttributed(t *testing.T) {
	f := &deletionSafetyTransport{exists: true}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	r.Provider = ""
	_, err := a.Observe(context.Background(), r)
	if err == nil || f.observations != 0 {
		t.Fatal("unknown provider identity was attributed to configured provider")
	}
}
func TestDeletionSafetyCreateActionCannotSelectDifferentDesiredResource(t *testing.T) {
	f := &deletionSafetyTransport{createID: "sandbox"}
	a := deletionSafetyAdapter(f)
	d := deletionSafetyDesired()
	_, err := a.Apply(context.Background(), PlannedResourceAction{OperationID: "create-op", DesiredResourceID: "other", Action: ReconcileCreate}, &d, nil)
	if err == nil || f.creates != 0 {
		t.Fatal("mismatched create action mutated provider")
	}
}
func TestDeletionSafetyObserveOnlyCreateCannotMutate(t *testing.T) {
	f := &deletionSafetyTransport{createID: "sandbox"}
	a := deletionSafetyAdapter(f)
	d := deletionSafetyDesired()
	d.Lifecycle = LifecycleObserveOnly
	_, err := a.Apply(context.Background(), PlannedResourceAction{OperationID: "create-op", DesiredResourceID: d.ID, Action: ReconcileCreate}, &d, nil)
	if err == nil || f.creates != 0 {
		t.Fatal("observe-only create mutated provider")
	}
}
func TestDeletionSafetyUnknownLifecycleCreateCannotMutate(t *testing.T) {
	f := &deletionSafetyTransport{createID: "sandbox"}
	a := deletionSafetyAdapter(f)
	d := deletionSafetyDesired()
	d.Lifecycle = ""
	_, err := a.Apply(context.Background(), PlannedResourceAction{OperationID: "create-op", DesiredResourceID: d.ID, Action: ReconcileCreate}, &d, nil)
	if err == nil || f.creates != 0 {
		t.Fatal("unknown lifecycle create mutated provider")
	}
}
func TestDeletionSafetyMissingCreateReceiptRemainsUnknown(t *testing.T) {
	f := &deletionSafetyTransport{}
	a := deletionSafetyAdapter(f)
	d := deletionSafetyDesired()
	result, err := a.Apply(context.Background(), PlannedResourceAction{OperationID: "create-op", DesiredResourceID: d.ID, Action: ReconcileCreate}, &d, nil)
	if err != nil || result.Outcome != InfrastructureApplyUnknown || result.Realized != nil || f.creates != 1 {
		t.Fatalf("possible creation lost: %v %+v", err, result)
	}
}
func TestDeletionSafetyBlankOperationNeverMutates(t *testing.T) {
	f := &deletionSafetyTransport{createID: "sandbox"}
	a := deletionSafetyAdapter(f)
	d := deletionSafetyDesired()
	_, err := a.Apply(context.Background(), PlannedResourceAction{OperationID: "create-op", DesiredResourceID: d.ID, Action: ReconcileCreate}, &d, nil)
	if err != nil || f.creates != 1 {
		t.Fatal("positive identity control failed")
	}
	f.creates = 0
	_, err = a.Apply(context.Background(), PlannedResourceAction{OperationID: "  ", DesiredResourceID: d.ID, Action: ReconcileCreate}, &d, nil)
	if err == nil || f.creates != 0 {
		t.Fatal("blank operation reached provider")
	}
}
func TestDeletionSafetyCancelledContextNeverMutates(t *testing.T) {
	f := &deletionSafetyTransport{createID: "sandbox"}
	a := deletionSafetyAdapter(f)
	d := deletionSafetyDesired()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := a.Apply(ctx, PlannedResourceAction{OperationID: "create-op", DesiredResourceID: d.ID, Action: ReconcileCreate}, &d, nil)
	if err == nil || f.creates != 0 {
		t.Fatal("cancelled invocation reached provider")
	}
}
func TestDeletionSafetyKnownCreateStillSucceeds(t *testing.T) {
	f := &deletionSafetyTransport{createID: "sandbox"}
	a := deletionSafetyAdapter(f)
	d := deletionSafetyDesired()
	result, err := a.Apply(context.Background(), PlannedResourceAction{OperationID: "create-op", DesiredResourceID: d.ID, Action: ReconcileCreate}, &d, nil)
	if err != nil || result.Outcome != InfrastructureApplyRealized || result.Realized == nil || result.Realized.ExternalID != "sandbox" || f.creates != 1 {
		t.Fatalf("valid create regressed: %v %+v", err, result)
	}
}
func TestDeletionSafetyReadDoesNotMutate(t *testing.T) {
	f := &deletionSafetyTransport{exists: true}
	a := deletionSafetyAdapter(f)
	r := deletionSafetyResource()
	observation, err := a.Observe(context.Background(), r)
	if err != nil || !observation.Fresh || f.creates+f.stops != 0 {
		t.Fatal("read changed provider or lost observation")
	}
}
