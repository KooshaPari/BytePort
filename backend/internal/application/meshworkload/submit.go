package meshworkload

import (
	"context"
	"time"

	domain "github.com/byteport/api/internal/domain/deployment"
)

// DesiredStateStore persists owner-scoped mesh intent in the BytePort control plane.
type DesiredStateStore interface {
	Save(ctx context.Context, owner string, req DesiredStateRequest) error
}

// DeploymentStore adapts the existing deployment repository to mesh desired state.
// Provider/runtime credentials are never stored; only portable metadata is recorded.
type DeploymentStore struct{ repository domain.Repository }

// NewDeploymentStore creates a persistent mesh store backed by deployments.
func NewDeploymentStore(repository domain.Repository) *DeploymentStore {
	return &DeploymentStore{repository: repository}
}

// Save stores mesh identity and portable placement as deployment metadata.
func (s *DeploymentStore) Save(ctx context.Context, owner string, req DesiredStateRequest) error {
	dep, err := domain.NewDeployment(req.Name, owner, nil)
	if err != nil {
		return err
	}
	dep.SetCompositionMetadata(domain.CompositionMetadata{Digest: req.CompositionDigest, ArtifactRef: req.ArtifactRef})
	dep.SetProvider("execution_backend", req.ExecutionBackend)
	if len(req.Placement.Labels)+len(req.Placement.Constraints) > 0 {
		dep.SetProvider("placement", req.Placement)
	}
	return s.repository.Create(ctx, dep)
}

// SubmitDesiredStateUseCase validates mesh intent and returns an acknowledgement.
// Persistence and provider execution are deliberately separate control-plane concerns.
type SubmitDesiredStateUseCase struct{ store DesiredStateStore }

// NewSubmitDesiredStateUseCase constructs the stateless desired-state validator.
func NewSubmitDesiredStateUseCase(stores ...DesiredStateStore) *SubmitDesiredStateUseCase {
	var store DesiredStateStore
	if len(stores) > 0 {
		store = stores[0]
	}
	return &SubmitDesiredStateUseCase{store: store}
}

// Execute validates req for owner and returns a provider-neutral acknowledgement.
func (uc *SubmitDesiredStateUseCase) Execute(ctx context.Context, owner string, req DesiredStateRequest) (*DesiredStateResponse, error) {
	if err := req.Validate(owner); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if uc.store != nil {
		if err := uc.store.Save(ctx, owner, req); err != nil {
			return nil, err
		}
	}
	return &DesiredStateResponse{
		Name:              req.Name,
		Owner:             owner,
		CompositionDigest: req.CompositionDigest,
		ArtifactRef:       req.ArtifactRef,
		ExecutionBackend:  req.ExecutionBackend,
		Placement:         req.Placement,
		Status:            "accepted",
		AcceptedAt:        time.Now().UTC(),
	}, nil
}
