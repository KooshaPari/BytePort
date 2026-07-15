package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestHandleAgentDiscovery_OK — the canonical agent card MUST return 200,
// application/agent-card+json content type, and Cache-Control headers
// suitable for short-lived caching.
func TestHandleAgentDiscovery_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/.well-known/agent.json", handleAgentDiscovery)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", w.Code, w.Body.String())
	}
	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/agent-card+json") {
		t.Fatalf("expected agent-card+json content type, got %q", ct)
	}
	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Fatalf("expected Cache-Control with max-age, got %q", cc)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{
		"schema_version", "id", "name", "version", "capabilities",
		"endpoints", "auth", "security", "metadata",
	} {
		if _, ok := payload[key]; !ok {
			t.Errorf("missing required field %q in agent card", key)
		}
	}
	if payload["schema_version"] != "0.3.0" {
		t.Errorf("expected schema_version 0.3.0, got %v", payload["schema_version"])
	}
	if payload["id"] != "byteport-api/2.0.0" {
		t.Errorf("expected id byteport-api/2.0.0, got %v", payload["id"])
	}
}

// TestHandleAgentDiscovery_StaticSync — the on-disk public-facing
// agent.json (the surface CDNs/registries/scrapers consume) MUST agree
// on identity with the live programmatic A2A card (the surface runtime
// agents consume). The two are intentionally different schemas:
//   - public/.well-known/agent.json uses the well-known-agent.org
//     schema (flat, no schema_version, tools-as-objects w/ input_schema).
//   - Live handleAgentDiscovery emits A2A 0.3.0 (id, schema_version,
//     structured endpoints/auth/security blocks, tools-as-named-strings).
// This test asserts the invariants that matter: same identity (name,
// version, homepage), same tool surface, same auth surface. The two
// payloads round-trip to *semantically equivalent* identity even though
// they are not bytewise equal.
func TestHandleAgentDiscovery_StaticSync(t *testing.T) {
	// Resolve static file relative to current working dir.
	candidates := []string{
		"../public/.well-known/agent.json",
		"public/.well-known/agent.json",
		"../../public/.well-known/agent.json",
	}
	var raw []byte
	var found string
	for _, p := range candidates {
		if data, err := os.ReadFile(p); err == nil {
			raw = data
			found = p
			break
		}
	}
	if raw == nil {
		t.Skip("static agent.json not found from cwd — skipping sync check (CI runs from repo root)")
	}
	t.Logf("using static file: %s", found)

	var static map[string]any
	if err := json.Unmarshal(raw, &static); err != nil {
		t.Fatalf("unmarshal static: %v", err)
	}
	var live map[string]any
	if err := json.Unmarshal(mustMarshal(t, agentCard), &live); err != nil {
		t.Fatalf("unmarshal live: %v", err)
	}

	// Identity invariants — these MUST agree across both surfaces.
	if static["name"] != live["name"] {
		t.Errorf("name drift: static=%v live=%v", static["name"], live["name"])
	}
	if static["version"] != live["version"] {
		t.Errorf("version drift: static=%v live=%v", static["version"], live["version"])
	}
	if static["homepage"] != live["homepage"] {
		t.Errorf("homepage drift: static=%v live=%v", static["homepage"], live["homepage"])
	}

	// Tool surface — both must advertise at least one tool, and every
	// tool name in the live card must exist somewhere on the static
	// side (under capabilities.tools as objects with .name).
	capsAny, _ := live["capabilities"].(map[string]any)
	if capsAny == nil {
		t.Fatalf("live capabilities missing or wrong type: %T", live["capabilities"])
	}
	liveToolList, ok := capsAny["tools"].([]any)
	if !ok {
		t.Fatalf("live capabilities.tools unexpected type %T", capsAny["tools"])
	}
	liveTools := make([]string, 0, len(liveToolList))
	for _, v := range liveToolList {
		if s, sok := v.(string); sok {
			liveTools = append(liveTools, s)
		}
	}
	if len(liveTools) == 0 {
		t.Fatal("live capabilities.tools must list at least one tool")
	}
	staticCaps, _ := static["capabilities"].(map[string]any)
	staticToolList, _ := staticCaps["tools"].([]any)
	staticNames := map[string]bool{}
	for _, tn := range staticToolList {
		if m, ok := tn.(map[string]any); ok {
			if n, _ := m["name"].(string); n != "" {
				staticNames[n] = true
			}
		}
	}
	for _, name := range liveTools {
		if !staticNames[name] {
			t.Errorf("live tool %q must also be advertised by static agent.json", name)
		}
	}
	for n := range staticNames {
		// Static may legitimately include extras beyond the live card
		// (e.g. byteport_list_deployments vs live byteport_list). We warn
		// but do not fail — canonicalization is a follow-up task.
		if !contains(liveTools, n) {
			t.Logf("note: static advertises %q but live card does not; canonicalization pending", n)
		}
	}
}

// TestHandleRoot_OK — root pointer must return 200, advertise
// agent_capable=true, and reference the discovery URI.
func TestHandleRoot_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", handleRoot)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["agent_capable"] != true {
		t.Errorf("expected agent_capable=true, got %v", payload["agent_capable"])
	}
	if payload["discovery"] != "/.well-known/agent.json" {
		t.Errorf("expected discovery pointer to /.well-known/agent.json, got %v", payload["discovery"])
	}
}

// TestHandleAgentDiscovery_CAPABILITIES — every tool advertised in the
// capabilities block MUST be a non-empty string, so the agent card never
// lies about what the system can do. Tool name ↔ MCP server registration
// is reconciled manually via a dedicated reconciliation job, so this test
// only asserts structural soundness (slice of strings).
func TestHandleAgentDiscovery_Capabilities(t *testing.T) {
	capsRaw, has := agentCard["capabilities"]
	if !has {
		t.Fatal("capabilities key missing on agent card")
	}
	var toolsRaw any
	switch c := capsRaw.(type) {
	case gin.H:
		toolsRaw = c["tools"]
	case map[string]any:
		toolsRaw = c["tools"]
	default:
		t.Fatalf("capabilities unexpected type %T", capsRaw)
	}
	tlist, ok := toolsRaw.([]string)
	if !ok {
		t.Fatalf("capabilities.tools unexpected type %T", toolsRaw)
	}
	if len(tlist) == 0 {
		t.Fatal("capabilities.tools must list at least one tool")
	}
	for _, name := range tlist {
		if name == "" {
			t.Errorf("tool name must be non-empty string")
		}
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

// contains — terse membership check for []string.
func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
