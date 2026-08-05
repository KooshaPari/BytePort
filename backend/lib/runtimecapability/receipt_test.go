package runtimecapability

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParseAndVerifyPodmanReceipt(t *testing.T) {
	outputs := []CommandOutput{{Stdout: []byte(`{"version":{"Version":"5.8.4"}}`), Stderr: []byte{}}}
	version := "5.8.4"
	receipt := Receipt{
		SchemaVersion: SchemaVersion,
		Provider:      ProviderPodman,
		Executable:    "podman",
		Commands:      [][]string{{"podman", "info", "--format", "json"}},
		Ready:         true,
		Version:       &version,
		OutputSHA256:  CanonicalOutputSHA256(outputs),
	}
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := VerifyJSON(data, outputs)
	if err != nil {
		t.Fatalf("VerifyJSON() error = %v", err)
	}
	if parsed.Provider != ProviderPodman || parsed.Version == nil || *parsed.Version != version {
		t.Fatalf("unexpected receipt: %+v", parsed)
	}
}

func TestValidateAcceptsProviderSpecificReadOnlyReceipts(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		exec     string
		commands [][]string
	}{
		{
			name:     "podman direct",
			provider: ProviderPodman,
			exec:     "podman",
			commands: [][]string{{"podman", "info", "--format", "json"}},
		},
		{
			name:     "podman through WSL",
			provider: ProviderPodman,
			exec:     "wsl.exe",
			commands: [][]string{{"wsl.exe", "-d", "FedoraLinux-44", "--", "podman", "info", "--format", "json"}},
		},
		{
			name:     "apple",
			provider: ProviderAppleContainers,
			exec:     "container",
			commands: [][]string{{"container", "system", "status", "--format", "json"}, {"container", "system", "version", "--format", "json"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := validReceipt(test.provider, test.exec, test.commands)
			if err := r.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
	for _, executable := range []string{"wslc", "wslc.exe", "container.exe"} {
		t.Run("wslc/"+executable, func(t *testing.T) {
			r := validReceipt(ProviderWSLContainers, executable, [][]string{{executable, "image", "ls"}})
			if err := r.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestValidateFailsClosedForSchemaProviderAndLifecycle(t *testing.T) {
	base := validReceipt(ProviderWSLContainers, "wslc.exe", [][]string{{"wslc.exe", "image", "ls"}})
	tests := []struct {
		name  string
		mutate func(*Receipt)
	}{
		{name: "schema", mutate: func(r *Receipt) { r.SchemaVersion = "phenocompose.runtime-capability/v2" }},
		{name: "provider", mutate: func(r *Receipt) { r.Provider = "docker" }},
		{name: "not ready", mutate: func(r *Receipt) { r.Ready = false }},
		{name: "path executable", mutate: func(r *Receipt) { r.Executable = `C:\\tools\\wslc.exe` }},
		{name: "lifecycle token", mutate: func(r *Receipt) { r.Commands = [][]string{{"wslc.exe", "container", "stop", "id"}} }},
		{name: "wrong argv", mutate: func(r *Receipt) { r.Commands = [][]string{{"wslc.exe", "image", "rm"}} }},
		{name: "bad digest", mutate: func(r *Receipt) { r.OutputSHA256 = strings.Repeat("A", 64) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := base
			test.mutate(&r)
			if err := r.Validate(); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("Validate() error = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func TestParseRejectsUnknownAndTrailingFields(t *testing.T) {
	r := validReceipt(ProviderPodman, "podman", [][]string{{"podman", "info", "--format", "json"}})
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	unknown := strings.TrimSuffix(string(data), "}") + `,"unexpected":true}`
	if _, err := Parse([]byte(unknown)); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("unknown field error = %v", err)
	}
	if _, err := Parse(append(data, []byte(` {}`)...)); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("trailing JSON error = %v", err)
	}
}

func TestCanonicalDigestIsOrderedAndVerifyFailsClosed(t *testing.T) {
	first := []CommandOutput{{Stdout: []byte("one"), Stderr: []byte("warn")}, {Stdout: []byte("two"), Stderr: nil}}
	second := []CommandOutput{first[1], first[0]}
	if CanonicalOutputSHA256(first) == CanonicalOutputSHA256(second) {
		t.Fatal("canonical digest ignored output order")
	}
	r := validReceipt(ProviderAppleContainers, "container", [][]string{{"container", "system", "status", "--format", "json"}, {"container", "system", "version", "--format", "json"}})
	r.OutputSHA256 = CanonicalOutputSHA256(first)
	if err := r.VerifyOutputs(second); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("VerifyOutputs() error = %v, want ErrInvalidEvidence", err)
	}
	if err := r.VerifyOutputs(first[:1]); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("count mismatch error = %v, want ErrInvalidEvidence", err)
	}
}

func validReceipt(provider Provider, executable string, commands [][]string) Receipt {
	return Receipt{
		SchemaVersion: SchemaVersion,
		Provider:      provider,
		Executable:    executable,
		Commands:      commands,
		Ready:         true,
		OutputSHA256:  strings.Repeat("0", 64),
	}
}
