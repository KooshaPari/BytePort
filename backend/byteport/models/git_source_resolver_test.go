package models

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitRun(t *testing.T,dir string,args ...string) string {
	t.Helper()
	cmd:=exec.Command("git",append([]string{"-C",dir},args...)...)
	out,err:=cmd.CombinedOutput(); if err!=nil{t.Fatalf("git %v: %v\n%s",args,err,out)}
	return strings.TrimSpace(string(out))
}

func TestGitSourceResolverFreezesMutableBranch(t *testing.T){
	dir:=t.TempDir()
	gitRun(t,dir,"init")
	gitRun(t,dir,"config","user.email","recovery@example.invalid")
	gitRun(t,dir,"config","user.name","Recovery Fixture")
	if err:=os.WriteFile(filepath.Join(dir,"app.txt"),[]byte("v1"),0644);err!=nil{t.Fatal(err)}
	gitRun(t,dir,"add","app.txt");gitRun(t,dir,"commit","-m","v1")
	gitRun(t,dir,"branch","-M","main")

	r:=GitSourceResolver{}
	first,err:=r.ResolveSource(context.Background(),ResolveSourceRequest{Repository:dir,Reference:"main"})
	if err!=nil{t.Fatal(err)}

	if err:=os.WriteFile(filepath.Join(dir,"app.txt"),[]byte("v2"),0644);err!=nil{t.Fatal(err)}
	gitRun(t,dir,"add","app.txt");gitRun(t,dir,"commit","-m","v2")
	second,err:=r.ResolveSource(context.Background(),ResolveSourceRequest{Repository:dir,Reference:"main"})
	if err!=nil{t.Fatal(err)}

	if first.Commit==second.Commit{t.Fatal("branch move did not change newly resolved snapshot")}
	if first.Commit!=gitRun(t,dir,"rev-parse","HEAD~1"){t.Fatalf("first snapshot drifted: %s",first.Commit)}
}

func TestGitSourceResolverRejectsMissingRef(t *testing.T){
	dir:=t.TempDir();gitRun(t,dir,"init")
	_,err:= (GitSourceResolver{}).ResolveSource(context.Background(),ResolveSourceRequest{Repository:dir,Reference:"missing"})
	if err==nil{t.Fatal("expected missing ref rejection")}
}
