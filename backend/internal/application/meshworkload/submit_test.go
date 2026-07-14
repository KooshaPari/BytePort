package meshworkload

import (
	"context"
	"strings"
	"testing"
)

func validRequest() DesiredStateRequest {
	return DesiredStateRequest{
		Name: "demo",
		CompositionDigest: "sha256:" + strings.Repeat("a", 64),
		ArtifactRef: "oci://registry/demo",
		ExecutionBackend: "podman",
	}
}

func TestSubmitDesiredStateUsesAuthenticatedOwner(t *testing.T) {
	response, err := NewSubmitDesiredStateUseCase().Execute(context.Background(), "user-1", validRequest())
	if err != nil { t.Fatalf("valid request rejected: %v", err) }
	if response.Owner != "user-1" || response.Status != "accepted" { t.Fatalf("unexpected response: %+v", response) }
}

func TestDesiredStateRejectsImpersonationAndUnsupportedBackend(t *testing.T) {
	request := validRequest()
	if err := request.Validate(""); err == nil { t.Fatal("missing authenticated owner accepted") }
	request.ExecutionBackend = "aws"
	if err := request.Validate("user-1"); err == nil { t.Fatal("unknown execution backend accepted") }
}
