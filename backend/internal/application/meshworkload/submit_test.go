package meshworkload

import (
	"context"
	domain "github.com/byteport/api/internal/domain/deployment"
	"strings"
	"testing"
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

type deploymentRepositoryStub struct {
	created *domain.Deployment
	listed  []*domain.Deployment
}

func (s *deploymentRepositoryStub) Create(_ context.Context, dep *domain.Deployment) error {
	s.created = dep
	return nil
}
func (s *deploymentRepositoryStub) Update(context.Context, *domain.Deployment) error { return nil }
func (s *deploymentRepositoryStub) Delete(context.Context, string) error             { return nil }
func (s *deploymentRepositoryStub) FindByUUID(context.Context, string) (*domain.Deployment, error) {
	return nil, nil
}
func (s *deploymentRepositoryStub) FindByOwner(context.Context, string) ([]*domain.Deployment, error) {
	return s.listed, nil
}
func (s *deploymentRepositoryStub) FindByProject(context.Context, string) ([]*domain.Deployment, error) {
	return nil, nil
}
func (s *deploymentRepositoryStub) FindByStatus(context.Context, domain.Status) ([]*domain.Deployment, error) {
	return nil, nil
}
func (s *deploymentRepositoryStub) List(context.Context, int, int) ([]*domain.Deployment, error) {
	return nil, nil
}
func (s *deploymentRepositoryStub) Count(context.Context) (int64, error) { return 0, nil }
func (s *deploymentRepositoryStub) CountByOwner(context.Context, string) (int64, error) {
	return 0, nil
}

func TestDeploymentStoreSavePersistsAllPlacementFields(t *testing.T) {
	repository := new(deploymentRepositoryStub)
	request := validRequest()
	request.Placement = Placement{Region: "us-east-1", Zone: "us-east-1a", NodePool: "gpu", Labels: map[string]string{"tier": "batch"}}
	if err := NewDeploymentStore(repository).Save(context.Background(), "user-1", request); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	placement, ok := repository.created.Providers()["placement"].(Placement)
	if !ok || placement.Region != request.Placement.Region || placement.Zone != request.Placement.Zone || placement.NodePool != request.Placement.NodePool {
		t.Fatalf("placement was not persisted: %#v", repository.created.Providers()["placement"])
	}
}

func TestDeploymentStoreListReadsPlacementMetadata(t *testing.T) {
	dep, err := domain.NewDeployment("demo", "user-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	dep.SetCompositionMetadata(domain.CompositionMetadata{Digest: "sha256:" + strings.Repeat("a", 64), ArtifactRef: "oci://registry/demo"})
	dep.SetProvider("execution_backend", "podman")
	dep.SetProvider("placement", map[string]interface{}{"region": "us-west-2", "zone": "us-west-2b", "node_pool": "general", "labels": map[string]interface{}{"tier": "batch"}})
	repository := &deploymentRepositoryStub{listed: []*domain.Deployment{dep}}
	responses, err := NewDeploymentStore(repository).List(context.Background(), "user-1")
	if err != nil || len(responses) != 1 {
		t.Fatalf("list failed: %v (%d responses)", err, len(responses))
	}
	response := responses[0]
	if response.Placement.Region != "us-west-2" || response.Placement.Zone != "us-west-2b" || response.Placement.NodePool != "general" || response.Placement.Labels["tier"] != "batch" {
		t.Fatalf("placement readback omitted or malformed: %#v", response.Placement)
	}
}
