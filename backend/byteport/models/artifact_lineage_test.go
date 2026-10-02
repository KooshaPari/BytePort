package models

import (
	"context"
	"testing"
)

type lineageArtifactResolver struct {
	artifact BuildArtifact
	found    bool
}

func (r lineageArtifactResolver) ResolveBuildArtifact(
	_ context.Context,
	_ BuildArtifactID,
) (BuildArtifact, bool, error) {
	return r.artifact, r.found, nil
}

func artifactIDPtr(id string) *BuildArtifactID {
	value := BuildArtifactID(id)
	return &value
}

func TestDesiredGraphArtifactLineageAcceptsExactManifestArtifact(t *testing.T) {
	graph := DesiredResourceGraph{
		ID:       "g",
		Manifest: "manifest-1",
		Resources: []DesiredResource{{
			ID: "service", Kind: DesiredResourceService, Target: "target-1",
			ConfigDigest: "cfg", Lifecycle: LifecycleManage,
			Artifact: artifactIDPtr("artifact-1"),
		}},
	}
	resolver := lineageArtifactResolver{
		found: true,
		artifact: BuildArtifact{
			ID:               "artifact-1",
			ImmutableRef:     "sha256:abc",
			ManifestRevision: "manifest-1",
		},
	}
	if err := ValidateDesiredGraphArtifactLineage(context.Background(), graph, resolver); err != nil {
		t.Fatal(err)
	}
}

func TestDesiredGraphArtifactLineageRejectsWrongManifestBeforeProvider(t *testing.T) {
	graph := DesiredResourceGraph{
		ID:       "g",
		Manifest: "manifest-current",
		Resources: []DesiredResource{{
			ID: "service", Kind: DesiredResourceService, Target: "target-1",
			ConfigDigest: "cfg", Lifecycle: LifecycleManage,
			Artifact: artifactIDPtr("artifact-old"),
		}},
	}
	resolver := lineageArtifactResolver{
		found: true,
		artifact: BuildArtifact{
			ID:               "artifact-old",
			ImmutableRef:     "sha256:old",
			ManifestRevision: "manifest-old",
		},
	}
	if err := ValidateDesiredGraphArtifactLineage(context.Background(), graph, resolver); err == nil {
		t.Fatal("wrong-manifest artifact was accepted")
	}
}

func TestDesiredGraphArtifactLineageRejectsResolverIdentitySubstitution(t *testing.T) {
	graph := DesiredResourceGraph{
		ID:       "g",
		Manifest: "manifest-1",
		Resources: []DesiredResource{{
			ID: "service", Kind: DesiredResourceService, Target: "target-1",
			ConfigDigest: "cfg", Lifecycle: LifecycleManage,
			Artifact: artifactIDPtr("artifact-requested"),
		}},
	}
	resolver := lineageArtifactResolver{
		found: true,
		artifact: BuildArtifact{
			ID:               "artifact-other",
			ImmutableRef:     "sha256:other",
			ManifestRevision: "manifest-1",
		},
	}
	if err := ValidateDesiredGraphArtifactLineage(context.Background(), graph, resolver); err == nil {
		t.Fatal("resolver identity substitution was accepted")
	}
}

func TestDesiredGraphArtifactLineageRejectsMutableOrMissingArtifactReference(t *testing.T) {
	graph := DesiredResourceGraph{
		ID:       "g",
		Manifest: "manifest-1",
		Resources: []DesiredResource{{
			ID: "service", Kind: DesiredResourceService, Target: "target-1",
			ConfigDigest: "cfg", Lifecycle: LifecycleManage,
			Artifact: artifactIDPtr("artifact-1"),
		}},
	}
	resolver := lineageArtifactResolver{
		found: true,
		artifact: BuildArtifact{
			ID:               "artifact-1",
			ManifestRevision: "manifest-1",
		},
	}
	if err := ValidateDesiredGraphArtifactLineage(context.Background(), graph, resolver); err == nil {
		t.Fatal("artifact without immutable reference was accepted")
	}
}
