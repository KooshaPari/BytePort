package models

import "testing"

func TestInterruptedCreateBecomesUnknownInsteadOfDuplicateCreate(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{{
			ID:           "db",
			Kind:         DesiredResourceManaged,
			ConfigDigest: "v1",
			Target:       "prod",
			Lifecycle:    LifecycleManage,
		}},
	}
	plan, err := PlanReconciliationWithInterruptions(
		graph,
		nil,
		ReconciliationPolicy{},
		[]InterruptedResourceOperation{{
			OperationID:       "op-1",
			DesiredResourceID: "db",
			TargetID:          "prod",
			Action:            ReconcileCreate,
			State:             RuntimeOperationUnknown,
			ExternalOperation: &ExternalOperationRef{
				Provider:   "fixture",
				TargetID:   "prod",
				ExternalID: "provider-op-1",
				LookupKind: "operation",
			},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) != 1 || plan.Actions[0].Action != ReconcileUnknown {
		t.Fatalf("ambiguous create was retried: %v", plan.Actions)
	}
}

func TestInterruptedCreateForWrongTargetDoesNotSuppressRequestedTarget(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{{
			ID:           "db",
			Kind:         DesiredResourceManaged,
			ConfigDigest: "v1",
			Target:       "prod",
			Lifecycle:    LifecycleManage,
		}},
	}
	plan, err := PlanReconciliationWithInterruptions(
		graph,
		nil,
		ReconciliationPolicy{},
		[]InterruptedResourceOperation{{
			OperationID:       "op-other",
			DesiredResourceID: "db",
			TargetID:          "staging",
			Action:            ReconcileCreate,
			State:             RuntimeOperationUnknown,
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actions[0].Action != ReconcileCreate {
		t.Fatalf("wrong-target interruption suppressed create: %v", plan.Actions[0])
	}
}

func TestInterruptedReplaceRequiresExactRealizedResourceIdentity(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{{
			ID:              "host",
			Kind:            DesiredResourceHost,
			ConfigDigest:    "new",
			Target:          "lab",
			Lifecycle:       LifecycleManage,
			ReplaceOnChange: true,
		}},
	}
	observed := []ObservedResourceState{{
		DesiredResourceID:  "host",
		RealizedResourceID: "real-host-2",
		ConfigDigest:       "old",
		Fresh:              true,
		Lifecycle:          LifecycleManage,
	}}

	wrong, err := PlanReconciliationWithInterruptions(
		graph,
		observed,
		ReconciliationPolicy{},
		[]InterruptedResourceOperation{{
			OperationID:        "op-old",
			DesiredResourceID:  "host",
			RealizedResourceID: "real-host-1",
			TargetID:           "lab",
			Action:             ReconcileReplace,
			State:              RuntimeOperationUnknown,
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if wrong.Actions[0].Action != ReconcileReplace {
		t.Fatalf("wrong realized-resource interruption suppressed replace: %v", wrong.Actions[0])
	}

	exact, err := PlanReconciliationWithInterruptions(
		graph,
		observed,
		ReconciliationPolicy{},
		[]InterruptedResourceOperation{{
			OperationID:        "op-current",
			DesiredResourceID:  "host",
			RealizedResourceID: "real-host-2",
			TargetID:           "lab",
			Action:             ReconcileReplace,
			State:              RuntimeOperationReconciling,
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if exact.Actions[0].Action != ReconcileUnknown {
		t.Fatalf("exact ambiguous replace was repeated: %v", exact.Actions[0])
	}
}

func TestCompletedOrFailedOperationDoesNotRemainAmbiguityBarrier(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{{
			ID:           "db",
			Kind:         DesiredResourceManaged,
			ConfigDigest: "v1",
			Target:       "prod",
			Lifecycle:    LifecycleManage,
		}},
	}

	for _, state := range []RuntimeOperationState{RuntimeOperationRealized, RuntimeOperationFailed} {
		plan, err := PlanReconciliationWithInterruptions(
			graph,
			nil,
			ReconciliationPolicy{},
			[]InterruptedResourceOperation{{
				OperationID:       "op-terminal",
				DesiredResourceID: "db",
				TargetID:          "prod",
				Action:            ReconcileCreate,
				State:             state,
			}},
		)
		if err != nil {
			t.Fatal(err)
		}
		if plan.Actions[0].Action != ReconcileCreate {
			t.Fatalf("terminal state %s remained ambiguity barrier: %v", state, plan.Actions[0])
		}
	}
}

func TestApplyingBeforeOutcomeResolutionAlsoFailsClosed(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{{
			ID:           "network",
			Kind:         DesiredResourceNetwork,
			ConfigDigest: "v1",
			Target:       "lab",
			Lifecycle:    LifecycleManage,
		}},
	}
	plan, err := PlanReconciliationWithInterruptions(
		graph,
		nil,
		ReconciliationPolicy{},
		[]InterruptedResourceOperation{{
			OperationID:       "op-applying",
			DesiredResourceID: "network",
			TargetID:          "lab",
			Action:            ReconcileCreate,
			State:             RuntimeOperationApplying,
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actions[0].Action != ReconcileUnknown {
		t.Fatalf("APPLYING interruption allowed a second create: %v", plan.Actions[0])
	}
}
