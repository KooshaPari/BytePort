// Package meshworkload contains the provider-neutral compute-mesh desired-state contract.
package meshworkload

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-fA-F]{64}$`)
var namePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// DesiredStateRequest is the provider-neutral request submitted by a composition client.
// Owner is deliberately absent: it is always taken from the authenticated request context.
type DesiredStateRequest struct {
	Name              string `json:"name" binding:"required"`
	CompositionDigest string `json:"composition_digest" binding:"required"`
	ArtifactRef       string `json:"artifact_ref" binding:"required"`
	ExecutionBackend  string `json:"execution_backend" binding:"required"`
	// Source and Evidence make the cross-repository handoff auditable without
	// allowing provider credentials or mutable runtime state into the request.
	Source    string    `json:"source" binding:"required"`
	Evidence  string    `json:"evidence" binding:"required"`
	Placement Placement `json:"placement"`
}

// Placement contains portable scheduling intent. It must not contain provider credentials
// or provider-specific resource IDs; provider adapters resolve those separately.
type Placement struct {
	Region      string            `json:"region,omitempty"`
	Zone        string            `json:"zone,omitempty"`
	NodePool    string            `json:"node_pool,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Constraints map[string]string `json:"constraints,omitempty"`
}

// DesiredStateResponse acknowledges validated intent without claiming provider deployment.
type DesiredStateResponse struct {
	// ID is the stable control-plane identity assigned to the persisted workload.
	// It is the deployment UUID internally, and is intentionally separate from
	// provider/runtime resource IDs.
	ID                string    `json:"id,omitempty"`
	Name              string    `json:"name"`
	Owner             string    `json:"owner"`
	CompositionDigest string    `json:"composition_digest"`
	ArtifactRef       string    `json:"artifact_ref"`
	ExecutionBackend  string    `json:"execution_backend"`
	Source            string    `json:"source"`
	Verified          bool      `json:"verified"`
	Evidence          string    `json:"evidence"`
	Placement         Placement `json:"placement"`
	Status            string    `json:"status"`
	AcceptedAt        time.Time `json:"accepted_at"`
}

// ValidationError describes a malformed desired-state request.
type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }

// ConflictError describes an idempotency conflict for an existing workload
// identity. It is kept separate from ValidationError so transports can map a
// changed composition digest to HTTP 409 without changing existing clients.
type ConflictError struct{ Message string }

func (e *ConflictError) Error() string { return e.Message }

// Validate checks the portable contract before any provider or runtime side effect.
func (r DesiredStateRequest) Validate(owner string) error {
	if strings.TrimSpace(owner) == "" {
		return &ValidationError{Message: "authenticated owner is required"}
	}
	if !namePattern.MatchString(r.Name) {
		return &ValidationError{Message: "name must be a DNS-compatible lowercase label"}
	}
	if !digestPattern.MatchString(r.CompositionDigest) {
		return &ValidationError{Message: "composition_digest must be a sha256 digest"}
	}
	if strings.TrimSpace(r.ArtifactRef) == "" {
		return &ValidationError{Message: "artifact_ref is required"}
	}
	if err := validateReference("source", r.Source); err != nil {
		return err
	}
	if err := validateReference("evidence", r.Evidence); err != nil {
		return err
	}
	if !supportedBackend(r.ExecutionBackend) {
		return &ValidationError{Message: fmt.Sprintf("unsupported execution_backend %q", r.ExecutionBackend)}
	}
	return validatePlacement(r.Placement)
}

func validateReference(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return &ValidationError{Message: fmt.Sprintf("%s is required", field)}
	}
	if len(value) > 1024 {
		return &ValidationError{Message: fmt.Sprintf("%s is too long", field)}
	}
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return &ValidationError{Message: fmt.Sprintf("%s contains a control character", field)}
		}
	}
	return nil
}

func supportedBackend(backend string) bool {
	switch backend {
	case "nanovms", "podman", "apple-containers", "wsl-containers":
		return true
	default:
		return false
	}
}

func validatePlacement(p Placement) error {
	for field, value := range map[string]string{"region": p.Region, "zone": p.Zone, "node_pool": p.NodePool} {
		if len(value) > 128 {
			return &ValidationError{Message: fmt.Sprintf("placement.%s is too long", field)}
		}
	}
	for kind, values := range map[string]map[string]string{"labels": p.Labels, "constraints": p.Constraints} {
		if len(values) > 32 {
			return &ValidationError{Message: fmt.Sprintf("placement.%s has too many entries", kind)}
		}
		for key, value := range values {
			if strings.TrimSpace(key) == "" || len(key) > 128 || len(value) > 256 {
				return &ValidationError{Message: fmt.Sprintf("placement.%s contains an invalid entry", kind)}
			}
		}
	}
	return nil
}
