package models

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestGeneralizedNanoVMSHTTPLifecycleCreateObserveRestartNoopDelete(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	deployCalls := 0
	stopCalls := 0
	sandboxExists := false
	configDigest := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/deploy":
			deployCalls++
			var body nanoVMSHTTPDeployBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode deploy: %v", err)
			}
			configDigest = body.Labels["byteport-config-digest"]
			sandboxExists = true
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"id":"sandbox-e2e-http","name":"service","status":"running"}`)

		case r.Method == http.MethodGet && r.URL.Path == "/v1/sandboxes/sandbox-e2e-http":
			if !sandboxExists {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(nanoVMSHTTPSandbox{
				ID: "sandbox-e2e-http", Name: "service", Status: "running",
				Labels: map[string]string{"byteport-config-digest": configDigest},
			})

		case r.Method == http.MethodPost && r.URL.Path == "/v1/stop":
			if r.URL.Query().Get("id") != "sandbox-e2e-http" {
				t.Fatalf("stop targeted %q", r.URL.Query().Get("id"))
			}
			stopCalls++
			sandboxExists = false
			_, _ = io.WriteString(w, `{"status":"stopped"}`)

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	artifactID := BuildArtifactID("artifact-http-e2e")
	manifestID := ManifestRevisionID("manifest-http-e2e")
	graph := DesiredResourceGraph{
		ID:       "graph-http-e2e",
		Manifest: manifestID,
		Resources: []DesiredResource{{
			ID: "service", Kind: DesiredResourceService, ConfigDigest: "cfg-http-e2e",
			Target: "nanovms-http", Lifecycle: LifecycleManage, Artifact: &artifactID,
		}},
	}
	artifacts := lifecycleArtifactResolver{
		artifacts: map[BuildArtifactID]BuildArtifact{
			artifactID: {
				ID: artifactID, ImmutableRef: "sha256:http-e2e", MediaKind: "oci-image",
				ManifestRevision: manifestID,
			},
		},
	}
	if err := ValidateDesiredGraphArtifactLineage(ctx, graph, artifacts); err != nil {
		t.Fatal(err)
	}

	adapter := NanoVMSInfrastructureAdapter{
		Transport: NanoVMSHTTPTransport{BaseURL: server.URL, Client: server.Client()},
		Artifacts: artifacts,
		TargetID:  "nanovms-http",
		Provider:  "nanovms",
	}

	create, err := ReconcileOnceWithTargetAdapterOperation(
		ctx, "root-create-http", graph, nil, map[string]RealizedResource{},
		ReconciliationPolicy{}, nil, "nanovms-http", adapter,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(create.Executions) != 1 || !create.Executions[0].Applied ||
		create.Executions[0].ApplyResult == nil ||
		create.Executions[0].ApplyResult.Realized == nil {
		t.Fatalf("create receipt=%+v", create)
	}
	realized := *create.Executions[0].ApplyResult.Realized

	observation, err := adapter.Observe(ctx, realized)
	if err != nil {
		t.Fatal(err)
	}
	if !observation.Fresh || observation.ConfigDigest != graph.Resources[0].ConfigDigest {
		t.Fatalf("observation=%+v", observation)
	}

	observed := []ObservedResourceState{{
		DesiredResourceID: graph.Resources[0].ID, RealizedResourceID: realized.ID,
		TargetID: realized.TargetID, Provider: realized.Provider,
		ConfigDigest: observation.ConfigDigest, Fresh: true, Lifecycle: LifecycleManage,
	}}
	restart, err := ReconcileOnceWithTargetAdapterOperation(
		ctx, "root-restart-http", graph, observed,
		map[string]RealizedResource{realized.ID: realized},
		ReconciliationPolicy{}, nil, "nanovms-http", adapter,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(restart.Plan.Actions) != 1 || restart.Plan.Actions[0].Action != ReconcileNoop {
		t.Fatalf("restart plan=%+v", restart.Plan.Actions)
	}
	mu.Lock()
	if deployCalls != 1 {
		t.Fatalf("restart duplicated deploy: %d", deployCalls)
	}
	mu.Unlock()

	deleteObserved := []ObservedResourceState{{
		DesiredResourceID: graph.Resources[0].ID, RealizedResourceID: realized.ID,
		TargetID: realized.TargetID, Provider: realized.Provider,
		ConfigDigest: observation.ConfigDigest, Fresh: true,
		Lifecycle: LifecycleDestroyOnExplicitIntent,
	}}
	policy := ReconciliationPolicy{DestructionIntents: []DestructionIntent{{
		ID: "destroy-http-e2e", DesiredResourceID: graph.Resources[0].ID,
		RealizedResourceID: realized.ID, AuthorizedBy: "http-e2e-fixture",
		Reason: "dispose exact fixture",
	}}}
	deleted, err := ReconcileOnceWithTargetAdapterOperation(
		ctx, "root-delete-http",
		DesiredResourceGraph{ID: "graph-http-e2e-removed", Manifest: manifestID},
		deleteObserved, map[string]RealizedResource{realized.ID: realized},
		policy, nil, "nanovms-http", adapter,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted.Executions) != 1 || !deleted.Executions[0].Applied {
		t.Fatalf("delete receipt=%+v", deleted)
	}
	mu.Lock()
	defer mu.Unlock()
	if deployCalls != 1 || stopCalls != 1 || sandboxExists {
		t.Fatalf("deploy=%d stop=%d exists=%v", deployCalls, stopCalls, sandboxExists)
	}
}
