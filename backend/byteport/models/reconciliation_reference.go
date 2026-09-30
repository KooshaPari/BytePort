package models

// Pure generalized infrastructure reconciliation reference model.
// This is not wired into live target adapters.

type ObservedResourceState struct {
	DesiredResourceID string
	RealizedResourceID string
	ConfigDigest string
	Fresh bool
	Lifecycle LifecyclePolicy
}

type DestructionIntent struct {
	DesiredResourceID string
	RealizedResourceID string
	Authorized bool
	Reason string
}

type ReconciliationPolicy struct {
	DestructionIntents []DestructionIntent
}

func findDestructionIntent(policy ReconciliationPolicy, desiredID, realizedID string) (DestructionIntent, bool) {
	for _, intent := range policy.DestructionIntents {
		if intent.DesiredResourceID == desiredID && intent.RealizedResourceID == realizedID && intent.Authorized {
			return intent, true
		}
	}
	return DestructionIntent{}, false
}

func PlanReconciliation(graph DesiredResourceGraph, observed []ObservedResourceState, policy ReconciliationPolicy) ResourcePlan {
	byDesired:=map[string]ObservedResourceState{}
	for _,o:=range observed { byDesired[o.DesiredResourceID]=o }

	actions:=make([]PlannedResourceAction,0,len(graph.Resources)+len(observed))
	desiredIDs:=map[string]struct{}{}
	for _,d:=range graph.Resources {
		desiredIDs[d.ID]=struct{}{}
		o,ok:=byDesired[d.ID]
		if !ok {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:d.ID,Action:ReconcileCreate,Reason:"desired resource absent"})
			continue
		}
		if !o.Fresh {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:d.ID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileUnknown,Reason:"observation stale or incomplete"})
			continue
		}
		if o.ConfigDigest==d.ConfigDigest {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:d.ID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileNoop,Reason:"observed configuration matches desired"})
		} else if d.ReplaceOnChange {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:d.ID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileReplace,Reason:"configuration changed and resource requires replacement"})
		} else {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:d.ID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileUpdate,Reason:"observed configuration differs"})
		}
	}

	for _,o:=range observed {
		if _,ok:=desiredIDs[o.DesiredResourceID];ok { continue }
		if !o.Fresh {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:o.DesiredResourceID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileUnknown,Reason:"orphan observation stale"})
			continue
		}
		intent, authorized := findDestructionIntent(policy, o.DesiredResourceID, o.RealizedResourceID)
		if o.Lifecycle == LifecycleDestroyOnExplicitIntent && authorized {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:o.DesiredResourceID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileDelete,Reason:intent.Reason})
		} else {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:o.DesiredResourceID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileRead,Reason:"resource absent from desired graph but lifecycle/intent does not authorize destruction"})
		}
	}

	return ResourcePlan{ID:"reference-plan",DesiredGraphID:graph.ID,Actions:actions}
}


func ConstrainPlanToTarget(plan ResourcePlan, caps TargetCapabilities) ResourcePlan {
	out := plan
	out.TargetID = caps.TargetID
	out.Actions = append([]PlannedResourceAction(nil), plan.Actions...)
	for i, action := range out.Actions {
		supported := true
		switch action.Action {
		case ReconcileRead:
			supported = caps.SupportsObserve
		case ReconcileCreate:
			supported = caps.SupportsCreate
		case ReconcileUpdate:
			supported = caps.SupportsUpdate
		case ReconcileReplace:
			supported = caps.SupportsReplace
		case ReconcileDelete:
			supported = caps.SupportsDelete
		case ReconcileNoop, ReconcileUnknown:
			supported = true
		}
		if !supported {
			out.Actions[i].Action = ReconcileUnknown
			out.Actions[i].Reason = "target capability does not support planned action"
		}
	}
	return out
}
