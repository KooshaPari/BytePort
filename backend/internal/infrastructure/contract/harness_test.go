// Package contract_test contains tests for the contract testing harness.
package contract_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	contract "github.com/byteport/api/internal/infrastructure/contract"
)

func TestContract_LoadSpec_JSON(t *testing.T) {
	raw := []byte(`{
		"openapi": "3.0.0",
		"paths": {"/v1/deployments": {"get": {"summary": "list"}}}
	}`)
	spec, err := contract.LoadSpec(raw, true)
	if err != nil {
		t.Fatalf("load json spec: %v", err)
	}
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if len(spec.Paths) == 0 {
		t.Fatal("expected 1 path in spec")
	}
}

func TestContract_LoadSpec_InvalidJSON(t *testing.T) {
	_, err := contract.LoadSpec([]byte(`{invalid json`), true)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestContract_LoadSpec_Empty(t *testing.T) {
	_, err := contract.LoadSpec([]byte{}, true)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestContract_LoadSpec_YAMLAliasType(t *testing.T) {
	// Verify the YAMLNode exported alias is usable.
	var n contract.YAMLNode
	n.Kind = 0
	if n.Kind != 0 {
		t.Errorf("unexpected yaml node kind: %d", n.Kind)
	}
}

func TestContract_EnumerateCases_Defaults(t *testing.T) {
	raw := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "BytePort", "version": "0.1.0"},
		"paths": {}
	}`)
	spec, err := contract.LoadSpec(raw, true)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	cases := spec.EnumerateCases()
	// Empty paths returns nil slice — that's valid, baseline expectation.
	if len(cases) != 0 {
		t.Fatalf("expected 0 cases for empty paths, got %d", len(cases))
	}
}

func TestContract_EnumerateCases_NonEmpty(t *testing.T) {
	raw := []byte(`{
		"openapi": "3.0.0",
		"paths": {
			"/v1/deployments": {
				"get":  {"summary": "list"},
				"post": {"summary": "create"}
			}
		}
	}`)
	spec, err := contract.LoadSpec(raw, true)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	cases := spec.EnumerateCases()
	if len(cases) != 2 {
		t.Fatalf("expected 2 cases (GET+POST), got %d", len(cases))
	}
	for _, c := range cases {
		if c.Path != "/v1/deployments" {
			t.Errorf("unexpected path: %s", c.Path)
		}
	}
}

func TestContract_Failure_Contract(t *testing.T) {
	f := contract.Failure{
		Case:   contract.Case{Method: "GET", Path: "/v1/deployments"},
		Status: 500,
		Body:   "schema drift",
	}
	if f.Error() == "" {
		t.Fatal("Failure.Error() must not be empty")
	}
	if !strings.Contains(f.Error(), "/v1/deployments") {
		t.Errorf("Failure.Error() should reference path, got %q", f.Error())
	}
	if !strings.Contains(f.Error(), "500") {
		t.Errorf("Failure.Error() should reference status, got %q", f.Error())
	}
	// Note: Failure.Error() intentionally summarizes Case + Status only;
	// the body is preserved on Failure.Body for forensic inspection but
	// not echoed in the message to keep error strings compact.
	_ = contract.Failure{}
}

func TestContract_Run_AllPass(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	cases := []contract.Case{
		{Method: "GET", Path: "/"},
	}
	fails := contract.Run(handler, cases)
	if len(fails) != 0 {
		t.Fatalf("expected 0 failures, got %d: %v", len(fails), fails)
	}
}

func TestContract_Run_FailureDetected(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	cases := []contract.Case{
		{Method: "GET", Path: "/"},
	}
	fails := contract.Run(handler, cases)
	if len(fails) == 0 {
		t.Fatal("expected at least 1 failure when server returns 500")
	}
}

func TestContract_Spec_JSONRoundTrip(t *testing.T) {
	raw := []byte(`{"openapi":"3.0.0","paths":{}}`)
	spec, err := contract.LoadSpec(raw, true)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Struct field is `Paths` so marshal produces `{"Paths":{...}}`.
	if !strings.Contains(string(encoded), "Paths") {
		t.Errorf("round-trip lost Paths field: %s", encoded)
	}
}
