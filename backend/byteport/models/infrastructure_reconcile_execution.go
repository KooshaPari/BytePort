package models

import (
	"context"
	"crypto/sha256"
	"fmt"
)

// InfrastructureActionExecution records exactly what the provider-backed
// reconciliation experiment did for one planned action.
type InfrastructureActionExecution struct {
	Action      PlannedResourceAction
	Attempted   bool
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

func deriveActionOperationID(
	root RuntimeOperationID,
	action PlannedResourceAction,
) RuntimeOperationID {
	payload := fmt.Sprintf(
		"%s\x00%s\x00%s\x00%s",
		root,
		action.Action,
		action.DesiredResourceID,
		action.RealizedResourceID,
	)
	sum := sha256.Sum256([]byte(payload))
	return RuntimeOperationID(fmt.Sprintf("%s:%x", root, sum[:8]))
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
				plan.Actions[i].OperationID = deriveActionOperationID(operationID, plan.Actions[i])
			}
		}
	}

	desiredByID := make(map[string]DesiredResource, len(graph.Resources))
	for _, desired := range graph.Resources {
		desiredByID[desired.ID] = desired
	}

	// Preflight every realized identity before the first provider mutation.
	// A later wrong identity must not be discovered after earlier actions have
	// already changed external state.
	for _, action := range plan.Actions {
		if action.RealizedResourceID == "" {
			continue
		}
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
	}

	receipt := InfrastructureExecutionReceipt{
		Plan:         plan,
		Capabilities: caps,
		Executions:   make([]InfrastructureActionExecution, 0, len(plan.Actions)),
	}
	resolved := make(map[string]bool, len(plan.Actions))

	for _, action := range plan.Actions {
		execution := InfrastructureActionExecution{Action: action}

		if desired, ok := desiredByID[action.DesiredResourceID]; ok {
			for _, dependencyID := range desired.DependsOn {
				if !resolved[dependencyID] {
					receipt.Executions = append(receipt.Executions, execution)
					return receipt, fmt.Errorf(
						"%s %q blocked by unresolved dependency %q",
						action.Action,
						action.DesiredResourceID,
						dependencyID,
					)
				}
			}
		}

		switch action.Action {
		case ReconcileUnknown:
			resolved[action.DesiredResourceID] = false
			receipt.Executions = append(receipt.Executions, execution)
			continue

		case ReconcileNoop:
			resolved[action.DesiredResourceID] = true
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
			resolved[action.DesiredResourceID] = observation.Fresh
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
			value := realized[action.RealizedResourceID]
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

		execution.Attempted = true
		result, err := adapter.Apply(ctx, action, desired, current)
		if err != nil {
			receipt.Executions = append(receipt.Executions, execution)
			return receipt, fmt.Errorf(
				"apply %s %q: %w",
				action.Action,
				action.DesiredResourceID,
				err,
			)
		}
		execution.ApplyResult = &result
		switch result.Outcome {
		case InfrastructureApplyUnknown:
			if result.Realized != nil {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"UNKNOWN %s %q fabricated realized state",
					action.Action,
					action.DesiredResourceID,
				)
			}
			if result.ExternalOperation == nil {
				return InfrastructureExecutionReceipt{}, fmt.Errorf(
					"UNKNOWN %s %q lacks reconciliation operation identity",
					action.Action,
					action.DesiredResourceID,
				)
			}
		case InfrastructureApplyRealized:
			switch action.Action {
			case ReconcileCreate, ReconcileUpdate, ReconcileReplace:
				if result.Realized == nil {
					return InfrastructureExecutionReceipt{}, fmt.Errorf(
						"REALIZED %s %q returned no realized resource",
						action.Action,
						action.DesiredResourceID,
					)
				}
			case ReconcileDelete:
				if result.ExternalOperation == nil {
					return InfrastructureExecutionReceipt{}, fmt.Errorf(
						"REALIZED DELETE %q returned no provider operation/resource receipt",
						action.DesiredResourceID,
					)
				}
			}
		default:
			return InfrastructureExecutionReceipt{}, fmt.Errorf(
				"provider returned invalid apply outcome %q for %s %q",
				result.Outcome,
				action.Action,
				action.DesiredResourceID,
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

		execution.Applied = result.Outcome == InfrastructureApplyRealized
		resolved[action.DesiredResourceID] = execution.Applied
		receipt.Executions = append(receipt.Executions, execution)
	}

	return receipt, nil
}
