package meshworkload

import (
	"context"
	"errors"
	"strings"
	"sync"
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

func TestValidationErrorExposesMessage(t *testing.T) {
	err := (&ValidationError{Message: "invalid desired state"}).Error()
	if err != "invalid desired state" {
		t.Fatalf("unexpected validation error message: %q", err)
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
	deployment  *deployment.Deployment
	createCalls int
}

func (r *roundTripRepository) Create(_ context.Context, dep *deployment.Deployment) error {
	r.createCalls++
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
	response, err := NewSubmitDesiredStateUseCase(store).Execute(context.Background(), "user-1", request)
	if err != nil {
		t.Fatalf("persisted request rejected: %v", err)
	}
	if response.ID == "" {
		t.Fatal("submit response did not expose a stable workload ID")
	}
	if response.ID != repository.deployment.UUID() {
		t.Fatalf("submit response ID %q did not match persisted deployment UUID %q", response.ID, repository.deployment.UUID())
	}
	responses, err := store.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected one persisted workload, got %d", len(responses))
	}
	if responses[0].ID != response.ID {
		t.Fatalf("list returned workload ID %q, want stable ID %q", responses[0].ID, response.ID)
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

func TestDesiredStateReplayReturnsExistingIdentity(t *testing.T) {
	repository := new(roundTripRepository)
	store := NewDeploymentStore(repository)
	useCase := NewSubmitDesiredStateUseCase(store)

	first, err := useCase.Execute(context.Background(), "user-1", validRequest())
	if err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	second, err := useCase.Execute(context.Background(), "user-1", validRequest())
	if err != nil {
		t.Fatalf("replay submit failed: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("replay returned ID %q, want original ID %q", second.ID, first.ID)
	}
	if repository.createCalls != 1 {
		t.Fatalf("replay created %d deployments, want exactly one", repository.createCalls)
	}
}

func TestDesiredStateChangedDigestConflictsWithExistingIdentity(t *testing.T) {
	repository := new(roundTripRepository)
	store := NewDeploymentStore(repository)
	useCase := NewSubmitDesiredStateUseCase(store)

	if _, err := useCase.Execute(context.Background(), "user-1", validRequest()); err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	changed := validRequest()
	changed.CompositionDigest = "sha256:" + strings.Repeat("b", 64)
	_, err := useCase.Execute(context.Background(), "user-1", changed)
	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("changed digest error = %v, want ConflictError", err)
	}
	if repository.createCalls != 1 {
		t.Fatalf("changed digest created %d deployments, want exactly one", repository.createCalls)
	}
}

func TestDesiredStateConcurrentReplayCreatesOneIdentity(t *testing.T) {
	repository := new(roundTripRepository)
	store := NewDeploymentStore(repository)
	useCase := NewSubmitDesiredStateUseCase(store)

	const attempts = 8
	responses := make(chan *DesiredStateResponse, attempts)
	errorsCh := make(chan error, attempts)
	var waitGroup sync.WaitGroup
	waitGroup.Add(attempts)
	for i := 0; i < attempts; i++ {
		go func() {
			defer waitGroup.Done()
			response, err := useCase.Execute(context.Background(), "user-1", validRequest())
			if err != nil {
				errorsCh <- err
				return
			}
			responses <- response
		}()
	}
	waitGroup.Wait()
	close(responses)
	close(errorsCh)
	for err := range errorsCh {
		t.Fatalf("concurrent replay failed: %v", err)
	}
	var stableID string
	for response := range responses {
		if stableID == "" {
			stableID = response.ID
			continue
		}
		if response.ID != stableID {
			t.Fatalf("concurrent replay returned IDs %q and %q", stableID, response.ID)
		}
	}
	if repository.createCalls != 1 {
		t.Fatalf("concurrent replay created %d deployments, want exactly one", repository.createCalls)
	}
}
