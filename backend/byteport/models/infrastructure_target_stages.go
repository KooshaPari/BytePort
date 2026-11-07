package models

import "fmt"

// TargetResourceBatch is one target-specific subset that may execute within a
// dependency stage. Resources inside one stage have no dependency on another
// resource in that same stage.
type TargetResourceBatch struct {
	TargetID  string
	Resources []DesiredResource
}

// TargetExecutionStage preserves global dependency order while allowing
// independent resources for different targets to remain explicit.
type TargetExecutionStage struct {
	Index   int
	Batches []TargetResourceBatch
}

// PlanTargetExecutionStages converts one heterogeneous DesiredResourceGraph
// into deterministic dependency stages without flattening target identity.
//
// A cross-target dependency always advances the dependent resource to a later
// stage. This is a planning reference only; it does not perform provider calls.
func PlanTargetExecutionStages(graph DesiredResourceGraph) ([]TargetExecutionStage, error) {
	ordered, err := ValidateAndTopologicallyOrderGraph(graph)
	if err != nil {
		return nil, err
	}

	byID := make(map[string]DesiredResource, len(graph.Resources))
	for _, resource := range graph.Resources {
		if resource.Target == "" {
			return nil, fmt.Errorf("desired resource %q has no target", resource.ID)
		}
		byID[resource.ID] = resource
	}

	level := make(map[string]int, len(ordered))
	maxLevel := -1
	for _, resource := range ordered {
		resourceLevel := 0
		for _, dependencyID := range resource.DependsOn {
			dependency, ok := byID[dependencyID]
			if !ok {
				return nil, fmt.Errorf(
					"desired resource %q depends on missing resource %q",
					resource.ID,
					dependencyID,
				)
			}
			dependencyLevel, ok := level[dependency.ID]
			if !ok {
				return nil, fmt.Errorf(
					"dependency %q was not topologically ordered before %q",
					dependency.ID,
					resource.ID,
				)
			}
			if dependencyLevel+1 > resourceLevel {
				resourceLevel = dependencyLevel + 1
			}
		}
		level[resource.ID] = resourceLevel
		if resourceLevel > maxLevel {
			maxLevel = resourceLevel
		}
	}

	if maxLevel < 0 {
		return []TargetExecutionStage{}, nil
	}

	stages := make([]TargetExecutionStage, maxLevel+1)
	for i := range stages {
		stages[i].Index = i
	}

	for _, resource := range ordered {
		stage := &stages[level[resource.ID]]
		batchIndex := -1
		for i := range stage.Batches {
			if stage.Batches[i].TargetID == resource.Target {
				batchIndex = i
				break
			}
		}
		if batchIndex < 0 {
			stage.Batches = append(stage.Batches, TargetResourceBatch{
				TargetID: resource.Target,
			})
			batchIndex = len(stage.Batches) - 1
		}
		stage.Batches[batchIndex].Resources = append(
			stage.Batches[batchIndex].Resources,
			resource,
		)
	}

	return stages, nil
}
