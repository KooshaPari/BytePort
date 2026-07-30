// Package mesh contains BytePort compute-mesh desired-state contracts.
package mesh

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const sha256Prefix = "sha256:"

// WorkloadIntent is an owner-scoped desired-state request for the compute mesh.
// Provider credentials and provider-specific state stay behind adapters.
type WorkloadIntent struct {
	Owner             string            `json:"owner"`
	CompositionName   string            `json:"composition_name"`
	CompositionDigest string            `json:"composition_digest"`
	ArtifactRef       string            `json:"artifact_ref"`
	ExecutionBackend  string            `json:"execution_backend"`
	Placement         map[string]string `json:"placement,omitempty"`
}

// Validate enforces immutable composition identity and a known runtime backend.
func (i WorkloadIntent) Validate() error {
	if strings.TrimSpace(i.Owner) == "" {
		return fmt.Errorf("owner is required")
	}
	if strings.TrimSpace(i.CompositionName) == "" {
		return fmt.Errorf("composition_name is required")
	}
	if !strings.HasPrefix(i.CompositionDigest, sha256Prefix) || len(i.CompositionDigest) != len(sha256Prefix)+64 {
		return fmt.Errorf("composition_digest must be sha256:<64 hex>")
	}
	if _, err := hex.DecodeString(strings.TrimPrefix(i.CompositionDigest, sha256Prefix)); err != nil {
		return fmt.Errorf("composition_digest must be sha256:<64 hex>")
	}
	if strings.TrimSpace(i.ArtifactRef) == "" {
		return fmt.Errorf("artifact_ref is required")
	}
	switch i.ExecutionBackend {
	case "nanovms", "podman", "apple-containers", "wsl-containers":
	default:
		return fmt.Errorf("unsupported execution_backend %q", i.ExecutionBackend)
	}
	return nil
}
