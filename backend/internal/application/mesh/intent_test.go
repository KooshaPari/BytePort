package mesh

import (
	"strings"
	"testing"
)

func TestWorkloadIntentValidate(t *testing.T) {
	i := WorkloadIntent{Owner: "alice", CompositionName: "demo", CompositionDigest: "sha256:" + strings.Repeat("a", 64), ArtifactRef: "oci://registry/demo", ExecutionBackend: "podman", Source: "git://github.com/KooshaPari/PhenoCompose/examples/composition-v0.yaml", Evidence: "run://phenocompose/69b4f35f"}
	if err := i.Validate(); err != nil {
		t.Fatalf("valid intent rejected: %v", err)
	}
	for _, mutate := range []func(*WorkloadIntent){func(x *WorkloadIntent) { x.Owner = "" }, func(x *WorkloadIntent) { x.ExecutionBackend = "aws" }, func(x *WorkloadIntent) { x.CompositionDigest = "sha256:bad" }} {
		bad := i
		mutate(&bad)
		if err := bad.Validate(); err == nil {
			t.Fatal("invalid intent accepted")
		}
	}
}
