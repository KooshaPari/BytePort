package models

import (
	"context"
	"fmt"
	"regexp"
)

var fullGitSHA = regexp.MustCompile("^[0-9a-fA-F]{40}$")

type GitRefResolver interface {
	ResolveRef(context.Context, string, string) (string, error)
}

func ResolveGitSourceSnapshot(ctx context.Context, resolver GitRefResolver, repository, reference string) (SourceSnapshot,error) {
	if repository==""||reference=="" { return SourceSnapshot{},fmt.Errorf("repository and reference are required") }
	commit,err:=resolver.ResolveRef(ctx,repository,reference)
	if err!=nil{return SourceSnapshot{},fmt.Errorf("resolve git ref: %w",err)}
	if !fullGitSHA.MatchString(commit){return SourceSnapshot{},fmt.Errorf("resolver returned non-immutable commit %q",commit)}
	return SourceSnapshot{ID:SourceSnapshotID(repository+"@"+commit),Provider:"git",Repository:repository,Commit:commit},nil
}
