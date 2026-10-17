package models

import "time"

type ProjectID string
type SourceSnapshotID string
type ManifestRevisionID string
type ComponentID string
type BuildOperationID string
type BuildArtifactID string
type DeploymentIntentID string
type RuntimeOperationID string
type DeploymentGenerationID string
type ProviderResourceID string
type ObservationID string
type PortfolioProjectionID string
type PublicationOperationID string
type EvidenceReceiptID string

type ArtifactResolutionKind string

const (
	ArtifactResolutionPrebuilt ArtifactResolutionKind = "prebuilt"
	ArtifactResolutionBuilt ArtifactResolutionKind = "built"
	ArtifactResolutionManaged ArtifactResolutionKind = "provider_managed"
)

type SourceSnapshot struct {
	ID SourceSnapshotID `json:"id"`
	Provider string `json:"provider"`
	Repository string `json:"repository"`
	Commit string `json:"commit"`
	ContentDigest string `json:"content_digest,omitempty"`
	ResolvedAt time.Time `json:"resolved_at"`
}

type ManifestRevision struct {
	ID ManifestRevisionID `json:"id"`
	SourceSnapshot SourceSnapshotID `json:"source_snapshot_id"`
	Path string `json:"path"`
	SchemaVersion string `json:"schema_version"`
	ContentDigest string `json:"content_digest"`
}

type Component struct {
	ID ComponentID `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type ArtifactResolution struct {
	Component ComponentID `json:"component_id"`
	Kind ArtifactResolutionKind `json:"kind"`
	Artifact *BuildArtifactID `json:"artifact_id,omitempty"`
}

type BuildArtifact struct {
	ID BuildArtifactID `json:"id"`
	ImmutableRef string `json:"immutable_ref"`
	MediaKind string `json:"media_kind"`
	Platforms []string `json:"platforms,omitempty"`
	SourceSnapshot SourceSnapshotID `json:"source_snapshot_id"`
	ManifestRevision ManifestRevisionID `json:"manifest_revision_id"`
	BuildOperation BuildOperationID `json:"build_operation_id"`
	Engine string `json:"engine"`
	EngineVersion string `json:"engine_version"`
	ProvenanceRef string `json:"provenance_ref,omitempty"`
	SBOMRef string `json:"sbom_ref,omitempty"`
}

type RuntimeOperationState string

const (
	RuntimeOperationPlanned RuntimeOperationState = "PLANNED"
	RuntimeOperationApplying RuntimeOperationState = "APPLYING"
	RuntimeOperationUnknown RuntimeOperationState = "UNKNOWN"
	RuntimeOperationReconciling RuntimeOperationState = "RECONCILING"
	RuntimeOperationRealized RuntimeOperationState = "REALIZED"
	RuntimeOperationFailed RuntimeOperationState = "FAILED"
)

type RuntimeOperation struct {
	ID RuntimeOperationID `json:"id"`
	Fingerprint string `json:"fingerprint"`
	State RuntimeOperationState `json:"state"`
	Target string `json:"target"`
	CreatedAt time.Time `json:"created_at"`
}

type DeploymentGeneration struct {
	ID DeploymentGenerationID `json:"id"`
	Intent DeploymentIntentID `json:"deployment_intent_id"`
	Predecessor *DeploymentGenerationID `json:"predecessor_id,omitempty"`
	RollbackOf *DeploymentGenerationID `json:"rollback_of_id,omitempty"`
	Role string `json:"role"`
}

type ProviderResource struct {
	ID ProviderResourceID `json:"id"`
	Operation RuntimeOperationID `json:"runtime_operation_id"`
	Generation DeploymentGenerationID `json:"deployment_generation_id"`
	Provider string `json:"provider"`
	Target string `json:"target"`
	ExternalID string `json:"external_id"`
}

type Observation struct {
	ID ObservationID `json:"id"`
	ProviderResource ProviderResourceID `json:"provider_resource_id"`
	ObservedAt time.Time `json:"observed_at"`
	FreshUntil *time.Time `json:"fresh_until,omitempty"`
	State string `json:"state"`
	RawRef string `json:"raw_ref,omitempty"`
}

type LegacyVerificationState string

const (
	LegacyVerificationUnverified LegacyVerificationState = "legacy_unverified"
)

type LegacyDeploymentImport struct {
	ProjectID ProjectID `json:"project_id"`
	Provider string `json:"provider"`
	ExternalID string `json:"external_id"`
	Verification LegacyVerificationState `json:"verification"`
	SourceSnapshot *SourceSnapshotID `json:"source_snapshot_id,omitempty"`
	ManifestRevision *ManifestRevisionID `json:"manifest_revision_id,omitempty"`
	BuildArtifact *BuildArtifactID `json:"build_artifact_id,omitempty"`
}
