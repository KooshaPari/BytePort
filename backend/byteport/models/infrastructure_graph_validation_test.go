package models

import (
	"strings"
	"testing"
)

func graphResource(id string, deps ...string) DesiredResource {
	return DesiredResource{
		ID:           id,
		Kind:         DesiredResourceManaged,
		ConfigDigest: "digest-" + id,
		DependsOn:    deps,
		Target:       "target-1",
		Lifecycle:    LifecycleManage,
	}
}

func orderedIDs(resources []DesiredResource) []string {
	ids := make([]string, 0, len(resources))
	for _, resource := range resources {
		ids = append(ids, resource.ID)
	}
	return ids
}

func TestDesiredResourceGraphTopologicalOrderRespectsDependencies(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{
			graphResource("api", "db", "network"),
			graphResource("worker", "db"),
			graphResource("db", "network"),
			graphResource("network"),
		},
	}
	ordered, err := ValidateAndTopologicallyOrderGraph(graph)
	if err != nil {
		t.Fatal(err)
	}
	got := orderedIDs(ordered)
	want := []string{"network", "db", "api", "worker"}
	if len(got) != len(want) {
		t.Fatalf("order=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order=%v want=%v", got, want)
		}
	}
}

func TestDesiredResourceGraphPreservesDeclarationOrderAmongReadyResources(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{
			graphResource("z-first"),
			graphResource("a-second"),
			graphResource("consumer", "z-first", "a-second"),
		},
	}
	ordered, err := ValidateAndTopologicallyOrderGraph(graph)
	if err != nil {
		t.Fatal(err)
	}
	got := orderedIDs(ordered)
	if got[0] != "z-first" || got[1] != "a-second" {
		t.Fatalf("ready resources reordered lexically: %v", got)
	}
}

func TestDesiredResourceGraphRejectsDuplicateIDs(t *testing.T) {
	graph := DesiredResourceGraph{
		ID:        "g",
		Resources: []DesiredResource{graphResource("db"), graphResource("db")},
	}
	if _, err := ValidateAndTopologicallyOrderGraph(graph); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate-ID error, got %v", err)
	}
}

func TestDesiredResourceGraphRejectsUnknownDependency(t *testing.T) {
	graph := DesiredResourceGraph{
		ID:        "g",
		Resources: []DesiredResource{graphResource("api", "missing")},
	}
	if _, err := ValidateAndTopologicallyOrderGraph(graph); err == nil || !strings.Contains(err.Error(), "unknown resource") {
		t.Fatalf("expected unknown-dependency error, got %v", err)
	}
}

func TestDesiredResourceGraphRejectsSelfDependency(t *testing.T) {
	graph := DesiredResourceGraph{
		ID:        "g",
		Resources: []DesiredResource{graphResource("db", "db")},
	}
	if _, err := ValidateAndTopologicallyOrderGraph(graph); err == nil || !strings.Contains(err.Error(), "itself") {
		t.Fatalf("expected self-dependency error, got %v", err)
	}
}

func TestDesiredResourceGraphRejectsCycles(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{
			graphResource("a", "b"),
			graphResource("b", "a"),
		},
	}
	if _, err := ValidateAndTopologicallyOrderGraph(graph); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

func TestCheckedReconciliationPlansDesiredActionsInTopologicalOrder(t *testing.T) {
	graph := DesiredResourceGraph{
		ID: "g",
		Resources: []DesiredResource{
			graphResource("service", "network"),
			graphResource("network"),
		},
	}
	plan, err := PlanReconciliationChecked(graph, nil, ReconciliationPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) != 2 {
		t.Fatalf("actions=%v", plan.Actions)
	}
	if plan.Actions[0].DesiredResourceID != "network" || plan.Actions[1].DesiredResourceID != "service" {
		t.Fatalf("actions not topological: %v", plan.Actions)
	}
}
