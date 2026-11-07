package models

import "fmt"

// PlanFixtureReconciliation is a bounded B08 reference model.
// It intentionally cannot plan deletion; removal/destruction needs a separate
// exact desired-resource-removal + DestructionIntent model.
type RealizedResourceView struct {
	Resource RealizedResource
	ConfigDigest string
	Observed bool
}

func PlanFixtureReconciliation(graph DesiredResourceGraph, realized map[string]RealizedResourceView, caps TargetCapabilities) (ResourcePlan,error) {
	plan:=ResourcePlan{ID:"fixture-plan",DesiredGraphID:graph.ID,TargetID:caps.TargetID}
	for _,d:=range graph.Resources {
		rv,ok:=realized[d.ID]
		if !ok {
			if d.Lifecycle==LifecycleObserveOnly {
				plan.Actions=append(plan.Actions,PlannedResourceAction{DesiredResourceID:d.ID,Action:ReconcileUnknown,Reason:"observe-only resource not found"})
				continue
			}
			if !caps.SupportsCreate { return plan,fmt.Errorf("target cannot create %s",d.ID) }
			plan.Actions=append(plan.Actions,PlannedResourceAction{DesiredResourceID:d.ID,Action:ReconcileCreate,Reason:"desired resource missing"})
			continue
		}
		if !rv.Observed {
			plan.Actions=append(plan.Actions,PlannedResourceAction{DesiredResourceID:d.ID,RealizedResourceID:rv.Resource.ID,Action:ReconcileUnknown,Reason:"observation unavailable"})
			continue
		}
		if rv.ConfigDigest==d.ConfigDigest {
			plan.Actions=append(plan.Actions,PlannedResourceAction{DesiredResourceID:d.ID,RealizedResourceID:rv.Resource.ID,Action:ReconcileNoop,Reason:"observed configuration matches desired"})
			continue
		}
		if d.ReplaceOnChange {
			if !caps.SupportsReplace { return plan,fmt.Errorf("target cannot replace %s",d.ID) }
			plan.Actions=append(plan.Actions,PlannedResourceAction{DesiredResourceID:d.ID,RealizedResourceID:rv.Resource.ID,Action:ReconcileReplace,Reason:"replacement-required change"})
		} else {
			if !caps.SupportsUpdate { return plan,fmt.Errorf("target cannot update %s",d.ID) }
			plan.Actions=append(plan.Actions,PlannedResourceAction{DesiredResourceID:d.ID,RealizedResourceID:rv.Resource.ID,Action:ReconcileUpdate,Reason:"configuration drift"})
		}
	}
	return plan,nil
}
