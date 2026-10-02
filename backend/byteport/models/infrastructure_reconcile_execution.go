package models

import (
	"context"
	"fmt"
)

// InfrastructureActionExecution records exactly what the provider-backed
// reconciliation experiment did for one planned action.
type InfrastructureActionExecution struct {
	Action      PlannedResourceAction
	Applied     bool
	Observed    *InfrastructureObservation
	ApplyResult *InfrastructureApplyResult
}

// InfrastructureExecutionReceipt binds the executed actions to the exact plan
// and target capability snapshot used for the run.
type InfrastructureExecutionReceipt struct {
	Plan         ResourcePlan
	Capabilities TargetCapabilities
	Executions   []InfrastructureActionExecution
}

// ReconcileOnceWithTargetAdapter is the operation-neutral fixture entrypoint.
// It is retained for pure/provider-fixture qualification where no durable
// mutation journal is being asserted.
func ReconcileOnceWithTargetAdapter(
	ctx context.Context,
	graph DesiredResourceGraph,
	observed []ObservedResourceState,
	realized map[string]RealizedResource,
	policy ReconciliationPolicy,
	interruptions []InterruptedResourceOperation,
	targetID string,
	adapter InfrastructureTargetAdapter,
) (InfrastructureExecutionReceipt, error) {
	return reconcileOnceWithTargetAdapter(
		ctx, "", graph, observed, realized, policy, interruptions, targetID, adapter,
	)
}

// ReconcileOnceWithTargetAdapterOperation binds one explicit durable runtime
// operation identity to every provider mutation in this reconciliation pass.
// Pure planning never invents this identity.
func ReconcileOnceWithTargetAdapterOperation(
	ctx context.Context,
	operationID RuntimeOperationID,
	graph DesiredResourceGraph,
	observed []ObservedResourceState,
	realized map[string]RealizedResource,
	policy ReconciliationPolicy,
	interruptions []InterruptedResourceOperation,
	targetID string,
	adapter InfrastructureTargetAdapter,
) (InfrastructureExecutionReceipt, error) {
	if operationID == "" {
		return InfrastructureExecutionReceipt{}, fmt.Errorf("runtime operation ID is required")
	}
	return reconcileOnceWithTargetAdapter(
		ctx, operationID, graph, observed, realized, policy, interruptions, targetID, adapter,
	)
}

func reconcileOnceWithTargetAdapter(
	ctx context.Context,
	operationID RuntimeOperationID,
	graph DesiredResourceGraph,
	observed []ObservedResourceState,
	realized map[string]RealizedResource,
	policy ReconciliationPolicy,
	interruptions []InterruptedResourceOperation,
	targetID string,
	adapter InfrastructureTargetAdapter,
) (InfrastructureExecutionReceipt, error) {
	if targetID == "" {
		return InfrastructureExecutionReceipt{}, fmt.Errorf("target ID is required")
	}

	caps, err := adapter.Capabilities(ctx, targetID)
	if err != nil {
		return InfrastructureExecutionReceipt{}, fmt.Errorf("target capabilities: %w", err)
	}
	if caps.TargetID != targetID {
		return InfrastructureExecutionReceipt{}, fmt.Errorf(
			"adapter returned capabilities for target %q, want %q",
			caps.TargetID,
			targetID,
		)
	}

	for _, desired := range graph.Resources {
		if desired.Target != targetID {
			return InfrastructureExecutionReceipt{}, fmt.Errorf(
				"desired resource %q targets %q, not execution target %q",
				desired.ID,
				desired.Target,
				targetID,
			)
		}
	}

	for _, observation := range observed {
		if observation.TargetID != "" && observation.TargetID != targetID {
			return InfrastructureExecutionReceipt{}, fmt.Errorf(
				"observation for %q targets %q, not execution target %q",
				observation.DesiredResourceID,
				observation.TargetID,
				targetID,
			)
		}
		if caps.Provider != "" && observation.Provider != "" &&
			observation.Provider != caps.Provider {
			return InfrastructureExecutionReceipt{}, fmt.Errorf(
				"observation for %q uses provider %q, want %q",
				observation.DesiredResourceID,
				observation.Provider,
				caps.Provider,
			)
		}
	}

	plan, err := PlanReconciliationWithInterruptions(graph, observed, policy, interruptions)
	if err != nil {
		return InfrastructureExecutionReceipt{}, err
	}
	plan = ConstrainPlanToTarget(plan, caps)

	if operationID != "" {
		for i := range plan.Actions {
			switch plan.Actions[i].Action {
			case ReconcileCreate, ReconcileUpdate, ReconcileReplace, ReconcileDelete:
				plan.Actions[i].OperationID = operationID
			}
		}
	}

	desiredByID := make(map[string]DesiredResource, len(graph.Resources))
	for _, desired := range graph.Resources {
		desiredByID[desired.ID] = desired
	}

	receipt := InfrastructureExecutionReceipt{
		Plan:         plan,
		Capabilities: caps,
		Executions:   make([]InfrastructureActionExecution, 0, len(plan.Actions)),
	}

	for _, action := range plan.Actions {
		execution := InfrastructureActionExecution{Action: action}

		switch action.Action {
		case ReconcileUnknown, ReconcileNoop:
			receipt.Executions = append(receipt.Executions, execution)
			continue

		case ReconcileRead:
			current, ok := realized[action.RealizedResourceID]
			if !ok {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"READ %q missing exact realized resource %q",
					action.DesiredResourceID,
					action.RealizedResourceID,
				)
			}
			observation, err := adapter.Observe(ctx, current)
			if err != nil {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"observe %q: %w",
					action.RealizedResourceID,
					err,
				)
			}
			if observation.RealizedResourceID != current.ID ||
				observation.TargetID != targetID ||
				(caps.Provider != "" && observation.Provider != "" &&
					observation.Provider != caps.Provider) {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"observation identity mismatch for realized resource %q",
					current.ID,
				)
			}
			execution.Observed = &observation
			receipt.Executions = append(receipt.Executions, execution)
			continue
		}

		var desired *DesiredResource
		if value, ok := desiredByID[action.DesiredResourceID]; ok {
			copy := value
			desired = &copy
		}

		var current *RealizedResource
		if action.RealizedResourceID != "" {
			value, ok := realized[action.RealizedResourceID]
			if !ok {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"%s %q missing exact realized resource %q",
					action.Action,
					action.DesiredResourceID,
					action.RealizedResourceID,
				)
			}
			if value.TargetID != targetID {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"%s %q realized target mismatch: got %q want %q",
					action.Action,
					action.DesiredResourceID,
					value.TargetID,
					targetID,
				)
			}
			if caps.Provider != "" && value.Provider != "" && value.Provider != caps.Provider {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"%s %q realized provider mismatch: got %q want %q",
					action.Action,
					action.DesiredResourceID,
					value.Provider,
					caps.Provider,
				)
			}
			copy := value
			current = &copy
		}

		switch action.Action {
		case ReconcileCreate:
			if desired == nil || current != nil {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"CREATE %q requires desired state and no realized resource",
					action.DesiredResourceID,
				)
			}
		case ReconcileUpdate, ReconcileReplace:
			if desired == nil || current == nil {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"%s %q requires desired and exact realized state",
					action.Action,
					action.DesiredResourceID,
				)
			}
		case ReconcileDelete:
			if current == nil {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"DELETE %q requires exact realized state",
					action.DesiredResourceID,
				)
			}
			// desired may intentionally be nil: an authorized deletion often
			// represents a resource removed from the desired graph.
		default:
			return InfrastructureExecutionReceipt{}, fmt.Errorf(
				"unsupported reconciliation action %q",
				action.Action,
			)
		}

		result, err := adapter.Apply(ctx, action, desired, current)
		if err != nil {
			return InfrastructureExecutionReceipt{}, fmt.Errorf(
				"apply %s %q: %w",
				action.Action,
				action.DesiredResourceID,
				err,
			)
		}
		if result.ExternalOperation != nil {
			if result.ExternalOperation.TargetID != targetID {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"provider operation target mismatch: got %q want %q",
					result.ExternalOperation.TargetID,
					targetID,
				)
			}
			if caps.Provider != "" &&
				result.ExternalOperation.Provider != "" &&
				result.ExternalOperation.Provider != caps.Provider {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"provider operation identity mismatch: got %q want %q",
					result.ExternalOperation.Provider,
					caps.Provider,
				)
			}
		}
		if result.Realized != nil {
			if result.Realized.TargetID != targetID {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"realized target mismatch: got %q want %q",
					result.Realized.TargetID,
					targetID,
				)
			}
			if desired != nil && result.Realized.DesiredResourceID != desired.ID {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"realized desired-resource mismatch: got %q want %q",
					result.Realized.DesiredResourceID,
					desired.ID,
				)
			}
		}

		execution.Applied = true
		execution.ApplyResult = &result
		receipt.Executions = append(receipt.Executions, execution)
	}

	return receipt, nil
}
