package models

import "testing"

func TestTargetAdapterResultCanReferenceProviderOperationWithoutCopyingState(t *testing.T){
	r:=InfrastructureApplyResult{
		ExternalOperation:&ExternalOperationRef{Provider:"ironic",TargetID:"lab",ExternalID:"node-op-1",LookupKind:"ironic-node"},
	}
	if r.ExternalOperation==nil || r.ExternalOperation.ExternalID!="node-op-1" {t.Fatal(r)}
	if r.Realized!=nil {t.Fatal("operation reference must not require invented realized state")}
}

func TestInfrastructureObservationSeparatesFreshnessFromState(t *testing.T){
	o:=InfrastructureObservation{RealizedResourceID:"r1",TargetID:"t1",State:"active",Fresh:false}
	if o.State!="active" || o.Fresh {t.Fatal(o)}
}
