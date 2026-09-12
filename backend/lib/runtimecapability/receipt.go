// Package runtimecapability validates read-only capability evidence produced
// by PhenoCompose. It is deliberately an ingestion boundary: validation never
// resolves or executes the executable recorded in a receipt.
package runtimecapability

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// SchemaVersion is the wire contract emitted by PhenoCompose's read-only
// runtime capability probe.
const SchemaVersion = "phenocompose.runtime-capability/v1"

// Provider identifies a provider-neutral local container backend.
type Provider string

const (
	ProviderPodman          Provider = "podman"
	ProviderAppleContainers Provider = "apple-containers"
	ProviderWSLContainers   Provider = "wsl-containers"
)

// CommandOutput is the output of one receipt command, in the same order as
// Receipt.Commands. It is supplied separately because PhenoCompose carries
// only the digest, not raw output, in the portable receipt.
type CommandOutput struct {
	Stdout []byte
	Stderr []byte
}

// Receipt is the stable JSON shape emitted by PhenoCompose. Commands contains
// ordered argv vectors; executable is the selected first token and is checked
// against every vector before any downstream consumer can trust the receipt.
type Receipt struct {
	SchemaVersion string     `json:"schema_version"`
	Provider      Provider   `json:"provider"`
	Executable    string     `json:"executable"`
	Commands      [][]string `json:"commands"`
	Ready         bool       `json:"ready"`
	Version       *string    `json:"version,omitempty"`
	OutputSHA256  string     `json:"output_sha256"`
}

var (
	// ErrInvalidReceipt means the receipt cannot be trusted as a capability
	// declaration. Callers should fail closed and not attempt a runtime action.
	ErrInvalidReceipt = errors.New("invalid runtime capability receipt")
	// ErrInvalidEvidence means the supplied raw outputs do not prove the
	// digest carried by an otherwise structurally valid receipt.
	ErrInvalidEvidence = errors.New("invalid runtime capability evidence")
)

// Parse decodes and validates one receipt. Unknown fields, trailing JSON, an
// unsupported schema/provider, and any non-read-only argv are rejected.
func Parse(data []byte) (Receipt, error) {
	var receipt Receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return Receipt{}, invalid("json", err.Error())
	}

	// A second JSON value is never part of a receipt. Checking for EOF also
	// catches non-whitespace trailing bytes instead of silently accepting them.
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Receipt{}, invalid("json", "multiple JSON values")
		}
		return Receipt{}, invalid("json", "trailing data")
	}
	if err := receipt.Validate(); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

// VerifyJSON parses a receipt and, when raw probe output is available at the
// handoff boundary, verifies its canonical ordered stdout/stderr digest.
func VerifyJSON(data []byte, outputs []CommandOutput) (Receipt, error) {
	receipt, err := Parse(data)
	if err != nil {
		return Receipt{}, err
	}
	if err := receipt.VerifyOutputs(outputs); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

// Validate checks receipt structure and read-only command policy. It does not
// claim that output_sha256 matches command output; use VerifyOutputs when raw
// output is available. This distinction is intentional because the portable
// PhenoCompose receipt does not contain raw stdout/stderr.
func (r Receipt) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return invalid("schema_version", fmt.Sprintf("must be %q", SchemaVersion))
	}
	if !isProvider(r.Provider) {
		return invalid("provider", fmt.Sprintf("unsupported provider %q", r.Provider))
	}
	if !isExecutableToken(r.Executable) {
		return invalid("executable", "must be a single executable token without path separators")
	}
	if !r.Ready {
		return invalid("ready", "must be true for an ingestible capability receipt")
	}
	if len(r.Commands) == 0 {
		return invalid("commands", "must contain at least one argv vector")
	}
	for index, argv := range r.Commands {
		if len(argv) == 0 {
			return invalid("commands", fmt.Sprintf("argv[%d] is empty", index))
		}
		if argv[0] != r.Executable {
			return invalid("commands", fmt.Sprintf("argv[%d] executable %q does not match %q", index, argv[0], r.Executable))
		}
		for argIndex, arg := range argv {
			if arg == "" || hasControl(arg) {
				return invalid("commands", fmt.Sprintf("argv[%d][%d] is empty or contains control characters", index, argIndex))
			}
			if isLifecycleToken(arg) {
				return invalid("commands", fmt.Sprintf("argv[%d] contains lifecycle token %q", index, arg))
			}
		}
	}

	switch r.Provider {
	case ProviderPodman:
		if err := validatePodman(r.Executable, r.Commands); err != nil {
			return err
		}
	case ProviderAppleContainers:
		if err := validateAppleContainers(r.Executable, r.Commands); err != nil {
			return err
		}
	case ProviderWSLContainers:
		if err := validateWSLContainers(r.Executable, r.Commands); err != nil {
			return err
		}
	}

	if r.Version != nil && (strings.TrimSpace(*r.Version) == "" || hasControl(*r.Version)) {
		return invalid("version", "must be non-empty and free of control characters when present")
	}
	if !isSHA256(r.OutputSHA256) {
		return invalid("output_sha256", "must be 64 lowercase hexadecimal characters")
	}
	return nil
}

// VerifyOutputs verifies the canonical digest over outputs in command order.
// The algorithm is intentionally byte-level and matches PhenoCompose:
// stdout, NUL, stderr, NUL for every ordered command.
func (r Receipt) VerifyOutputs(outputs []CommandOutput) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if len(outputs) != len(r.Commands) {
		return evidenceError(fmt.Sprintf("output count %d does not match command count %d", len(outputs), len(r.Commands)))
	}
	got := CanonicalOutputSHA256(outputs)
	if subtle.ConstantTimeCompare([]byte(got), []byte(r.OutputSHA256)) != 1 {
		return evidenceError("canonical stdout/stderr digest does not match output_sha256")
	}
	return nil
}

// CanonicalOutputSHA256 returns the digest used by the PhenoCompose receipt
// producer. Output ordering is significant and empty stdout/stderr are still
// represented by their NUL separators.
func CanonicalOutputSHA256(outputs []CommandOutput) string {
	digest := sha256.New()
	for _, output := range outputs {
		_, _ = digest.Write(output.Stdout)
		_, _ = digest.Write([]byte{0})
		_, _ = digest.Write(output.Stderr)
		_, _ = digest.Write([]byte{0})
	}
	return hex.EncodeToString(digest.Sum(nil))
}

func validatePodman(executable string, commands [][]string) error {
	if executable == "podman" || executable == "podman.exe" {
		if len(commands) == 1 && equalArgs(commands[0], executable, "info", "--format", "json") {
			return nil
		}
		return invalid("commands", "podman must use exactly: info --format json")
	}
	if executable != "wsl.exe" {
		return invalid("executable", "podman receipts must select podman, podman.exe, or wsl.exe")
	}
	if len(commands) != 1 || len(commands[0]) != 8 {
		return invalid("commands", "WSL-routed podman must use one fixed read-only argv")
	}
	argv := commands[0]
	if argv[1] != "-d" || argv[2] == "" || argv[3] != "--" || (argv[4] != "podman" && argv[4] != "podman.exe") || !equalArgs(argv[5:], "info", "--format", "json") {
		return invalid("commands", "WSL-routed podman argv is not read-only")
	}
	return nil
}

func validateAppleContainers(executable string, commands [][]string) error {
	if executable != "container" {
		return invalid("executable", "Apple Containers receipts must select container")
	}
	if len(commands) != 2 ||
		!equalArgs(commands[0], "container", "system", "status", "--format", "json") ||
		!equalArgs(commands[1], "container", "system", "version", "--format", "json") {
		return invalid("commands", "Apple Containers must use ordered system status/version JSON probes")
	}
	return nil
}

func validateWSLContainers(executable string, commands [][]string) error {
	if executable != "wslc" && executable != "wslc.exe" && executable != "container.exe" {
		return invalid("executable", "WSLc receipts must select wslc, wslc.exe, or container.exe")
	}
	if len(commands) != 1 || !equalArgs(commands[0], executable, "image", "ls") {
		return invalid("commands", "WSLc must use exactly: image ls")
	}
	return nil
}

func isProvider(provider Provider) bool {
	return provider == ProviderPodman || provider == ProviderAppleContainers || provider == ProviderWSLContainers
}

func isExecutableToken(value string) bool {
	return value != "" && !strings.ContainsAny(value, `/\\`) && !hasControl(value) && !strings.ContainsAny(value, " \t")
}

func hasControl(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func isLifecycleToken(value string) bool {
	switch strings.ToLower(value) {
	case "run", "create", "start", "stop", "rm", "remove", "delete", "kill", "exec", "up", "down":
		return true
	default:
		return false
	}
}

func isSHA256(value string) bool {
	if len(value) != sha256.Size*2 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func equalArgs(got []string, want ...string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func invalid(field, detail string) error {
	return fmt.Errorf("%w: %s: %s", ErrInvalidReceipt, field, detail)
}

func evidenceError(detail string) error {
	return fmt.Errorf("%w: %s", ErrInvalidEvidence, detail)
}
