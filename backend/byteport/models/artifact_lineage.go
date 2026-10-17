package models

import (
	"context"
	"fmt"
)

// ValidateDesiredGraphArtifactLineage verifies that every artifact-bearing
// desired resource resolves to the exact artifact identity and manifest
// revision authorized by the graph before any provider mutation is attempted.
//
// This is intentionally separate from graph construction: a graph may exist
// before build resolution, but an execution path that requires artifacts must
// prove lineage first.
func ValidateDesiredGraphArtifactLineage(
	ctx context.Context,
	graph DesiredResourceGraph,
	resolver BuildArtifactResolver,
) error {
	if resolver == nil {
		return fmt.Errorf("build artifact resolver is required")
	}

	for _, resource := range graph.Resources {
		if resource.Artifact == nil {
			continue
		}
		if *resource.Artifact == "" {
			return fmt.Errorf("desired resource %q has empty artifact ID", resource.ID)
		}

		artifact, found, err := resolver.ResolveBuildArtifact(ctx, *resource.Artifact)
		if err != nil {
			return fmt.Errorf("resolve artifact %q for %q: %w", *resource.Artifact, resource.ID, err)
		}
		if !found {
			return fmt.Errorf("artifact %q for %q was not found", *resource.Artifact, resource.ID)
		}
		if artifact.ID != *resource.Artifact {
			return fmt.Errorf(
				"artifact resolver returned identity %q for requested %q",
				artifact.ID,
				*resource.Artifact,
			)
		}
		if artifact.ImmutableRef == "" {
			return fmt.Errorf("artifact %q has no immutable reference", artifact.ID)
		}
		if graph.Manifest != "" && artifact.ManifestRevision != graph.Manifest {
			return fmt.Errorf(
				"artifact %q belongs to manifest %q, graph requires %q",
				artifact.ID,
				artifact.ManifestRevision,
				graph.Manifest,
			)
		}
	}

	return nil
}
