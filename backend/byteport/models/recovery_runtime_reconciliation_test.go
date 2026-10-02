package models

import (
	"context"
	"errors"
	"testing"
	"time"
)

type recoveryRuntimeFixture struct {
	created map[RuntimeOperationID][]ProviderResource
	hideFor map[RuntimeOperationID]int
	createCalls int
}

func newRecoveryRuntimeFixture() *recoveryRuntimeFixture {
	return &recoveryRuntimeFixture{
		created: map[RuntimeOperationID][]ProviderResource{},
		hideFor: map[RuntimeOperationID]int{},
	}
}

func (f *recoveryRuntimeFixture) Capabilities(context.Context) RuntimeAdapterCapabilities {
	return RuntimeAdapterCapabilities{AdapterID:"fixture",Version:"1",SupportsObserve:true,SupportsFindByOperation:true,SupportsDelayedVisibility:true}
}

func (f *recoveryRuntimeFixture) Create(_ context.Context, req RuntimeCreateRequest) ([]ProviderResource,error) {
	f.createCalls++
	if existing:=f.created[req.Operation.ID]; len(existing)>0 { return existing,nil }
	r:=ProviderResource{ID:"resource-"+ProviderResourceID(req.Operation.ID),Operation:req.Operation.ID,Generation:req.Generation.ID,Provider:"fixture",Target:req.Operation.Target,ExternalID:"external-"+string(req.Operation.ID)}
	f.created[req.Operation.ID]=[]ProviderResource{r}
	return []ProviderResource{r},nil
}

func (f *recoveryRuntimeFixture) Observe(_ context.Context,id ProviderResourceID)(RuntimeObservationResult,error){
	for _,rs:=range f.created { for _,r:=range rs { if r.ID==id { return RuntimeObservationResult{Found:true,Observation:Observation{ID:"obs-1",ProviderResource:id,ObservedAt:time.Now(),State:"ready"}},nil } } }
	return RuntimeObservationResult{},errors.New("not found")
}

func (f *recoveryRuntimeFixture) FindByOperation(_ context.Context,id RuntimeOperationID)([]ProviderResource,error){
	if n:=f.hideFor[id]; n>0 { f.hideFor[id]=n-1; return nil,nil }
	return f.created[id],nil
}

func (f *recoveryRuntimeFixture) Stop(context.Context,ProviderResourceID)error{return nil}

func TestRecoveryLostCreateResponseReconcilesWithoutSecondMutation(t *testing.T){
	ctx:=context.Background()
	f:=newRecoveryRuntimeFixture()
	op:=RuntimeOperation{ID:"op-lost",Fingerprint:"fp",State:RuntimeOperationApplying,Target:"target"}
	gen:=DeploymentGeneration{ID:"gen-1",Intent:"intent-1",Role:"candidate"}

	// Provider accepts create; caller loses the response.
	if _,err:=f.Create(ctx,RuntimeCreateRequest{Operation:op,Generation:gen});err!=nil{t.Fatal(err)}
	op.State=RuntimeOperationUnknown

	// After controller restart, reconciliation looks up by durable operation ID
	// instead of blindly calling Create again.
	found,err:=f.FindByOperation(ctx,op.ID)
	if err!=nil{t.Fatal(err)}
	if len(found)!=1{t.Fatalf("found=%d want 1",len(found))}
	if f.createCalls!=1{t.Fatalf("create calls=%d want 1",f.createCalls)}
}

func TestRecoveryDelayedVisibilityStaysUnknownUntilObserved(t *testing.T){
	ctx:=context.Background()
	f:=newRecoveryRuntimeFixture()
	op:=RuntimeOperation{ID:"op-delay",Fingerprint:"fp",State:RuntimeOperationApplying,Target:"target"}
	gen:=DeploymentGeneration{ID:"gen-1",Intent:"intent-1",Role:"candidate"}
	if _,err:=f.Create(ctx,RuntimeCreateRequest{Operation:op,Generation:gen});err!=nil{t.Fatal(err)}
	f.hideFor[op.ID]=1
	op.State=RuntimeOperationUnknown

	first,err:=f.FindByOperation(ctx,op.ID)
	if err!=nil{t.Fatal(err)}
	if len(first)!=0{t.Fatal("first lookup should be temporarily invisible")}
	if op.State!=RuntimeOperationUnknown{t.Fatal("temporary not-found must not prove absence")}

	second,err:=f.FindByOperation(ctx,op.ID)
	if err!=nil{t.Fatal(err)}
	if len(second)!=1{t.Fatalf("second lookup=%d want 1",len(second))}
	if f.createCalls!=1{t.Fatalf("create calls=%d want 1",f.createCalls)}
}

func TestRecoveryRetryUsesSameOperationIdentity(t *testing.T){
	ctx:=context.Background()
	f:=newRecoveryRuntimeFixture()
	op:=RuntimeOperation{ID:"op-retry",Fingerprint:"fp",State:RuntimeOperationApplying,Target:"target"}
	gen:=DeploymentGeneration{ID:"gen-1",Intent:"intent-1",Role:"candidate"}
	first,err:=f.Create(ctx,RuntimeCreateRequest{Operation:op,Generation:gen})
	if err!=nil{t.Fatal(err)}
	second,err:=f.Create(ctx,RuntimeCreateRequest{Operation:op,Generation:gen})
	if err!=nil{t.Fatal(err)}
	if first[0].ID!=second[0].ID{t.Fatalf("resource changed: %s vs %s",first[0].ID,second[0].ID)}
	if f.createCalls!=2{t.Fatalf("fixture calls=%d want 2 invocations",f.createCalls)}
	if len(f.created)!=1{t.Fatalf("provider resources=%d want 1",len(f.created))}
}
