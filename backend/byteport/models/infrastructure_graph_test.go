package models

import "testing"

func TestDesiredGraphAllowsNonArtifactInfrastructure(t *testing.T) {
	g := DesiredResourceGraph{
		ID:"graph-1",
		Manifest:"manifest-1",
		Resources:[]DesiredResource{
			{ID:"svc",Kind:DesiredResourceService,ConfigDigest:"a",Target:"t1"},
			{ID:"db",Kind:DesiredResourceManaged,ConfigDigest:"b",Target:"t1"},
		},
	}
	if g.Resources[1].Artifact != nil {
		t.Fatal("managed resource must not require fake build artifact")
	}
}

func TestResourcePlanIsNotRealizedState(t *testing.T) {
	p:=ResourcePlan{
		ID:"plan-1",DesiredGraphID:"graph-1",TargetID:"t1",
		Actions:[]PlannedResourceAction{{DesiredResourceID:"svc",Action:ReconcileCreate,Reason:"missing"}},
	}
	if p.Actions[0].Action!=ReconcileCreate { t.Fatal(p) }
}

func TestBareMetalIsTargetCapabilityNotResourceKind(t *testing.T) {
	c:=TargetCapabilities{TargetID:"bm-1",Provider:"ironic",BareMetal:true,SupportsObserve:true}
	if !c.BareMetal || c.Provider!="ironic" { t.Fatal(c) }
}

func TestUnknownReconciliationIsFirstClass(t *testing.T) {
	a:=PlannedResourceAction{DesiredResourceID:"svc",Action:ReconcileUnknown,Reason:"observation stale"}
	if a.Action!=ReconcileUnknown { t.Fatal(a) }
}

func TestDeleteRequiresExactExplicitDestructionIntent(t *testing.T) {
	d:=DesiredResource{ID:"db",Kind:DesiredResourceManaged,Target:"t1",Lifecycle:LifecycleDestroyOnExplicitIntent}
	r:=RealizedResource{ID:"real-db",DesiredResourceID:"db",TargetID:"t1",Provider:"fixture",ExternalID:"ext"}
	if d.AllowsDelete(nil,r) { t.Fatal("nil intent authorized delete") }
	wrong:=DestructionIntent{ID:"destroy-1",DesiredResourceID:"db",RealizedResourceID:"other",AuthorizedBy:"user",Reason:"test"}
	if d.AllowsDelete(&wrong,r) { t.Fatal("wrong resource authorized delete") }
	ok:=DestructionIntent{ID:"destroy-2",DesiredResourceID:"db",RealizedResourceID:"real-db",AuthorizedBy:"user",Reason:"test"}
	if !d.AllowsDelete(&ok,r) { t.Fatal("exact explicit intent did not authorize delete") }
}

func TestObserveOnlyNeverAllowsDelete(t *testing.T) {
	d:=DesiredResource{ID:"db",Kind:DesiredResourceManaged,Target:"t1",Lifecycle:LifecycleObserveOnly}
	r:=RealizedResource{ID:"real-db",DesiredResourceID:"db",TargetID:"t1"}
	intent:=DestructionIntent{ID:"destroy",DesiredResourceID:"db",RealizedResourceID:"real-db",AuthorizedBy:"user"}
	if d.AllowsDelete(&intent,r) { t.Fatal("observe-only resource allowed delete") }
}
