package models

// InterruptedResourceOperation records a provider mutation whose outcome may be
// ambiguous after transport failure, process interruption, or restart.
//
// It is deliberately separate from observed/realized state: an operation may
// exist even when BytePort cannot yet prove whether the provider created or
// replaced the resource.
type InterruptedResourceOperation struct {
	OperationID        RuntimeOperationID
	DesiredResourceID  string
	RealizedResourceID string
	TargetID           string
	Action             ReconciliationAction
	State              RuntimeOperationState
	ExternalOperation  *ExternalOperationRef
}

func ambiguousMutationState(state RuntimeOperationState) bool {
	switch state {
	case RuntimeOperationApplying, RuntimeOperationUnknown, RuntimeOperationReconciling:
		return true
	default:
		return false
	}
}

func interruptionMatchesAction(
	interruption InterruptedResourceOperation,
	action PlannedResourceAction,
	desired DesiredResource,
) bool {
	if interruption.OperationID == "" ||
		interruption.DesiredResourceID != action.DesiredResourceID ||
		interruption.TargetID != desired.Target ||
		interruption.Action != action.Action ||
		!ambiguousMutationState(interruption.State) {
		return false
	}

	if action.Action == ReconcileReplace {
		return action.RealizedResourceID != "" &&
			interruption.RealizedResourceID == action.RealizedResourceID
	}

	return action.Action == ReconcileCreate
}

// PlanReconciliationWithInterruptions prevents an ambiguous prior CREATE or
// REPLACE from being repeated blindly. The result becomes UNKNOWN until an
// adapter observation/reconciliation step resolves the provider outcome.
func PlanReconciliationWithInterruptions(
	graph DesiredResourceGraph,
	observed []ObservedResourceState,
	policy ReconciliationPolicy,
	interruptions []InterruptedResourceOperation,
) (ResourcePlan, error) {
	plan, err := PlanReconciliationChecked(graph, observed, policy)
	if err != nil {
		return ResourcePlan{}, err
	}

	desiredByID := make(map[string]DesiredResource, len(graph.Resources))
	for _, desired := range graph.Resources {
		desiredByID[desired.ID] = desired
	}

	for i, action := range plan.Actions {
		if action.Action != ReconcileCreate && action.Action != ReconcileReplace {
			continue
		}
		desired, exists := desiredByID[action.DesiredResourceID]
		if !exists {
			continue
		}
		for _, interruption := range interruptions {
			if interruptionMatchesAction(interruption, action, desired) {
				plan.Actions[i].Action = ReconcileUnknown
				plan.Actions[i].Reason = "prior provider mutation outcome is ambiguous; observe/reconcile before retry"
				break
			}
		}
	}

	return plan, nil
}
