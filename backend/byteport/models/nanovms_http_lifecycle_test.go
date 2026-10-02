// FR: BP-WP-B08. HTTP fixture composition is not upstream provider validation.
package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

// Only the test provider has this removal API. Do not add it to
// NanoVMSHTTPTransport without an independently demonstrated provider contract.
type fixtureDeletionHTTPTransport struct{ NanoVMSHTTPTransport }

func (f fixtureDeletionHTTPTransport) Delete(ctx context.Context, id, operation string) error {
	if id == "" || operation == "" {
		return fmt.Errorf("fixture deletion identity required")
	}
	path := "/fixture/delete?id=" + url.QueryEscape(id) + "&operation=" + url.QueryEscape(operation)
	response, err := f.request(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	_, err = readNanoVMSBody(response)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("fixture delete status %d", response.StatusCode)
	}
	return nil
}

func TestGeneralizedNanoVMSHTTPLifecycleCreateObserveRestartNoopDelete(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	deploys, stops, deletes := 0, 0, 0
	exists, status, config, deletionOperation := false, "", "", ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/deploy":
			var body nanoVMSHTTPDeployBody
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", 400)
				return
			}
			deploys++
			exists, status, config = true, "running", body.Labels["byteport-config-digest"]
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(nanoVMSHTTPSandbox{ID: "sandbox-e2e-http", Name: "service", Status: status})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/sandboxes/sandbox-e2e-http":
			if !exists {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(nanoVMSHTTPSandbox{ID: "sandbox-e2e-http", Name: "service", Status: status, Labels: map[string]string{"byteport-config-digest": config}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/stop":
			if r.URL.Query().Get("id") != "sandbox-e2e-http" {
				http.Error(w, "wrong identity", 400)
				return
			}
			stops++
			status = "stopped" // The resource remains present.
			_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
		case r.Method == http.MethodDelete && r.URL.Path == "/fixture/delete":
			if r.URL.Query().Get("id") != "sandbox-e2e-http" || r.URL.Query().Get("operation") == "" {
				http.Error(w, "wrong identity", 400)
				return
			}
			deletes++
			deletionOperation = r.URL.Query().Get("operation")
			exists = false
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	artifactID := BuildArtifactID("artifact-http-e2e")
	manifestID := ManifestRevisionID("manifest-http-e2e")
	graph := DesiredResourceGraph{ID: "graph-http-e2e", Manifest: manifestID, Resources: []DesiredResource{{ID: "service", Kind: DesiredResourceService, ConfigDigest: "cfg-http-e2e", Target: "nanovms-http", Lifecycle: LifecycleManage, Artifact: &artifactID}}}
	artifacts := lifecycleArtifactResolver{artifacts: map[BuildArtifactID]BuildArtifact{artifactID: {ID: artifactID, ImmutableRef: "sha256:http-e2e", MediaKind: "oci-image", ManifestRevision: manifestID}}}
	if err := ValidateDesiredGraphArtifactLineage(ctx, graph, artifacts); err != nil {
		t.Fatal(err)
	}
	transport := fixtureDeletionHTTPTransport{NanoVMSHTTPTransport{BaseURL: server.URL, Client: server.Client()}}
	adapter := NanoVMSInfrastructureAdapter{Transport: transport, Artifacts: artifacts, TargetID: "nanovms-http", Provider: "nanovms"}
	created, err := ReconcileOnceWithTargetAdapterOperation(ctx, "root-create-http", graph, nil, map[string]RealizedResource{}, ReconciliationPolicy{}, nil, "nanovms-http", adapter)
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Executions) != 1 || !created.Executions[0].Applied || created.Executions[0].ApplyResult == nil || created.Executions[0].ApplyResult.Realized == nil {
		t.Fatalf("create=%+v", created)
	}
	realized := *created.Executions[0].ApplyResult.Realized
	observation, err := adapter.Observe(ctx, realized)
	if err != nil {
		t.Fatal(err)
	}
	if !observation.Fresh || observation.ConfigDigest != graph.Resources[0].ConfigDigest {
		t.Fatalf("observe=%+v", observation)
	}
	observed := []ObservedResourceState{{DesiredResourceID: graph.Resources[0].ID, RealizedResourceID: realized.ID, TargetID: realized.TargetID, Provider: realized.Provider, ConfigDigest: observation.ConfigDigest, Fresh: true, Lifecycle: LifecycleManage}}
	resources := map[string]RealizedResource{realized.ID: realized}
	// This reconstructs an adapter; it is not a persisted journal/process restart.
	resumed := NanoVMSInfrastructureAdapter{Transport: transport, Artifacts: artifacts, TargetID: "nanovms-http", Provider: "nanovms"}
	replay, err := ReconcileOnceWithTargetAdapterOperation(ctx, "root-restart-http", graph, observed, resources, ReconciliationPolicy{}, nil, "nanovms-http", resumed)
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.Plan.Actions) != 1 || replay.Plan.Actions[0].Action != ReconcileNoop {
		t.Fatalf("replay=%+v", replay)
	}
	observed[0].Lifecycle = LifecycleDestroyOnExplicitIntent
	policy := ReconciliationPolicy{DestructionIntents: []DestructionIntent{{ID: "destroy-http-e2e", DesiredResourceID: graph.Resources[0].ID, RealizedResourceID: realized.ID, AuthorizedBy: "http-e2e-fixture", Reason: "dispose exact fixture"}}}
	deleted, err := ReconcileOnceWithTargetAdapterOperation(ctx, "root-delete-http", DesiredResourceGraph{ID: "graph-http-e2e-removed", Manifest: manifestID}, observed, resources, policy, nil, "nanovms-http", resumed)
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted.Executions) != 1 || !deleted.Executions[0].Applied || deleted.Executions[0].ApplyResult == nil || deleted.Executions[0].ApplyResult.Observation == nil || deleted.Executions[0].ApplyResult.Observation.State != "absent" {
		t.Fatalf("delete=%+v", deleted)
	}
	mu.Lock()
	defer mu.Unlock()
	if deploys != 1 || stops != 0 || deletes != 1 || exists || deletionOperation != string(deleted.Executions[0].Action.OperationID) {
		t.Fatalf("deploys=%d stops=%d deletes=%d exists=%v operation=%q", deploys, stops, deletes, exists, deletionOperation)
	}
}

func TestGeneralizedNanoVMSHTTPStopCannotSatisfyDelete(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		if r.Method == http.MethodPost && r.URL.Path == "/v1/stop" {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v1/sandboxes/retained" {
			_ = json.NewEncoder(w).Encode(nanoVMSHTTPSandbox{ID: "retained", Name: "service", Status: "stopped"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	transport := NanoVMSHTTPTransport{BaseURL: server.URL, Client: server.Client()}
	adapter := NanoVMSInfrastructureAdapter{Transport: transport, TargetID: "target", Provider: "nanovms"}
	caps, err := adapter.Capabilities(ctx, "target")
	if err != nil {
		t.Fatal(err)
	}
	if caps.SupportsDelete {
		t.Fatal("ordinary HTTP Stop transport advertised Delete")
	}
	realized := RealizedResource{ID: "real", DesiredResourceID: "service", TargetID: "target", Provider: "nanovms", ExternalID: "retained"}
	_, err = adapter.Apply(ctx, PlannedResourceAction{OperationID: "delete-op", DesiredResourceID: "service", RealizedResourceID: "real", Action: ReconcileDelete}, nil, &realized)
	if err == nil {
		t.Fatal("unsupported DELETE accepted")
	}
	mu.Lock()
	mutationCalls := calls
	mu.Unlock()
	if mutationCalls != 0 {
		t.Fatalf("unsupported DELETE made %d HTTP calls", mutationCalls)
	}
	if err := transport.Stop(ctx, "retained"); err != nil {
		t.Fatal(err)
	}
	result, found, err := transport.Observe(ctx, "retained")
	if err != nil || !found || result.Status != "stopped" {
		t.Fatalf("Stop incorrectly established absence: result=%+v found=%v err=%v", result, found, err)
	}
}
