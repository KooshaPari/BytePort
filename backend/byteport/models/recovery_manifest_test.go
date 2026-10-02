package models

import "testing"

const recoveryManifestFixture = `NAME: "example"
DESCRIPTION: "fixture"
SERVICES:
  - NAME: "main"
    PATH: "./frontend"
    PORT: 8080
  - NAME: "api"
    PATH: "./backend"
    PORT: 8081
`

func TestRecoveryManifestParsesAndDigestsExactBytes(t *testing.T){
	rev,m,err:=ParseRecoveryManifest([]byte(recoveryManifestFixture),"src-1","odin.nvms")
	if err!=nil{t.Fatal(err)}
	if m.Name!="example"||len(m.Services)!=2{t.Fatalf("%#v",m)}
	if rev.ContentDigest==""||string(rev.ID)!=rev.ContentDigest{t.Fatal("missing exact digest identity")}
	if rev.SourceSnapshot!="src-1"{t.Fatal(rev.SourceSnapshot)}
}

func TestRecoveryManifestRejectsInvalidCases(t *testing.T){
	cases:=map[string]string{
		"missing-name": "SERVICES:\n  - NAME: main\n    PATH: .\n    PORT: 80\n",
		"no-services": "NAME: x\nSERVICES: []\n",
		"duplicate": "NAME: x\nSERVICES:\n  - {NAME: a, PATH: ./a, PORT: 80}\n  - {NAME: a, PATH: ./b, PORT: 81}\n",
		"escape": "NAME: x\nSERVICES:\n  - {NAME: a, PATH: ../secret, PORT: 80}\n",
		"bad-port": "NAME: x\nSERVICES:\n  - {NAME: a, PATH: ./a, PORT: 70000}\n",
		"unknown-key": "NAME: x\nUNKNOWN: true\nSERVICES:\n  - {NAME: a, PATH: ./a, PORT: 80}\n",
		"malformed": "NAME: [\n",
	}
	for name,raw:=range cases{
		t.Run(name,func(t *testing.T){if _,_,err:=ParseRecoveryManifest([]byte(raw),"src","odin.nvms");err==nil{t.Fatal("expected rejection")}})
	}
}

func TestRecoveryManifestDigestChangesWithBytes(t *testing.T){
	a,_,err:=ParseRecoveryManifest([]byte(recoveryManifestFixture),"src","odin.nvms");if err!=nil{t.Fatal(err)}
	b,_,err:=ParseRecoveryManifest([]byte(recoveryManifestFixture+"\n"),"src","odin.nvms");if err!=nil{t.Fatal(err)}
	if a.ContentDigest==b.ContentDigest{t.Fatal("different exact bytes shared digest")}
}
