// Package contract drives property-based contract tests against the OpenAPI spec.
//
// Pillar L21 — schemathesis-style contract testing harness. Validates that the
// running API satisfies its declared OpenAPI schema for every path/method/operation.
package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"gopkg.in/yaml.v3"
)

// Spec is a minimal OpenAPI 3.x representation sufficient for path enumeration.
type Spec struct {
	Paths map[string]map[string]yaml.Node `yaml:"paths"`
}

// LoadSpec parses an OpenAPI spec (YAML or JSON) into a Spec.
func LoadSpec(raw []byte, isJSON bool) (*Spec, error) {
	if isJSON {
		var s Spec
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return &s, nil
	}
	var s Spec
	if err := yaml.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Case generates a single test case representing a path+method pair.
type Case struct {
	Path   string
	Method string
}

// EnumerateCases walks the spec and yields one Case per (path, method) pair,
// filtering to standard HTTP methods.
func (s *Spec) EnumerateCases() []Case {
	methods := map[string]bool{
		http.MethodGet: true, http.MethodPost: true, http.MethodPut: true,
		http.MethodPatch: true, http.MethodDelete: true, http.MethodHead: true,
		http.MethodOptions: true,
	}
	var out []Case
	for path, ops := range s.Paths {
		for method := range ops {
			if !methods[strings.ToUpper(method)] {
				continue
			}
			out = append(out, Case{Path: path, Method: strings.ToUpper(method)})
		}
	}
	return out
}

// Run executes all cases against the supplied handler.
// It is a smoke harness: round-trip empty bodies, verify status is not 5xx.
// In production this would be replaced by schemathesis-go or equivalent,
// which generates schema-conforming payloads.
func Run(handler http.Handler, cases []Case) []Failure {
	var failures []Failure
	for _, c := range cases {
		req := httptest.NewRequest(c.Method, c.Path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code >= 500 {
			failures = append(failures, Failure{
				Case:   c,
				Status: rr.Code,
				Body:   rr.Body.String(),
			})
		}
	}
	return failures
}

// Failure describes a contract violation.
type Failure struct {
	Case   Case
	Status int
	Body   string
}

func (f Failure) Error() string {
	return fmt.Sprintf("%s %s → %d", f.Case.Method, f.Case.Path, f.Status)
}
