package models

// Additive generalized infrastructure domain for BytePort v1.2.
// No existing deployment route consumes these types yet.

type DesiredResourceKind string

const (
	DesiredResourceService DesiredResourceKind = "service"
	DesiredResourceJob DesiredResourceKind = "job"
	DesiredResourceHost DesiredResourceKind = "host"
	DesiredResourceNetwork DesiredResourceKind = "network"
	DesiredResourceStorage DesiredResourceKind = "storage"
	DesiredResourceManaged DesiredResourceKind = "managed_service"
)

type LifecyclePolicy string

const (
	LifecycleObserveOnly LifecyclePolicy = "observe_only"
	LifecycleCreateObserve LifecyclePolicy = "create_observe"
	LifecycleManage LifecyclePolicy = "manage"
	LifecycleOrphanOnRemove LifecyclePolicy = "orphan_on_remove"
	LifecycleDestroyOnExplicitIntent LifecyclePolicy = "destroy_on_explicit_intent"
)

type DesiredResource struct {
	ID string `json:"id"`
	Kind DesiredResourceKind `json:"kind"`
	ConfigDigest string `json:"config_digest"`
	DependsOn []string `json:"depends_on,omitempty"`
	Artifact *BuildArtifactID `json:"artifact_id,omitempty"`
	Target string `json:"target"`
	Lifecycle LifecyclePolicy `json:"lifecycle"`
	ReplaceOnChange bool `json:"replace_on_change,omitempty"`
}

type DesiredResourceGraph struct {
	ID string `json:"id"`
	Manifest ManifestRevisionID `json:"manifest_revision_id"`
	Resources []DesiredResource `json:"resources"`
}

type TargetCapabilities struct {
	TargetID string `json:"target_id"`
	Provider string `json:"provider"`
	BareMetal bool `json:"bare_metal"`
	SupportsObserve bool `json:"supports_observe"`
	SupportsCreate bool `json:"supports_create"`
	SupportsUpdate bool `json:"supports_update"`
	SupportsReplace bool `json:"supports_replace"`
	SupportsDelete bool `json:"supports_delete"`
	Extensions []string `json:"extensions,omitempty"`
}

type ReconciliationAction string

const (
	ReconcileNoop ReconciliationAction = "NOOP"
	ReconcileRead ReconciliationAction = "READ"
	ReconcileCreate ReconciliationAction = "CREATE"
	ReconcileUpdate ReconciliationAction = "UPDATE"
	ReconcileReplace ReconciliationAction = "REPLACE"
	ReconcileDelete ReconciliationAction = "DELETE"
	ReconcileUnknown ReconciliationAction = "UNKNOWN"
)

type PlannedResourceAction struct {
	DesiredResourceID string `json:"desired_resource_id"`
	RealizedResourceID string `json:"realized_resource_id,omitempty"`
	Action ReconciliationAction `json:"action"`
	Reason string `json:"reason"`
}

type ResourcePlan struct {
	ID string `json:"id"`
	DesiredGraphID string `json:"desired_graph_id"`
	TargetID string `json:"target_id"`
	Actions []PlannedResourceAction `json:"actions"`
}

type RealizedResource struct {
	ID string `json:"id"`
	DesiredResourceID string `json:"desired_resource_id"`
	TargetID string `json:"target_id"`
	Provider string `json:"provider"`
	ExternalID string `json:"external_id"`
	Generation DeploymentGenerationID `json:"generation_id"`
}
