package models

import "context"

type BuildEngineCapabilities struct {
	EngineID string `json:"engine_id"`
	Version string `json:"version"`
	SupportsExplicitBuild bool `json:"supports_explicit_build"`
	SupportsDetection bool `json:"supports_detection"`
	SupportsPrebuilt bool `json:"supports_prebuilt"`
	Platforms []string `json:"platforms,omitempty"`
	Provenance bool `json:"provenance"`
	SBOM bool `json:"sbom"`
}

type BuildRequest struct {
	Operation BuildOperationID `json:"operation_id"`
	Source SourceSnapshotID `json:"source_snapshot_id"`
	Manifest ManifestRevisionID `json:"manifest_revision_id"`
	Component ComponentID `json:"component_id"`
	Fingerprint string `json:"fingerprint"`
	Platform string `json:"platform"`
}

type BuildResult struct {
	Artifact BuildArtifact `json:"artifact"`
	RawReceiptRef string `json:"raw_receipt_ref"`
}

type BuildEngine interface {
	Capabilities(context.Context) BuildEngineCapabilities
	Build(context.Context, BuildRequest) (BuildResult, error)
}
