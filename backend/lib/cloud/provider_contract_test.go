package cloud

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type contractRoundTripper func(*http.Request) (*http.Response, error)

func (f contractRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func contractResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func providerClient(fn contractRoundTripper) *http.Client {
	return &http.Client{Transport: fn}
}

func TestNetlifyProviderContract(t *testing.T) {
	ctx := context.Background()
	provider, err := NewNetlifyProvider(Credentials{Data: map[string]string{"token": "netlify-token"}})
	if err != nil {
		t.Fatalf("construct provider: %v", err)
	}
	n := provider.(*NetlifyProvider)
	n.httpClient = providerClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/api/v1/user":
			return contractResponse(req, http.StatusOK, `{"id":"user-1"}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v1/sites":
			return contractResponse(req, http.StatusOK, `{"id":"site-1","name":"demo","url":"https://demo.netlify.app","state":"ready","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T01:00:00Z"}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/api/v1/sites/site-1":
			return contractResponse(req, http.StatusOK, `{"id":"site-1","name":"demo","url":"https://demo.netlify.app","state":"ready"}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/api/v1/sites":
			return contractResponse(req, http.StatusOK, `[{"id":"site-1","name":"demo","url":"https://demo.netlify.app","state":"ready"}]`), nil
		case req.Method == http.MethodDelete && req.URL.Path == "/api/v1/sites/site-1":
			return contractResponse(req, http.StatusNoContent, ``), nil
		case req.Method == http.MethodPost && req.URL.Path == "/api/v1/sites/site-1/deploys":
			return contractResponse(req, http.StatusOK, `{"id":"deploy-1","site_id":"site-1","state":"ready","url":"https://deploy.netlify.app","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T01:00:00Z"}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/api/v1/deploys/deploy-1":
			return contractResponse(req, http.StatusOK, `{"id":"deploy-1","site_id":"site-1","state":"ready","url":"https://deploy.netlify.app"}`), nil
		default:
			return contractResponse(req, http.StatusNotFound, `{"error":"unexpected request"}`), nil
		}
	})

	if got := n.GetMetadata(); got.Name != "netlify" || !n.SupportsResource(ResourceTypeComputeEdge) || n.SupportsResource(ResourceTypeComputeVM) {
		t.Fatalf("unexpected metadata or resource support: %#v", got)
	}
	if err := n.Initialize(ctx, Credentials{Data: map[string]string{"token": "netlify-token"}}); err != nil {
		t.Fatalf("validate credentials: %v", err)
	}
	resource, err := n.CreateResource(ctx, ResourceConfig{
		Name: "demo", Type: ResourceTypeComputeEdge,
		Spec: map[string]any{"custom_domain": "demo.example", "repo_url": "https://github.com/example/demo", "branch": "release", "build_command": "make build", "publish_dir": "dist"},
	})
	if err != nil || resource.ID != "site-1" || len(resource.Endpoints) != 1 {
		t.Fatalf("create resource: resource=%#v err=%v", resource, err)
	}
	if got, err := n.GetResource(ctx, "site-1"); err != nil || got.Name != "demo" {
		t.Fatalf("get resource: resource=%#v err=%v", got, err)
	}
	if got, err := n.ListResources(ctx, ResourceFilter{}); err != nil || len(got) != 1 {
		t.Fatalf("list resources: resources=%#v err=%v", got, err)
	}
	if err := n.DeleteResource(ctx, "site-1"); err != nil {
		t.Fatalf("delete resource: %v", err)
	}
	dep, err := n.Deploy(ctx, DeploymentConfig{ResourceID: "site-1", Env: map[string]string{"MODE": "prod"}, Config: map[string]any{"clear_cache": "true"}})
	if err != nil || dep.ID != "deploy-1" || dep.State != DeploymentStateActive {
		t.Fatalf("deploy: deployment=%#v err=%v", dep, err)
	}
	status, err := n.GetDeploymentStatus(ctx, "deploy-1")
	if err != nil || status.Health != HealthStatusHealthy {
		t.Fatalf("deployment status: status=%#v err=%v", status, err)
	}
	if estimate, err := n.EstimateCost(ctx, ResourceConfig{}); err != nil || estimate.Currency != "USD" {
		t.Fatalf("estimate cost: estimate=%#v err=%v", estimate, err)
	}
	for name, call := range map[string]func() error{
		"update":      func() error { _, err := n.UpdateResource(ctx, "site-1", ResourceConfig{}); return err },
		"rollback":    func() error { return n.RollbackDeployment(ctx, "deploy-1") },
		"logs":        func() error { _, err := n.GetLogs(ctx, resource, LogOptions{}); return err },
		"metrics":     func() error { _, err := n.GetMetrics(ctx, resource, MetricOptions{}); return err },
		"actual cost": func() error { _, err := n.GetActualCost(ctx, resource, TimeRange{}); return err },
	} {
		if call() == nil {
			t.Errorf("%s unexpectedly succeeded", name)
		}
	}
	if _, err := n.CreateResource(ctx, ResourceConfig{Name: "vm", Type: ResourceTypeComputeVM}); err == nil {
		t.Error("unsupported resource unexpectedly succeeded")
	}
	if _, err := NewNetlifyProvider(Credentials{}); err == nil {
		t.Error("missing token unexpectedly succeeded")
	}
	n.token = ""
	if err := n.ValidateCredentials(ctx); err == nil {
		t.Error("empty token unexpectedly validated")
	}
}

func TestRailwayProviderContract(t *testing.T) {
	ctx := context.Background()
	provider, err := NewRailwayProvider(Credentials{Data: map[string]string{"token": "railway-token"}})
	if err != nil {
		t.Fatalf("construct provider: %v", err)
	}
	r := provider.(*RailwayProvider)
	r.httpClient = providerClient(func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		query := string(body)
		switch {
		case strings.Contains(query, "me { id }"):
			return contractResponse(req, http.StatusOK, `{"data":{"me":{"id":"user-1"}}}`), nil
		case strings.Contains(query, "projectCreate"):
			return contractResponse(req, http.StatusOK, `{"data":{"projectCreate":{"id":"project-1","name":"demo","description":"test","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T01:00:00Z"}}}`), nil
		case strings.Contains(query, "query Project"):
			return contractResponse(req, http.StatusOK, `{"data":{"project":{"id":"project-1","name":"demo","description":"test"}}}`), nil
		case strings.Contains(query, "projects"):
			return contractResponse(req, http.StatusOK, `{"data":{"me":{"projects":{"edges":[{"node":{"id":"project-1","name":"demo","description":"test"}}]}}}}`), nil
		case strings.Contains(query, "projectDelete"):
			return contractResponse(req, http.StatusOK, `{"data":{"projectDelete":true}}`), nil
		case strings.Contains(query, "serviceInstanceDeploy"):
			return contractResponse(req, http.StatusOK, `{"data":{"serviceInstanceDeploy":true}}`), nil
		default:
			return contractResponse(req, http.StatusNotFound, `{"errors":[{"message":"unexpected request"}]}`), nil
		}
	})

	if got := r.GetMetadata(); got.Name != "railway" || !r.SupportsResource(ResourceTypeComputeContainer) || r.SupportsResource(ResourceTypeComputeEdge) {
		t.Fatalf("unexpected metadata or resource support: %#v", got)
	}
	if err := r.Initialize(ctx, Credentials{Data: map[string]string{"token": "railway-token"}}); err != nil {
		t.Fatalf("validate credentials: %v", err)
	}
	resource, err := r.CreateResource(ctx, ResourceConfig{Name: "demo", Type: ResourceTypeComputeContainer, Spec: map[string]any{"description": "test", "team_id": "team-1"}})
	if err != nil || resource.ID != "project-1" {
		t.Fatalf("create resource: resource=%#v err=%v", resource, err)
	}
	if got, err := r.GetResource(ctx, "project-1"); err != nil || got.Name != "demo" {
		t.Fatalf("get resource: resource=%#v err=%v", got, err)
	}
	if got, err := r.ListResources(ctx, ResourceFilter{}); err != nil || len(got) != 1 {
		t.Fatalf("list resources: resources=%#v err=%v", got, err)
	}
	if err := r.DeleteResource(ctx, "project-1"); err != nil {
		t.Fatalf("delete resource: %v", err)
	}
	dep, err := r.Deploy(ctx, DeploymentConfig{ResourceID: "service-1", Source: &DeploymentSource{Commit: "abc123"}, Config: map[string]any{"environment_id": "env-1"}})
	if err != nil || dep.ID != "service-1-env-1" || dep.State != DeploymentStateDeploying {
		t.Fatalf("deploy: deployment=%#v err=%v", dep, err)
	}
	if estimate, err := r.EstimateCost(ctx, ResourceConfig{}); err != nil || estimate.Currency != "USD" {
		t.Fatalf("estimate cost: estimate=%#v err=%v", estimate, err)
	}
	if _, err := r.Deploy(ctx, DeploymentConfig{}); err == nil {
		t.Error("deploy without resource unexpectedly succeeded")
	}
	if _, err := r.Deploy(ctx, DeploymentConfig{ResourceID: "service-1"}); err == nil {
		t.Error("deploy without environment unexpectedly succeeded")
	}
	for name, call := range map[string]func() error{
		"update":      func() error { _, err := r.UpdateResource(ctx, "project-1", ResourceConfig{}); return err },
		"status":      func() error { _, err := r.GetDeploymentStatus(ctx, "deployment-1"); return err },
		"rollback":    func() error { return r.RollbackDeployment(ctx, "deployment-1") },
		"logs":        func() error { _, err := r.GetLogs(ctx, resource, LogOptions{}); return err },
		"metrics":     func() error { _, err := r.GetMetrics(ctx, resource, MetricOptions{}); return err },
		"actual cost": func() error { _, err := r.GetActualCost(ctx, resource, TimeRange{}); return err },
	} {
		if call() == nil {
			t.Errorf("%s unexpectedly succeeded", name)
		}
	}
	if _, err := r.CreateResource(ctx, ResourceConfig{Name: "edge", Type: ResourceTypeComputeEdge}); err == nil {
		t.Error("unsupported resource unexpectedly succeeded")
	}
	if _, err := NewRailwayProvider(Credentials{}); err == nil {
		t.Error("missing token unexpectedly succeeded")
	}
	r.token = ""
	if err := r.ValidateCredentials(ctx); err == nil {
		t.Error("empty token unexpectedly validated")
	}
}

func TestVercelProviderContract(t *testing.T) {
	ctx := context.Background()
	provider, err := NewVercelProvider(Credentials{Data: map[string]string{"token": "vercel-token", "team_id": "team-1"}})
	if err != nil {
		t.Fatalf("construct provider: %v", err)
	}
	v := provider.(*VercelProvider)
	v.httpClient = providerClient(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/v2/user":
			return contractResponse(req, http.StatusOK, `{"id":"user-1"}`), nil
		case req.Method == http.MethodPost && req.URL.Path == "/v9/projects":
			return contractResponse(req, http.StatusOK, `{"id":"project-1","name":"demo","framework":"nextjs","createdAt":1767225600000,"updatedAt":1767229200000}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v9/projects/project-1":
			return contractResponse(req, http.StatusOK, `{"id":"project-1","name":"demo","framework":"nextjs"}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v9/projects":
			return contractResponse(req, http.StatusOK, `{"projects":[{"id":"project-1","name":"demo","framework":"nextjs"}]}`), nil
		case req.Method == http.MethodDelete && req.URL.Path == "/v9/projects/project-1":
			return contractResponse(req, http.StatusNoContent, ``), nil
		case req.Method == http.MethodPost && req.URL.Path == "/v13/deployments":
			return contractResponse(req, http.StatusOK, `{"id":"deployment-1","url":"demo.vercel.app","name":"demo","readyState":"READY","createdAt":1767225600000}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v13/deployments/deployment-1":
			return contractResponse(req, http.StatusOK, `{"id":"deployment-1","url":"demo.vercel.app","name":"demo","readyState":"READY","createdAt":1767225600000}`), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v2/deployments/deployment-1/events":
			return contractResponse(req, http.StatusOK, "{\"type\":\"stdout\",\"created\":1767225600000,\"payload\":{\"text\":\"built\"}}\n"), nil
		default:
			return contractResponse(req, http.StatusNotFound, `{"error":"unexpected request"}`), nil
		}
	})

	if got := v.GetMetadata(); got.Name != "vercel" || !v.SupportsResource(ResourceTypeComputeFunction) || v.SupportsResource(ResourceTypeComputeVM) || len(v.GetCapabilities()) == 0 {
		t.Fatalf("unexpected metadata or resource support: %#v", got)
	}
	if err := v.Initialize(ctx, Credentials{Data: map[string]string{"token": "vercel-token", "team_id": "team-1"}}); err != nil {
		t.Fatalf("validate credentials: %v", err)
	}
	resource, err := v.CreateResource(ctx, ResourceConfig{
		Name: "demo", Type: ResourceTypeComputeEdge,
		Spec: map[string]any{"framework": "nextjs", "git_repository": "https://github.com/example/demo", "build_command": "npm run build", "output_directory": ".next", "root_directory": "web"},
	})
	if err != nil || resource.ID != "project-1" || resource.Metadata["framework"] != "nextjs" {
		t.Fatalf("create resource: resource=%#v err=%v", resource, err)
	}
	if got, err := v.GetResource(ctx, "project-1"); err != nil || got.Name != "demo" {
		t.Fatalf("get resource: resource=%#v err=%v", got, err)
	}
	if got, err := v.ListResources(ctx, ResourceFilter{}); err != nil || len(got) != 1 {
		t.Fatalf("list resources: resources=%#v err=%v", got, err)
	}
	if err := v.DeleteResource(ctx, "project-1"); err != nil {
		t.Fatalf("delete resource: %v", err)
	}
	dep, err := v.Deploy(ctx, DeploymentConfig{ResourceID: "demo", Source: &DeploymentSource{Repository: "repo-1", Branch: "main", Commit: "abc123"}, Env: map[string]string{"MODE": "prod"}, Config: map[string]any{"target": "production"}})
	if err != nil || dep.ID != "deployment-1" || dep.Message != "https://demo.vercel.app" {
		t.Fatalf("deploy: deployment=%#v err=%v", dep, err)
	}
	status, err := v.GetDeploymentStatus(ctx, "deployment-1")
	if err != nil || status.Health != HealthStatusHealthy {
		t.Fatalf("deployment status: status=%#v err=%v", status, err)
	}
	logs, err := v.GetLogs(ctx, &Resource{ID: "deployment-1"}, LogOptions{})
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	entry, err := logs.Next()
	if err != nil || entry.Message != "built" {
		t.Fatalf("next log: entry=%#v err=%v", entry, err)
	}
	if err := logs.Close(); err != nil {
		t.Fatalf("close logs: %v", err)
	}
	if estimate, err := v.EstimateCost(ctx, ResourceConfig{}); err != nil || estimate.Currency != "USD" {
		t.Fatalf("estimate cost: estimate=%#v err=%v", estimate, err)
	}
	for name, call := range map[string]func() error{
		"update":      func() error { _, err := v.UpdateResource(ctx, "project-1", ResourceConfig{}); return err },
		"rollback":    func() error { return v.RollbackDeployment(ctx, "deployment-1") },
		"metrics":     func() error { _, err := v.GetMetrics(ctx, resource, MetricOptions{}); return err },
		"actual cost": func() error { _, err := v.GetActualCost(ctx, resource, TimeRange{}); return err },
	} {
		if call() == nil {
			t.Errorf("%s unexpectedly succeeded", name)
		}
	}
	if _, err := v.CreateResource(ctx, ResourceConfig{Name: "vm", Type: ResourceTypeComputeVM}); err == nil {
		t.Error("unsupported resource unexpectedly succeeded")
	}
	if _, err := NewVercelProvider(Credentials{}); err == nil {
		t.Error("missing token unexpectedly succeeded")
	}
	v.token = ""
	if err := v.ValidateCredentials(ctx); err == nil {
		t.Error("empty token unexpectedly validated")
	}
}

func TestProviderContractStateMappings(t *testing.T) {
	for _, state := range []string{"ready", "error", "building", "uploading", "uploaded", "new", "pending_review", "enqueued", "unknown"} {
		if netlifyStateToDeploymentState(state) == "" || netlifyStateToHealth(state) == "" {
			t.Fatalf("netlify state %q did not map", state)
		}
	}
	for _, state := range []string{"READY", "ERROR", "BUILDING", "INITIALIZING", "QUEUED", "CANCELED", "UNKNOWN"} {
		if vercelReadyStateToDeploymentState(state) == "" || vercelStateToHealth(state) == "" {
			t.Fatalf("vercel state %q did not map", state)
		}
	}
	if got := netlifySiteToResource(netlifySite{ID: "site", Name: "site"}); got.Provider != "netlify" {
		t.Fatalf("netlify conversion: %#v", got)
	}
	if got := netlifyDeployToCloudDep(netlifyDeploy{ID: "deploy", State: "ready"}); got.State != DeploymentStateActive {
		t.Fatalf("netlify deployment conversion: %#v", got)
	}
	if got := railwayProjectToResource(railwayProject{ID: "project", Name: "project"}); got.Provider != "railway" {
		t.Fatalf("railway conversion: %#v", got)
	}
	if got := vercelProjectToResource(vercelProject{ID: "project", Name: "project"}); got.Provider != "vercel" {
		t.Fatalf("vercel conversion: %#v", got)
	}
	if got := vercelDeploymentToCloudDep(vercelDeployment{ID: "deploy", ReadyState: "READY"}); got.State != DeploymentStateActive {
		t.Fatalf("vercel deployment conversion: %#v", got)
	}
}
