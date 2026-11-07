package models

import "context"

// Mature application-operation interfaces.
//
// These are additive seams for CLI/Desktop/API convergence. Current Gin/Tauri
// handlers are intentionally not migrated in Tier A.

type ResolveSourceRequest struct {
	Project ProjectID
	Repository string
	Reference string
}

type SourceResolver interface {
	ResolveSource(context.Context, ResolveSourceRequest) (SourceSnapshot, error)
}

type ManifestResolver interface {
	ResolveManifest(context.Context, SourceSnapshot) (ManifestRevision, []Component, error)
}

type ArtifactResolver interface {
	ResolveArtifacts(context.Context, SourceSnapshot, ManifestRevision, []Component) ([]ArtifactResolution, error)
}

type RuntimeCreateRequest struct {
	Operation RuntimeOperation
	Generation DeploymentGeneration
	Artifacts []ArtifactResolution
}

type RuntimeAdapter interface {
	Create(context.Context, RuntimeCreateRequest) ([]ProviderResource, error)
	Observe(context.Context, ProviderResourceID) (Observation, error)
	Stop(context.Context, ProviderResourceID) error
}

type PortfolioProjector interface {
	Project(context.Context, DeploymentGenerationID) (PortfolioProjectionID, error)
}

type PortfolioPublisher interface {
	Publish(context.Context, PortfolioProjectionID, PublicationOperationID) error
}
