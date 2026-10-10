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


func TestHistoricalManifestGraphDoesNotAuthorizeBuildOrEnv(t *testing.T) {
	rawA := []byte(`NAME: app
SERVICES:
  - NAME: api
    PATH: .
    PORT: 8080
    RUNTIME: python
    BUILD: ["echo", "first"]
    ENV:
      SECRET: first
`)
	rawB := []byte(`NAME: app
SERVICES:
  - NAME: api
    PATH: .
    PORT: 8080
    RUNTIME: python
    BUILD: ["rm", "-rf", "/"]
    ENV:
      SECRET: second
`)

	revA, manifestA, err := ParseRecoveryManifest(rawA, "src-1", "byteport.yaml")
	if err != nil { t.Fatal(err) }
	revB, manifestB, err := ParseRecoveryManifest(rawB, "src-1", "byteport.yaml")
	if err != nil { t.Fatal(err) }
	if revA.ID == revB.ID {
		t.Fatal("exact manifest revision must still capture BUILD/ENV byte changes")
	}

	graphA, err := RecoveryManifestToDesiredGraph(revA, manifestA, "target-1")
	if err != nil { t.Fatal(err) }
	graphB, err := RecoveryManifestToDesiredGraph(revB, manifestB, "target-1")
	if err != nil { t.Fatal(err) }
	if graphA.Resources[0].ConfigDigest != graphB.Resources[0].ConfigDigest {
		t.Fatal("graph projection treated untrusted BUILD/ENV as authorized desired-resource configuration")
	}
	if graphA.Resources[0].Artifact != nil || graphB.Resources[0].Artifact != nil {
		t.Fatal("manifest projection fabricated a build artifact")
	}
}
