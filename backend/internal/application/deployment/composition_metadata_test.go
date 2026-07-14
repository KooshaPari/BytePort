package deployment

import (
	"strings"
	"testing"
)

func TestValidateCompositionMetadata(t *testing.T) {
	valid := "sha256:" + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := validateCompositionMetadata(valid, "oci://registry.example/app@sha256:abc"); err != nil {
		t.Fatalf("valid metadata rejected: %v", err)
	}
	for _, digest := range []string{"sha256:abc", "sha1:" + strings.Repeat("0", 64), "sha256:" + strings.Repeat("g", 64)} {
		if err := validateCompositionMetadata(digest, ""); err == nil {
			t.Errorf("digest %q unexpectedly accepted", digest)
		}
	}
	if err := validateCompositionMetadata("", "bad\nref"); err == nil {
		t.Fatal("control character in artifact ref unexpectedly accepted")
	}
}
