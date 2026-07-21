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
	Name              string    `json:"name" binding:"required"`
	CompositionDigest string    `json:"composition_digest" binding:"required"`
	ArtifactRef       string    `json:"artifact_ref" binding:"required"`
	ExecutionBackend  string    `json:"execution_backend" binding:"required"`
	Placement         Placement `json:"placement"`
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
	Name              string    `json:"name"`
	Owner             string    `json:"owner"`
	CompositionDigest string    `json:"composition_digest"`
	ArtifactRef       string    `json:"artifact_ref"`
	ExecutionBackend  string    `json:"execution_backend"`
	Placement         Placement `json:"placement"`
	Status            string    `json:"status"`
	AcceptedAt        time.Time `json:"accepted_at"`
}

// ValidationError describes a malformed desired-state request.
type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }

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
	if err := validateArtifactRef(r.ArtifactRef); err != nil {
		return err
	}
	if !supportedBackend(r.ExecutionBackend) {
		return &ValidationError{Message: fmt.Sprintf("unsupported execution_backend %q", r.ExecutionBackend)}
	}
	return validatePlacement(r.Placement)
}

func validateArtifactRef(value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Message: "artifact_ref is required"}
	}
	if len(value) > 512 || strings.IndexFunc(value, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0 {
		return &ValidationError{Message: "artifact_ref must be at most 512 characters and contain no control characters"}
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
