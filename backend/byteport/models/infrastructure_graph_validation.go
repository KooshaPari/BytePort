package models

import "fmt"

// ValidateAndTopologicallyOrderGraph validates dependency identity and returns a
// deterministic topological order. Among resources that are ready at the same
// time, original declaration order is preserved.
//
// This is an additive reference planner. Live deployment routes do not consume
// it yet.
func ValidateAndTopologicallyOrderGraph(graph DesiredResourceGraph) ([]DesiredResource, error) {
	if len(graph.Resources) == 0 {
		return []DesiredResource{}, nil
	}

	byID := make(map[string]int, len(graph.Resources))
	for i, resource := range graph.Resources {
		if resource.ID == "" {
			return nil, fmt.Errorf("resource at index %d has empty ID", i)
		}
		if _, exists := byID[resource.ID]; exists {
			return nil, fmt.Errorf("duplicate desired resource ID %q", resource.ID)
		}
		byID[resource.ID] = i
	}

	indegree := make([]int, len(graph.Resources))
	dependents := make([][]int, len(graph.Resources))
	for i, resource := range graph.Resources {
		seenDeps := make(map[string]struct{}, len(resource.DependsOn))
		for _, dependencyID := range resource.DependsOn {
			if dependencyID == resource.ID {
				return nil, fmt.Errorf("resource %q depends on itself", resource.ID)
			}
			dependencyIndex, exists := byID[dependencyID]
			if !exists {
				return nil, fmt.Errorf(
					"resource %q depends on unknown resource %q",
					resource.ID,
					dependencyID,
				)
			}
			if _, duplicate := seenDeps[dependencyID]; duplicate {
				return nil, fmt.Errorf(
					"resource %q repeats dependency %q",
					resource.ID,
					dependencyID,
				)
			}
			seenDeps[dependencyID] = struct{}{}
			indegree[i]++
			dependents[dependencyIndex] = append(dependents[dependencyIndex], i)
		}
	}

	emitted := make([]bool, len(graph.Resources))
	ordered := make([]DesiredResource, 0, len(graph.Resources))

	for len(ordered) < len(graph.Resources) {
		next := -1
		for i := range graph.Resources {
			if !emitted[i] && indegree[i] == 0 {
				next = i
				break
			}
		}
		if next < 0 {
			return nil, fmt.Errorf("desired resource graph contains a dependency cycle")
		}

		emitted[next] = true
		ordered = append(ordered, graph.Resources[next])
		for _, dependent := range dependents[next] {
			indegree[dependent]--
		}
	}

	return ordered, nil
}

// PlanReconciliationChecked adds graph validity/topological ordering to the pure
// reconciliation reference model without changing the legacy unchecked helper.
func PlanReconciliationChecked(
	graph DesiredResourceGraph,
	observed []ObservedResourceState,
	policy ReconciliationPolicy,
) (ResourcePlan, error) {
	ordered, err := ValidateAndTopologicallyOrderGraph(graph)
	if err != nil {
		return ResourcePlan{}, err
	}
	graph.Resources = ordered
	return PlanReconciliation(graph, observed, policy), nil
}
