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
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{DestructionIntents:[]DestructionIntent{{DesiredResourceID:"old",RealizedResourceID:"r1",Authorized:true,Reason:"explicit cleanup"}}})
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
	obs:=[]ObservedResourceState{{DesiredResourceID:"old",RealizedResourceID:"r1",Fresh:true}}
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{AllowDelete:true})
	if p.Actions[0].Action!=ReconcileDelete {t.Fatal(p.Actions[0])}
	if p.Actions[0].RealizedResourceID!="r1" {t.Fatal("delete lost exact realized-resource identity")}
}


func TestObserveOnlyResourceCannotDeleteEvenWithIntent(t *testing.T){
	g:=DesiredResourceGraph{ID:"g"}
	obs:=[]ObservedResourceState{{DesiredResourceID:"old",RealizedResourceID:"r1",Fresh:true,Lifecycle:LifecycleObserveOnly}}
	p:=PlanReconciliation(g,obs,ReconciliationPolicy{DestructionIntents:[]DestructionIntent{{DesiredResourceID:"old",RealizedResourceID:"r1",Authorized:true,Reason:"requested"}}})
	if p.Actions[0].Action==ReconcileDelete { t.Fatal("observe-only lifecycle allowed deletion") }
}
