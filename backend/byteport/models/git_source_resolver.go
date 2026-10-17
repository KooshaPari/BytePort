package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type GitSourceResolver struct{}

func (GitSourceResolver) ResolveSource(ctx context.Context, req ResolveSourceRequest) (SourceSnapshot, error) {
	if strings.TrimSpace(req.Repository)=="" { return SourceSnapshot{},fmt.Errorf("repository is required") }
	ref:=strings.TrimSpace(req.Reference); if ref=="" { ref="HEAD" }
	cmd:=exec.CommandContext(ctx,"git","-C",req.Repository,"rev-parse","--verify",ref+"^{commit}")
	out,err:=cmd.Output(); if err!=nil{return SourceSnapshot{},fmt.Errorf("resolve git ref %q: %w",ref,err)}
	commit:=strings.TrimSpace(string(out))
	if len(commit)!=40 { return SourceSnapshot{},fmt.Errorf("unexpected commit identity %q",commit) }
	sum:=sha256.Sum256([]byte(req.Repository+"\x00"+commit))
	return SourceSnapshot{
		ID:SourceSnapshotID(hex.EncodeToString(sum[:])),
		Provider:"git",
		Repository:req.Repository,
		Commit:commit,
		ResolvedAt:time.Now().UTC(),
	},nil
}
