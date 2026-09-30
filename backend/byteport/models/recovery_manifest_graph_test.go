package models

import "testing"

func TestHistoricalManifestProjectsIntoMatureDesiredGraph(t *testing.T) {
	rev,m,err:=ParseRecoveryManifest([]byte(recoveryManifestFixture),"src-1","odin.nvms")
	if err!=nil{t.Fatal(err)}
	g,err:=RecoveryManifestToDesiredGraph(rev,m,"target-1")
	if err!=nil{t.Fatal(err)}
	if g.Manifest!=rev.ID {t.Fatalf("manifest=%s want %s",g.Manifest,rev.ID)}
	if len(g.Resources)!=2 {t.Fatalf("resources=%d",len(g.Resources))}
	for _,r:=range g.Resources {
		if r.Kind!=DesiredResourceService {t.Fatal(r)}
		if r.Target!="target-1" {t.Fatal(r)}
		if r.Lifecycle!=LifecycleManage {t.Fatal(r)}
		if r.ConfigDigest=="" {t.Fatal("missing per-resource digest")}
		if r.Artifact!=nil {t.Fatal("compatibility import fabricated build artifact")}
	}
}

func TestHistoricalManifestImportRequiresExplicitTarget(t *testing.T) {
	rev,m,err:=ParseRecoveryManifest([]byte(recoveryManifestFixture),"src-1","odin.nvms")
	if err!=nil{t.Fatal(err)}
	if _,err:=RecoveryManifestToDesiredGraph(rev,m,"");err==nil{
		t.Fatal("historical import invented target placement")
	}
}
