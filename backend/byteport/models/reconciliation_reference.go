package models

// Pure generalized infrastructure reconciliation reference model.
// This is not wired into live target adapters.

type ObservedResourceState struct {
	DesiredResourceID string
	RealizedResourceID string
	ConfigDigest string
	Fresh bool
}

type ReconciliationPolicy struct {
	AllowDelete bool
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
		if policy.AllowDelete {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:o.DesiredResourceID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileDelete,Reason:"resource absent from desired graph and destruction authorized"})
		} else {
			actions=append(actions,PlannedResourceAction{DesiredResourceID:o.DesiredResourceID,RealizedResourceID:o.RealizedResourceID,Action:ReconcileRead,Reason:"resource absent from desired graph but destruction not authorized"})
		}
	}

	return ResourcePlan{ID:"reference-plan",DesiredGraphID:graph.ID,Actions:actions}
}
