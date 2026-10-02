package models

import (
	"context"
	"fmt"
	"testing"
)

type fixtureInfrastructureAdapter struct {
	caps               TargetCapabilities
	applied            []PlannedResourceAction
	deleteSawNilDesired bool
}

func (f *fixtureInfrastructureAdapter) Capabilities(
	_ context.Context,
	targetID string,
) (TargetCapabilities, error) {
	if targetID != f.caps.TargetID {
		return TargetCapabilities{}, fmt.Errorf("unknown target %q", targetID)
	}
	return f.caps, nil
}

func (f *fixtureInfrastructureAdapter) Observe(
	_ context.Context,
	realized RealizedResource,
) (InfrastructureObservation, error) {
	return InfrastructureObservation{
		RealizedResourceID: realized.ID,
		TargetID:           realized.TargetID,
		Provider:           realized.Provider,
		State:              "observed",
		Fresh:              true,
	}, nil
}

func (f *fixtureInfrastructureAdapter) Apply(
	_ context.Context,
	action PlannedResourceAction,
	desired *DesiredResource,
	realized *RealizedResource,
) (InfrastructureApplyResult, error) {
	f.applied = append(f.applied, action)

	switch action.Action {
	case ReconcileCreate:
		if desired == nil || realized != nil {
			return InfrastructureApplyResult{}, fmt.Errorf("invalid create inputs")
		}
		out := RealizedResource{
			ID:                "real-" + desired.ID,
			DesiredResourceID: desired.ID,
			TargetID:          desired.Target,
			Provider:          f.caps.Provider,
			ExternalID:        "ext-" + desired.ID,
		}
		return InfrastructureApplyResult{Outcome: InfrastructureApplyRealized, Realized: &out}, nil

	case ReconcileUpdate, ReconcileReplace:
		if desired == nil || realized == nil {
			return InfrastructureApplyResult{}, fmt.Errorf("invalid update/replace inputs")
		}
		out := *realized
		out.DesiredResourceID = desired.ID
		return InfrastructureApplyResult{Outcome: InfrastructureApplyRealized, Realized: &out}, nil

	case ReconcileDelete:
		if realized == nil {
			return InfrastructureApplyResult{}, fmt.Errorf("delete missing realized resource")
		}
		f.deleteSawNilDesired = desired == nil
		return InfrastructureApplyResult{
			Outcome: InfrastructureApplyRealized,
			ExternalOperation: &ExternalOperationRef{
				Provider:   f.caps.Provider,
				TargetID:   f.caps.TargetID,
				ExternalID: "delete-" + realized.ExternalID,
				LookupKind: "fixture-delete",
			},
		}, nil
	default:
		return InfrastructureApplyResult{}, fmt.Errorf("unexpected apply action %q", action.Action)
	}
}

func fullFixtureCapabilities() TargetCapabilities {
	return TargetCapabilities{
		TargetID:        "target-1",
		Provider:        "fixture",
		SupportsObserve: true,
		SupportsCreate:  true,
		SupportsUpdate:  true,
		SupportsReplace: true,
		SupportsDelete:  true,
	}
}

func TestProviderBackedReconcileExecutesMixedCanonicalPlan(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "graph-1",
		Resources: []DesiredResource{
			{
				ID:           "service",
				Kind:         DesiredResourceService,
				ConfigDigest: "service-v2",
				Target:       "target-1",
				Lifecycle:    LifecycleManage,
			},
			{
				ID:           "database",
				Kind:         DesiredResourceManaged,
				ConfigDigest: "db-v1",
				Target:       "target-1",
				Lifecycle:    LifecycleManage,
			},
		},
	}
	observed := []ObservedResourceState{
		{
			DesiredResourceID:  "service",
			RealizedResourceID: "real-service",
			ConfigDigest:       "service-v1",
			Fresh:              true,
			Lifecycle:          LifecycleManage,
		},
		{
			DesiredResourceID:  "removed-cache",
			RealizedResourceID: "real-cache",
			ConfigDigest:       "cache-v1",
			Fresh:              true,
			Lifecycle:          LifecycleDestroyOnExplicitIntent,
		},
	}
	realized := map[string]RealizedResource{
		"real-service": {
			ID:                "real-service",
			DesiredResourceID: "service",
			TargetID:          "target-1",
			Provider:          "fixture",
			ExternalID:        "ext-service",
		},
		"real-cache": {
			ID:                "real-cache",
			DesiredResourceID: "removed-cache",
			TargetID:          "target-1",
			Provider:          "fixture",
			ExternalID:        "ext-cache",
		},
	}
	policy := ReconciliationPolicy{
		DestructionIntents: []DestructionIntent{{
			ID:                 "destroy-cache",
			DesiredResourceID:  "removed-cache",
			RealizedResourceID: "real-cache",
			AuthorizedBy:       "fixture-authority",
			Reason:             "removed from desired graph",
		}},
	}
	adapter := &fixtureInfrastructureAdapter{caps: fullFixtureCapabilities()}

	receipt, err := ReconcileOnceWithTargetAdapter(
		context.Background(),
		graph,
		observed,
		realized,
		policy,
		nil,
		"target-1",
		adapter,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(receipt.Plan.Actions) != 3 {
		t.Fatalf("actions=%v", receipt.Plan.Actions)
	}
	if receipt.Plan.Actions[0].Action != ReconcileUpdate {
		t.Fatalf("first action=%v", receipt.Plan.Actions[0])
	}
	if receipt.Plan.Actions[1].Action != ReconcileCreate {
		t.Fatalf("second action=%v", receipt.Plan.Actions[1])
	}
	if receipt.Plan.Actions[2].Action != ReconcileDelete {
		t.Fatalf("third action=%v", receipt.Plan.Actions[2])
	}
	if len(adapter.applied) != 3 {
		t.Fatalf("provider mutations=%d want 3", len(adapter.applied))
	}
	if !adapter.deleteSawNilDesired {
		t.Fatal("orphan DELETE fabricated desired state instead of using exact realized identity")
	}
}

func TestProviderBackedReconcileUnknownInterruptionDoesNotMutate(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "graph-1",
		Resources: []DesiredResource{{
			ID:           "database",
			Kind:         DesiredResourceManaged,
			ConfigDigest: "db-v1",
			Target:       "target-1",
			Lifecycle:    LifecycleManage,
		}},
	}
	adapter := &fixtureInfrastructureAdapter{caps: fullFixtureCapabilities()}
	receipt, err := ReconcileOnceWithTargetAdapter(
		context.Background(),
		graph,
		nil,
		map[string]RealizedResource{},
		ReconciliationPolicy{},
		[]InterruptedResourceOperation{{
			OperationID:       "op-1",
			DesiredResourceID: "database",
			TargetID:          "target-1",
			Provider:          "fixture",
			Action:            ReconcileCreate,
			State:             RuntimeOperationUnknown,
		}},
		"target-1",
		adapter,
	)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Plan.Actions[0].Action != ReconcileUnknown {
		t.Fatalf("action=%v", receipt.Plan.Actions[0])
	}
	if len(adapter.applied) != 0 {
		t.Fatalf("UNKNOWN caused %d provider mutations", len(adapter.applied))
	}
}

func TestProviderBackedReconcileUnsupportedReplaceDoesNotMutate(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "graph-1",
		Resources: []DesiredResource{{
			ID:              "host",
			Kind:            DesiredResourceHost,
			ConfigDigest:    "host-v2",
			Target:          "target-1",
			Lifecycle:       LifecycleManage,
			ReplaceOnChange: true,
		}},
	}
	observed := []ObservedResourceState{{
		DesiredResourceID:  "host",
		RealizedResourceID: "real-host",
		ConfigDigest:       "host-v1",
		Fresh:              true,
		Lifecycle:          LifecycleManage,
	}}
	realized := map[string]RealizedResource{
		"real-host": {
			ID:                "real-host",
			DesiredResourceID: "host",
			TargetID:          "target-1",
			Provider:          "fixture",
			ExternalID:        "ext-host",
		},
	}
	caps := fullFixtureCapabilities()
	caps.SupportsReplace = false
	adapter := &fixtureInfrastructureAdapter{caps: caps}

	receipt, err := ReconcileOnceWithTargetAdapter(
		context.Background(),
		graph,
		observed,
		realized,
		ReconciliationPolicy{},
		nil,
		"target-1",
		adapter,
	)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Plan.Actions[0].Action != ReconcileUnknown {
		t.Fatalf("unsupported replace was not UNKNOWN: %v", receipt.Plan.Actions[0])
	}
	if len(adapter.applied) != 0 {
		t.Fatalf("unsupported replace caused %d provider mutations", len(adapter.applied))
	}
}

func TestProviderBackedReconcileRejectsMixedTargetGraph(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "graph-1",
		Resources: []DesiredResource{
			{
				ID: "a", Kind: DesiredResourceManaged, ConfigDigest: "a",
				Target: "target-1", Lifecycle: LifecycleManage,
			},
			{
				ID: "b", Kind: DesiredResourceManaged, ConfigDigest: "b",
				Target: "target-2", Lifecycle: LifecycleManage,
			},
		},
	}
	adapter := &fixtureInfrastructureAdapter{caps: fullFixtureCapabilities()}
	_, err := ReconcileOnceWithTargetAdapter(
		context.Background(),
		graph,
		nil,
		map[string]RealizedResource{},
		ReconciliationPolicy{},
		nil,
		"target-1",
		adapter,
	)
	if err == nil {
		t.Fatal("mixed-target graph silently executed through one target adapter")
	}
	if len(adapter.applied) != 0 {
		t.Fatalf("mixed-target rejection occurred after %d mutations", len(adapter.applied))
	}
}


func TestProviderBackedReconcileRejectsWrongTargetRealizedBeforeDelete(t *testing.T) {
	graph := DesiredResourceGraph{ID: "graph-1"}
	observed := []ObservedResourceState{{
		DesiredResourceID:  "removed-cache",
		RealizedResourceID: "real-cache",
		ConfigDigest:       "cache-v1",
		Fresh:              true,
		Lifecycle:          LifecycleDestroyOnExplicitIntent,
	}}
	realized := map[string]RealizedResource{
		"real-cache": {
			ID:                "real-cache",
			DesiredResourceID: "removed-cache",
			TargetID:          "other-target",
			Provider:          "fixture",
			ExternalID:        "ext-cache",
		},
	}
	policy := ReconciliationPolicy{
		DestructionIntents: []DestructionIntent{{
			ID:                 "destroy-cache",
			DesiredResourceID:  "removed-cache",
			RealizedResourceID: "real-cache",
			AuthorizedBy:       "fixture-authority",
			Reason:             "removed from desired graph",
		}},
	}
	adapter := &fixtureInfrastructureAdapter{caps: fullFixtureCapabilities()}

	_, err := ReconcileOnceWithTargetAdapter(
		context.Background(),
		graph,
		observed,
		realized,
		policy,
		nil,
		"target-1",
		adapter,
	)
	if err == nil {
		t.Fatal("wrong-target realized resource reached delete path")
	}
	if len(adapter.applied) != 0 {
		t.Fatalf("wrong-target rejection occurred after %d mutations", len(adapter.applied))
	}
}

func TestProviderBackedReconcileRejectsWrongProviderRealizedBeforeUpdate(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "graph-1",
		Resources: []DesiredResource{{
			ID:           "service",
			Kind:         DesiredResourceService,
			ConfigDigest: "service-v2",
			Target:       "target-1",
			Lifecycle:    LifecycleManage,
		}},
	}
	observed := []ObservedResourceState{{
		DesiredResourceID:  "service",
		RealizedResourceID: "real-service",
		ConfigDigest:       "service-v1",
		Fresh:              true,
		Lifecycle:          LifecycleManage,
	}}
	realized := map[string]RealizedResource{
		"real-service": {
			ID:                "real-service",
			DesiredResourceID: "service",
			TargetID:          "target-1",
			Provider:          "other-provider",
			ExternalID:        "ext-service",
		},
	}
	adapter := &fixtureInfrastructureAdapter{caps: fullFixtureCapabilities()}

	_, err := ReconcileOnceWithTargetAdapter(
		context.Background(),
		graph,
		observed,
		realized,
		ReconciliationPolicy{},
		nil,
		"target-1",
		adapter,
	)
	if err == nil {
		t.Fatal("wrong-provider realized resource reached update path")
	}
	if len(adapter.applied) != 0 {
		t.Fatalf("wrong-provider rejection occurred after %d mutations", len(adapter.applied))
	}
}


func TestProviderBackedReconcileRejectsWrongProviderObservationBeforePlanning(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "graph-1",
		Resources: []DesiredResource{{
			ID:           "service",
			Kind:         DesiredResourceService,
			ConfigDigest: "service-v2",
			Target:       "target-1",
			Lifecycle:    LifecycleManage,
		}},
	}
	observed := []ObservedResourceState{{
		DesiredResourceID:  "service",
		RealizedResourceID: "real-service",
		TargetID:           "target-1",
		Provider:           "other-provider",
		ConfigDigest:       "service-v1",
		Fresh:              true,
		Lifecycle:          LifecycleManage,
	}}
	adapter := &fixtureInfrastructureAdapter{caps: fullFixtureCapabilities()}

	_, err := ReconcileOnceWithTargetAdapter(
		context.Background(),
		graph,
		observed,
		map[string]RealizedResource{},
		ReconciliationPolicy{},
		nil,
		"target-1",
		adapter,
	)
	if err == nil {
		t.Fatal("wrong-provider observation entered planning")
	}
	if len(adapter.applied) != 0 {
		t.Fatalf("wrong-provider observation caused %d mutations", len(adapter.applied))
	}
}

func TestProviderBackedReconcileRejectsWrongTargetObservationBeforePlanning(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "graph-1",
		Resources: []DesiredResource{{
			ID:           "service",
			Kind:         DesiredResourceService,
			ConfigDigest: "service-v2",
			Target:       "target-1",
			Lifecycle:    LifecycleManage,
		}},
	}
	observed := []ObservedResourceState{{
		DesiredResourceID:  "service",
		RealizedResourceID: "real-service",
		TargetID:           "other-target",
		Provider:           "fixture",
		ConfigDigest:       "service-v1",
		Fresh:              true,
		Lifecycle:          LifecycleManage,
	}}
	adapter := &fixtureInfrastructureAdapter{caps: fullFixtureCapabilities()}

	_, err := ReconcileOnceWithTargetAdapter(
		context.Background(),
		graph,
		observed,
		map[string]RealizedResource{},
		ReconciliationPolicy{},
		nil,
		"target-1",
		adapter,
	)
	if err == nil {
		t.Fatal("wrong-target observation entered planning")
	}
	if len(adapter.applied) != 0 {
		t.Fatalf("wrong-target observation caused %d mutations", len(adapter.applied))
	}
}


type invalidApplyOutcomeAdapter struct {
	caps   TargetCapabilities
	result InfrastructureApplyResult
}

func (a *invalidApplyOutcomeAdapter) Capabilities(
	_ context.Context,
	_ string,
) (TargetCapabilities, error) {
	return a.caps, nil
}

func (a *invalidApplyOutcomeAdapter) Observe(
	_ context.Context,
	_ RealizedResource,
) (InfrastructureObservation, error) {
	return InfrastructureObservation{}, fmt.Errorf("unexpected observe")
}

func (a *invalidApplyOutcomeAdapter) Apply(
	_ context.Context,
	_ PlannedResourceAction,
	_ *DesiredResource,
	_ *RealizedResource,
) (InfrastructureApplyResult, error) {
	return a.result, nil
}

func oneCreateGraph() DesiredResourceGraph {
	return DesiredResourceGraph{
		ID: "g-invalid-outcome",
		Resources: []DesiredResource{{
			ID:           "service",
			Kind:         DesiredResourceService,
			ConfigDigest: "v1",
			Target:       "target-1",
			Lifecycle:    LifecycleManage,
		}},
	}
}

func TestProviderBackedReconcileRejectsUnknownWithFabricatedRealizedState(t *testing.T) {
	adapter := &invalidApplyOutcomeAdapter{
		caps: fullFixtureCapabilities(),
		result: InfrastructureApplyResult{
			Outcome: InfrastructureApplyUnknown,
			Realized: &RealizedResource{
				ID:                "fabricated",
				DesiredResourceID: "service",
				TargetID:          "target-1",
				Provider:          "fixture",
				ExternalID:        "ext",
			},
			ExternalOperation: &ExternalOperationRef{
				Provider: "fixture", TargetID: "target-1",
				ExternalID: "op-1", LookupKind: "operation",
			},
		},
	}
	_, err := ReconcileOnceWithTargetAdapterOperation(
		context.Background(),
		"runtime-op",
		oneCreateGraph(),
		nil,
		map[string]RealizedResource{},
		ReconciliationPolicy{},
		nil,
		"target-1",
		adapter,
	)
	if err == nil {
		t.Fatal("UNKNOWN outcome with realized state was accepted")
	}
}

func TestProviderBackedReconcileRejectsUnknownWithoutReconciliationIdentity(t *testing.T) {
	adapter := &invalidApplyOutcomeAdapter{
		caps: fullFixtureCapabilities(),
		result: InfrastructureApplyResult{
			Outcome: InfrastructureApplyUnknown,
		},
	}
	_, err := ReconcileOnceWithTargetAdapterOperation(
		context.Background(),
		"runtime-op",
		oneCreateGraph(),
		nil,
		map[string]RealizedResource{},
		ReconciliationPolicy{},
		nil,
		"target-1",
		adapter,
	)
	if err == nil {
		t.Fatal("UNKNOWN outcome without external operation identity was accepted")
	}
}

func TestProviderBackedReconcileRejectsMissingApplyOutcome(t *testing.T) {
	adapter := &invalidApplyOutcomeAdapter{
		caps: fullFixtureCapabilities(),
		result: InfrastructureApplyResult{
			Realized: &RealizedResource{
				ID:                "real-service",
				DesiredResourceID: "service",
				TargetID:          "target-1",
				Provider:          "fixture",
				ExternalID:        "ext-service",
			},
		},
	}
	_, err := ReconcileOnceWithTargetAdapterOperation(
		context.Background(),
		"runtime-op",
		oneCreateGraph(),
		nil,
		map[string]RealizedResource{},
		ReconciliationPolicy{},
		nil,
		"target-1",
		adapter,
	)
	if err == nil {
		t.Fatal("missing apply outcome was accepted")
	}
}
