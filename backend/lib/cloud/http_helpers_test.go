package cloud

import (
	"strings"
	"testing"
)

func TestReadProviderErrorBodyCapsLargeResponses(t *testing.T) {
	got := readProviderErrorBody(strings.NewReader(strings.Repeat("x", maxProviderErrorBody+128)))
	if len(got) != maxProviderErrorBody+len("...[truncated]") {
		t.Fatalf("got length %d, want %d", len(got), maxProviderErrorBody+len("...[truncated]"))
	}
	if !strings.HasSuffix(got, "...[truncated]") {
		t.Fatalf("missing truncation marker")
	}
}

func TestReadProviderErrorBodyPreservesSmallResponses(t *testing.T) {
	const want = "provider rejected request"
	if got := readProviderErrorBody(strings.NewReader(want)); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
