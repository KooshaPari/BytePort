package meshworkload

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	domain "github.com/byteport/api/internal/domain/deployment"
)

// DesiredStateSaver persists owner-scoped mesh intent in the BytePort control plane.
type DesiredStateSaver interface {
	Save(ctx context.Context, owner string, req DesiredStateRequest) error
}

// DesiredStateIdentitySaver is an additive extension for stores that can
// return the stable control-plane identity assigned while saving. Keeping this
// separate from DesiredStateSaver preserves compatibility with existing
// in-memory/test savers that only need the original error result.
type DesiredStateIdentitySaver interface {
	DesiredStateSaver
	SaveWithID(ctx context.Context, owner string, req DesiredStateRequest) (string, error)
}

// DesiredStateReader reads owner-scoped desired state for reconciliation.
type DesiredStateReader interface {
	List(ctx context.Context, owner string) ([]DesiredStateResponse, error)
}

// DeploymentStore adapts the existing deployment repository to mesh desired state.
// Provider/runtime credentials are never stored; only portable metadata is recorded.
type DeploymentStore struct {
	repository domain.Repository
	mu         sync.Mutex
}

// NewDeploymentStore creates a persistent mesh store backed by deployments.
func NewDeploymentStore(repository domain.Repository) *DeploymentStore {
	return &DeploymentStore{repository: repository}
}

// Save stores mesh identity and portable placement as deployment metadata.
func (s *DeploymentStore) Save(ctx context.Context, owner string, req DesiredStateRequest) error {
	_, err := s.SaveWithID(ctx, owner, req)
	return err
}

// SaveWithID stores mesh identity and returns the stable deployment UUID.
func (s *DeploymentStore) SaveWithID(ctx context.Context, owner string, req DesiredStateRequest) (string, error) {
	// Serialize the read/create pair for callers sharing this process. A
	// database-level uniqueness constraint or transaction can extend this
	// guarantee across API instances without changing this contract.
	s.mu.Lock()
	defer s.mu.Unlock()

	// The owner/name pair is the stable workload slot. A retry with the same
	// composition digest replays the original control-plane identity, while a
	// changed digest is rejected instead of silently creating a second workload
	// under the same name.
	existing, err := s.repository.FindByOwner(ctx, owner)
	if err != nil {
		return "", err
	}
	var replayID string
	for _, candidate := range existing {
		if candidate == nil || candidate.Name() != req.Name {
			continue
		}
		metadata := candidate.CompositionMetadata()
		if metadata != nil && metadata.Digest == req.CompositionDigest {
			if metadata.Source != req.Source || metadata.Evidence != req.Evidence {
				return "", &ConflictError{Message: fmt.Sprintf("mesh workload %q already exists with different handoff evidence", req.Name)}
			}
			replayID = candidate.UUID()
			continue
		}
		return "", &ConflictError{Message: fmt.Sprintf("mesh workload %q already exists with a different composition digest", req.Name)}
	}
	if replayID != "" {
		return replayID, nil
	}

	dep, err := domain.NewDeployment(req.Name, owner, nil)
	if err != nil {
		return "", err
	}
	dep.SetCompositionMetadata(domain.CompositionMetadata{
		Digest:      req.CompositionDigest,
		ArtifactRef: req.ArtifactRef,
		Source:      req.Source,
		Verified:    true,
		Evidence:    req.Evidence,
	})
	dep.SetProvider("execution_backend", req.ExecutionBackend)
	if len(req.Placement.Labels)+len(req.Placement.Constraints) > 0 {
		dep.SetProvider("placement", req.Placement)
	}
	if err := s.repository.Create(ctx, dep); err != nil {
		return "", err
	}
	return dep.UUID(), nil
}

// List returns persisted mesh intents owned by the authenticated principal.
func (s *DeploymentStore) List(ctx context.Context, owner string) ([]DesiredStateResponse, error) {
	deployments, err := s.repository.FindByOwner(ctx, owner)
	if err != nil {
		return nil, err
	}
	responses := make([]DesiredStateResponse, 0, len(deployments))
	for _, dep := range deployments {
		metadata := dep.CompositionMetadata()
		if metadata == nil {
			continue
		}
		backend, _ := dep.Providers()["execution_backend"].(string)
		responses = append(responses, DesiredStateResponse{ID: dep.UUID(), Name: dep.Name(), Owner: dep.Owner(), CompositionDigest: metadata.Digest, ArtifactRef: metadata.ArtifactRef, ExecutionBackend: backend, Source: metadata.Source, Verified: metadata.Verified, Evidence: metadata.Evidence, Placement: placementFromProvider(dep.Providers()["placement"]), Status: dep.Status().String(), AcceptedAt: dep.CreatedAt()})
	}
	return responses, nil
}

// placementFromProvider decodes the portable scheduling intent after a persistence
// round trip. JSON-backed provider configuration is reconstructed as
// map[string]interface{}, while in-memory stores may retain the concrete type.
// Keep this conversion here so provider adapters remain unaware of persistence
// representation details.
func placementFromProvider(value interface{}) Placement {
	switch placement := value.(type) {
	case Placement:
		return placement
	case map[string]string:
		encoded, err := json.Marshal(placement)
		if err != nil {
			return Placement{}
		}
		var decoded Placement
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			return Placement{}
		}
		return decoded
	case map[string]interface{}:
		encoded, err := json.Marshal(placement)
		if err != nil {
			return Placement{}
		}
		var decoded Placement
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			return Placement{}
		}
		return decoded
	default:
		return Placement{}
	}
}

// SubmitDesiredStateUseCase validates mesh intent and returns an acknowledgement.
// Persistence and provider execution are deliberately separate control-plane concerns.
type SubmitDesiredStateUseCase struct{ store DesiredStateSaver }

// NewSubmitDesiredStateUseCase constructs the stateless desired-state validator.
func NewSubmitDesiredStateUseCase(stores ...DesiredStateSaver) *SubmitDesiredStateUseCase {
	var store DesiredStateSaver
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
	var id string
	if uc.store != nil {
		if identitySaver, ok := uc.store.(DesiredStateIdentitySaver); ok {
			var err error
			id, err = identitySaver.SaveWithID(ctx, owner, req)
			if err != nil {
				return nil, err
			}
		} else if err := uc.store.Save(ctx, owner, req); err != nil {
			return nil, err
		}
	}
	return &DesiredStateResponse{
		ID:                id,
		Name:              req.Name,
		Owner:             owner,
		CompositionDigest: req.CompositionDigest,
		ArtifactRef:       req.ArtifactRef,
		ExecutionBackend:  req.ExecutionBackend,
		Source:            req.Source,
		Verified:          true,
		Evidence:          req.Evidence,
		Placement:         req.Placement,
		Status:            "accepted",
		AcceptedAt:        time.Now().UTC(),
	}, nil
}

// List returns persisted desired state for an authenticated owner.
func (uc *SubmitDesiredStateUseCase) List(ctx context.Context, owner string) ([]DesiredStateResponse, error) {
	reader, ok := uc.store.(DesiredStateReader)
	if !ok {
		return nil, &ValidationError{Message: "mesh desired-state reader is not configured"}
	}
	if strings.TrimSpace(owner) == "" {
		return nil, &ValidationError{Message: "authenticated owner is required"}
	}
	return reader.List(ctx, owner)
}
