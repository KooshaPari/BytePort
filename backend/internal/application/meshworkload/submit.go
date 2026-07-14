package meshworkload

import (
	"context"
	"time"
)

// SubmitDesiredStateUseCase validates mesh intent and returns an acknowledgement.
// Persistence and provider execution are deliberately separate control-plane concerns.
type SubmitDesiredStateUseCase struct{}

// NewSubmitDesiredStateUseCase constructs the stateless desired-state validator.
func NewSubmitDesiredStateUseCase() *SubmitDesiredStateUseCase { return &SubmitDesiredStateUseCase{} }

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
