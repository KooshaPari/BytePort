package models

import "context"

type InfrastructureObservation struct {
	RealizedResourceID string `json:"realized_resource_id"`
	TargetID string `json:"target_id"`
	ConfigDigest string `json:"config_digest,omitempty"`
	State string `json:"state"`
	Fresh bool `json:"fresh"`
	RawRef string `json:"raw_ref,omitempty"`
}

type InfrastructureApplyResult struct {
	Realized *RealizedResource `json:"realized_resource,omitempty"`
	ExternalOperation *ExternalOperationRef `json:"external_operation,omitempty"`
	Observation *InfrastructureObservation `json:"observation,omitempty"`
}

type InfrastructureTargetAdapter interface {
	Capabilities(context.Context, string) (TargetCapabilities, error)
	Observe(context.Context, RealizedResource) (InfrastructureObservation, error)
	Apply(context.Context, PlannedResourceAction, *DesiredResource, *RealizedResource) (InfrastructureApplyResult, error)
}

type ExternalOperationRef struct {
	Provider string `json:"provider"`
	TargetID string `json:"target_id"`
	ExternalID string `json:"external_id"`
	LookupKind string `json:"lookup_kind"`
}
