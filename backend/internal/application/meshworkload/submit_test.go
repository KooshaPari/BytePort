package meshworkload

import (
	"context"
	"strings"
	"testing"

	"github.com/byteport/api/internal/domain/deployment"
	postgres "github.com/byteport/api/internal/infrastructure/persistence/postgres"
)

func validRequest() DesiredStateRequest {
	return DesiredStateRequest{
		Name:              "demo",
		CompositionDigest: "sha256:" + strings.Repeat("a", 64),
		ArtifactRef:       "oci://registry/demo",
		ExecutionBackend:  "podman",
	}
}

func TestSubmitDesiredStateUsesAuthenticatedOwner(t *testing.T) {
	response, err := NewSubmitDesiredStateUseCase().Execute(context.Background(), "user-1", validRequest())
	if err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	if response.Owner != "user-1" || response.Status != "accepted" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestDesiredStateRejectsImpersonationAndUnsupportedBackend(t *testing.T) {
	request := validRequest()
	if err := request.Validate(""); err == nil {
		t.Fatal("missing authenticated owner accepted")
	}
	request.ExecutionBackend = "aws"
	if err := request.Validate("user-1"); err == nil {
		t.Fatal("unknown execution backend accepted")
	}
}

type recordingStore struct {
	owner   string
	request DesiredStateRequest
}

func (s *recordingStore) Save(_ context.Context, owner string, request DesiredStateRequest) error {
	s.owner, s.request = owner, request
	return nil
}

func TestSubmitDesiredStatePersistsOwnerScopedIntent(t *testing.T) {
	store := new(recordingStore)
	_, err := NewSubmitDesiredStateUseCase(store).Execute(context.Background(), "user-1", validRequest())
	if err != nil {
		t.Fatalf("persisted request rejected: %v", err)
	}
	if store.owner != "user-1" || store.request.Name != "demo" {
		t.Fatalf("unexpected persisted intent: %+v", store)
	}
}

type roundTripRepository struct {
	deployment *deployment.Deployment
}

func (r *roundTripRepository) Create(_ context.Context, dep *deployment.Deployment) error {
	model, err := postgres.DomainToModel(dep)
	if err != nil {
		return err
	}
	r.deployment, err = postgres.ModelToDomain(model)
	return err
}

func (r *roundTripRepository) Update(context.Context, *deployment.Deployment) error { return nil }
func (r *roundTripRepository) Delete(context.Context, string) error                 { return nil }
func (r *roundTripRepository) FindByUUID(context.Context, string) (*deployment.Deployment, error) {
	return r.deployment, nil
}
func (r *roundTripRepository) FindByOwner(context.Context, string) ([]*deployment.Deployment, error) {
	if r.deployment == nil {
		return nil, nil
	}
	return []*deployment.Deployment{r.deployment}, nil
}
func (r *roundTripRepository) FindByProject(context.Context, string) ([]*deployment.Deployment, error) {
	return nil, nil
}
func (r *roundTripRepository) FindByStatus(context.Context, deployment.Status) ([]*deployment.Deployment, error) {
	return nil, nil
}
func (r *roundTripRepository) List(context.Context, int, int) ([]*deployment.Deployment, error) {
	return nil, nil
}
func (r *roundTripRepository) Count(context.Context) (int64, error) { return 0, nil }
func (r *roundTripRepository) CountByOwner(context.Context, string) (int64, error) {
	return 0, nil
}

func TestDesiredStatePlacementRoundTripsThroughPersistence(t *testing.T) {
	request := validRequest()
	request.Placement = Placement{
		Region:      "us-west-2",
		Zone:        "us-west-2b",
		NodePool:    "gpu",
		Labels:      map[string]string{"accelerator": "nvidia"},
		Constraints: map[string]string{"arch": "arm64"},
	}
	repository := new(roundTripRepository)
	store := NewDeploymentStore(repository)
	if _, err := NewSubmitDesiredStateUseCase(store).Execute(context.Background(), "user-1", request); err != nil {
		t.Fatalf("persisted request rejected: %v", err)
	}
	responses, err := store.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected one persisted workload, got %d", len(responses))
	}
	if got := responses[0].Placement; got.Region != request.Placement.Region || got.Zone != request.Placement.Zone || got.NodePool != request.Placement.NodePool {
		t.Fatalf("placement scalar fields did not round-trip: %+v", got)
	}
	if got := responses[0].Placement.Labels["accelerator"]; got != "nvidia" {
		t.Fatalf("placement labels did not round-trip: %+v", responses[0].Placement.Labels)
	}
	if got := responses[0].Placement.Constraints["arch"]; got != "arm64" {
		t.Fatalf("placement constraints did not round-trip: %+v", responses[0].Placement.Constraints)
	}
}
