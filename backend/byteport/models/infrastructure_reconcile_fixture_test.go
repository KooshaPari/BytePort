package models

import "testing"

func TestFixtureReconciliationMixedGraph(t *testing.T){
	g:=DesiredResourceGraph{ID:"g",Resources:[]DesiredResource{
		{ID:"svc",Kind:DesiredResourceService,ConfigDigest:"v2",Target:"t",Lifecycle:LifecycleManage},
		{ID:"db",Kind:DesiredResourceManaged,ConfigDigest:"db1",Target:"t",Lifecycle:LifecycleManage},
	}}
	real:=map[string]RealizedResourceView{"svc":{Resource:RealizedResource{ID:"r-svc",DesiredResourceID:"svc"},ConfigDigest:"v1",Observed:true}}
	p,err:=PlanFixtureReconciliation(g,real,TargetCapabilities{TargetID:"t",SupportsCreate:true,SupportsUpdate:true})
	if err!=nil{t.Fatal(err)}
	if p.Actions[0].Action!=ReconcileUpdate||p.Actions[1].Action!=ReconcileCreate{t.Fatalf("%#v",p.Actions)}
}

func TestFixtureReconciliationUnknownNeverDeletes(t *testing.T){
	g:=DesiredResourceGraph{ID:"g",Resources:[]DesiredResource{{ID:"x",ConfigDigest:"v",Target:"t",Lifecycle:LifecycleManage}}}
	real:=map[string]RealizedResourceView{"x":{Resource:RealizedResource{ID:"rx",DesiredResourceID:"x"},Observed:false}}
	p,err:=PlanFixtureReconciliation(g,real,TargetCapabilities{TargetID:"t",SupportsDelete:true})
	if err!=nil{t.Fatal(err)}
	if p.Actions[0].Action!=ReconcileUnknown{t.Fatalf("%#v",p.Actions[0])}
}

func TestFixtureReconciliationCapabilityMismatchFails(t *testing.T){
	g:=DesiredResourceGraph{ID:"g",Resources:[]DesiredResource{{ID:"x",ConfigDigest:"v2",Target:"t",Lifecycle:LifecycleManage,ReplaceOnChange:true}}}
	real:=map[string]RealizedResourceView{"x":{Resource:RealizedResource{ID:"rx",DesiredResourceID:"x"},ConfigDigest:"v1",Observed:true}}
	if _,err:=PlanFixtureReconciliation(g,real,TargetCapabilities{TargetID:"t",SupportsReplace:false});err==nil{t.Fatal("expected explicit capability failure")}
}
