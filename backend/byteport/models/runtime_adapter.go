package models

import "context"

type RuntimeAdapterCapabilities struct {
	AdapterID string `json:"adapter_id"`
	Version string `json:"version"`
	SupportsObserve bool `json:"supports_observe"`
	SupportsFindByOperation bool `json:"supports_find_by_operation"`
	SupportsDelayedVisibility bool `json:"supports_delayed_visibility"`
}

type RuntimeObservationResult struct {
	Observation Observation `json:"observation"`
	Found bool `json:"found"`
	VisibilityUncertain bool `json:"visibility_uncertain"`
}

type MatureRuntimeAdapter interface {
	Capabilities(context.Context) RuntimeAdapterCapabilities
	Create(context.Context, RuntimeCreateRequest) ([]ProviderResource, error)
	Observe(context.Context, ProviderResourceID) (RuntimeObservationResult, error)
	FindByOperation(context.Context, RuntimeOperationID) ([]ProviderResource, error)
	Stop(context.Context, ProviderResourceID) error
}
