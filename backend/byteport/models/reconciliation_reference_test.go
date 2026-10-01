package models

import "testing"

func TestReconciliationCreatesUpdatesAndNoopsExplicitly(t *testing.T){
	g:=DesiredResourceGraph{ID:"g",Manifest:"m",Resources:[]DesiredResource{
		{ID:"a",Kind:DesiredResourceService,ConfigDigest:"same"},
		{ID:"b",Kind:DesiredResourceManaged,ConfigDigest:"new"},
		{ID:"c",Kind:DesiredResourceNetwork,ConfigDigest:"create"},
	}}
	obs:=[]ObservedResourceState{
		{DesiredResourceID:"a",RealizedResourceID:"ra",ConfigDigest:"same",Fresh:true},
		{DesiredResourceID:"b",RealizedResourceID:"rb",ConfigDigest:"old",Fresh:true},
	}
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{})
	if p.Actions[0].Action!=ReconcileNoop {t.Fatal(p.Actions[0])}
	if p.Actions[1].Action!=ReconcileUpdate {t.Fatal(p.Actions[1])}
	if p.Actions[2].Action!=ReconcileCreate {t.Fatal(p.Actions[2])}
}

func TestStaleObservationProducesUnknownNotDestruction(t *testing.T){
	g:=DesiredResourceGraph{ID:"g"}
	obs:=[]ObservedResourceState{{DesiredResourceID:"old",RealizedResourceID:"r1",Fresh:false}}
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{DestructionIntents:[]DestructionIntent{{DesiredResourceID:"old",RealizedResourceID:"r1",AuthorizedBy:"test-authority",Reason:"explicit cleanup"}}})
	if p.Actions[0].Action!=ReconcileUnknown {t.Fatal(p.Actions[0])}
}

func TestMissingDesiredResourceDoesNotDeleteWithoutAuthorization(t *testing.T){
	g:=DesiredResourceGraph{ID:"g"}
	obs:=[]ObservedResourceState{{DesiredResourceID:"old",RealizedResourceID:"r1",Fresh:true,Lifecycle:LifecycleDestroyOnExplicitIntent}}
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{})
	if p.Actions[0].Action==ReconcileDelete {t.Fatal("destruction occurred without explicit authorization")}
	if p.Actions[0].Action!=ReconcileRead {t.Fatal(p.Actions[0])}
}

func TestExplicitDestructionCanPlanExactDelete(t *testing.T){
	g:=DesiredResourceGraph{ID:"g"}
	obs:=[]ObservedResourceState{{DesiredResourceID:"old",RealizedResourceID:"r1",Fresh:true,Lifecycle:LifecycleDestroyOnExplicitIntent}}
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{DestructionIntents:[]DestructionIntent{{
		ID:"destroy-1",DesiredResourceID:"old",RealizedResourceID:"r1",AuthorizedBy:"test-authority",Reason:"explicit cleanup",
	}}})
	if p.Actions[0].Action!=ReconcileDelete {t.Fatal(p.Actions[0])}
	if p.Actions[0].RealizedResourceID!="r1" {t.Fatal("delete lost exact realized-resource identity")}
}


func TestObserveOnlyResourceCannotDeleteEvenWithIntent(t *testing.T){
	g:=DesiredResourceGraph{ID:"g"}
	obs:=[]ObservedResourceState{{DesiredResourceID:"old",RealizedResourceID:"r1",Fresh:true,Lifecycle:LifecycleObserveOnly}}
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{DestructionIntents:[]DestructionIntent{{DesiredResourceID:"old",RealizedResourceID:"r1",AuthorizedBy:"test-authority",Reason:"requested"}}})
	if p.Actions[0].Action==ReconcileDelete { t.Fatal("observe-only lifecycle allowed deletion") }
}


func TestImmutableMismatchPlansReplaceNotUpdate(t *testing.T){
	g:=DesiredResourceGraph{ID:"g",Resources:[]DesiredResource{
		{ID:"host",Kind:DesiredResourceHost,ConfigDigest:"new",ReplaceOnChange:true},
	}}
	obs:=[]ObservedResourceState{{DesiredResourceID:"host",RealizedResourceID:"rh",ConfigDigest:"old",Fresh:true,Lifecycle:LifecycleManage}}
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{})
	if p.Actions[0].Action!=ReconcileReplace { t.Fatal(p.Actions[0]) }
}


func TestUnsupportedTargetActionBecomesUnknown(t *testing.T){
	p:=ResourcePlan{ID:"p",DesiredGraphID:"g",Actions:[]PlannedResourceAction{
		{DesiredResourceID:"host",Action:ReconcileReplace,Reason:"immutable change"},
	}}
	caps:=TargetCapabilities{TargetID:"limited",Provider:"fixture",SupportsObserve:true,SupportsCreate:true,SupportsUpdate:true,SupportsReplace:false,SupportsDelete:false}
	out:=ConstrainPlanToTarget(p,caps)
	if out.TargetID!="limited" {t.Fatal(out.TargetID)}
	if out.Actions[0].Action!=ReconcileUnknown {t.Fatal(out.Actions[0])}
}

func TestSamePlanCanRemainValidOnMoreCapableTarget(t *testing.T){
	p:=ResourcePlan{ID:"p",DesiredGraphID:"g",Actions:[]PlannedResourceAction{
		{DesiredResourceID:"host",Action:ReconcileReplace,Reason:"immutable change"},
	}}
	caps:=TargetCapabilities{TargetID:"baremetal",Provider:"ironic",BareMetal:true,SupportsObserve:true,SupportsCreate:true,SupportsUpdate:true,SupportsReplace:true,SupportsDelete:true}
	out:=ConstrainPlanToTarget(p,caps)
	if out.Actions[0].Action!=ReconcileReplace {t.Fatal(out.Actions[0])}
}
