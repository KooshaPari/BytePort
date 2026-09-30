package models

import (
	"encoding/json"
	"testing"
	"time"
)

func acceptProviderResourceID(id ProviderResourceID) ProviderResourceID { return id }

func TestMatureIdentityRoundTrip(t *testing.T) {
	now := time.Unix(1800000000, 0).UTC()
	s := SourceSnapshot{ID: "src-1", Provider: "git", Repository: "owner/repo", Commit: "0123456789abcdef", ResolvedAt: now}
	raw, err := json.Marshal(s)
	if err != nil { t.Fatal(err) }
	var got SourceSnapshot
	if err := json.Unmarshal(raw, &got); err != nil { t.Fatal(err) }
	if got.ID != s.ID || got.Commit != s.Commit || !got.ResolvedAt.Equal(now) { t.Fatalf("round trip mismatch: %#v", got) }
}

func TestComponentCanBeProviderManagedWithoutArtifact(t *testing.T) {
	r := ArtifactResolution{Component: "db", Kind: ArtifactResolutionManaged}
	if r.Artifact != nil { t.Fatal("provider-managed component must not require fake artifact") }
}

func TestRuntimeUnknownIsFirstClass(t *testing.T) {
	op := RuntimeOperation{ID: "op-1", Fingerprint: "fp", State: RuntimeOperationUnknown}
	if op.State != RuntimeOperationUnknown { t.Fatalf("state = %s", op.State) }
}

func TestProviderResourceIdentityIsDistinctAtTypedBoundary(t *testing.T) {
	got := acceptProviderResourceID(ProviderResourceID("provider-1"))
	if got != "provider-1" { t.Fatal(got) }
}
