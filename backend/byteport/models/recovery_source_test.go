package models

import (
	"context"
	"testing"
)

type movingRefResolver struct{ commits []string; calls int }
func (r *movingRefResolver) ResolveRef(context.Context,string,string)(string,error){
	v:=r.commits[r.calls];r.calls++;return v,nil
}

func TestResolveGitSourceSnapshotFreezesCommitAgainstBranchMove(t *testing.T){
	a:="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	b:="bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	r:=&movingRefResolver{commits:[]string{a,b}}
	snap,err:=ResolveGitSourceSnapshot(context.Background(),r,"owner/repo","main")
	if err!=nil{t.Fatal(err)}
	// Branch moves after resolution. The already-created snapshot remains A.
	moved,err:=r.ResolveRef(context.Background(),"owner/repo","main");if err!=nil{t.Fatal(err)}
	if moved!=b{t.Fatal(moved)}
	if snap.Commit!=a{t.Fatalf("snapshot changed to %s",snap.Commit)}
}

func TestResolveGitSourceSnapshotRejectsMutableResolverOutput(t *testing.T){
	r:=&movingRefResolver{commits:[]string{"main"}}
	if _,err:=ResolveGitSourceSnapshot(context.Background(),r,"owner/repo","main");err==nil{t.Fatal("expected immutable SHA rejection")}
}
