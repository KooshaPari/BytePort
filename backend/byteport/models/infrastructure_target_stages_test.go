package models

import "testing"

func TestTargetExecutionStagesPreserveCrossTargetDependencies(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{
			{
				ID: "service", Kind: DesiredResourceService, Target: "runtime",
				ConfigDigest: "svc", Lifecycle: LifecycleManage,
				DependsOn: []string{"host"},
			},
			{
				ID: "host", Kind: DesiredResourceHost, Target: "baremetal",
				ConfigDigest: "host", Lifecycle: LifecycleManage,
				DependsOn: []string{"network"},
			},
			{
				ID: "network", Kind: DesiredResourceNetwork, Target: "cloud",
				ConfigDigest: "net", Lifecycle: LifecycleManage,
			},
		},
	}

	stages, err := PlanTargetExecutionStages(graph)
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) != 3 {
		t.Fatalf("stages=%v", stages)
	}
	if stages[0].Batches[0].TargetID != "cloud" ||
		stages[0].Batches[0].Resources[0].ID != "network" {
		t.Fatalf("stage0=%v", stages[0])
	}
	if stages[1].Batches[0].TargetID != "baremetal" ||
		stages[1].Batches[0].Resources[0].ID != "host" {
		t.Fatalf("stage1=%v", stages[1])
	}
	if stages[2].Batches[0].TargetID != "runtime" ||
		stages[2].Batches[0].Resources[0].ID != "service" {
		t.Fatalf("stage2=%v", stages[2])
	}
}

func TestTargetExecutionStagesGroupIndependentSameTargetResources(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{
			{
				ID: "network-a", Kind: DesiredResourceNetwork, Target: "cloud",
				ConfigDigest: "a", Lifecycle: LifecycleManage,
			},
			{
				ID: "network-b", Kind: DesiredResourceNetwork, Target: "cloud",
				ConfigDigest: "b", Lifecycle: LifecycleManage,
			},
			{
				ID: "host", Kind: DesiredResourceHost, Target: "metal",
				ConfigDigest: "h", Lifecycle: LifecycleManage,
			},
		},
	}
	stages, err := PlanTargetExecutionStages(graph)
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) != 1 {
		t.Fatalf("stages=%v", stages)
	}
	if len(stages[0].Batches) != 2 {
		t.Fatalf("batches=%v", stages[0].Batches)
	}
	if stages[0].Batches[0].TargetID != "cloud" ||
		len(stages[0].Batches[0].Resources) != 2 {
		t.Fatalf("cloud batch=%v", stages[0].Batches[0])
	}
	if stages[0].Batches[1].TargetID != "metal" {
		t.Fatalf("metal batch=%v", stages[0].Batches[1])
	}
}

func TestTargetExecutionStagesRejectMissingTarget(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{{
			ID: "db", Kind: DesiredResourceManaged,
			ConfigDigest: "v1", Lifecycle: LifecycleManage,
		}},
	}
	if _, err := PlanTargetExecutionStages(graph); err == nil {
		t.Fatal("missing target was accepted")
	}
}

func TestTargetExecutionStagesRetainCycleRejection(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{
			{
				ID: "a", Kind: DesiredResourceManaged, Target: "one",
				ConfigDigest: "a", Lifecycle: LifecycleManage,
				DependsOn: []string{"b"},
			},
			{
				ID: "b", Kind: DesiredResourceManaged, Target: "two",
				ConfigDigest: "b", Lifecycle: LifecycleManage,
				DependsOn: []string{"a"},
			},
		},
	}
	if _, err := PlanTargetExecutionStages(graph); err == nil {
		t.Fatal("cross-target dependency cycle was accepted")
	}
}
