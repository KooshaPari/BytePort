package models

import "testing"

func TestBuildEngineCapabilitiesAreNotBrandIdentity(t *testing.T) {
	c := BuildEngineCapabilities{EngineID:"fixture", Version:"1", SupportsExplicitBuild:true, Provenance:true}
	if !c.SupportsExplicitBuild || !c.Provenance { t.Fatal(c) }
}

func TestRuntimeObservationCanRepresentVisibilityUncertainty(t *testing.T) {
	r := RuntimeObservationResult{Found:false, VisibilityUncertain:true}
	if r.Found || !r.VisibilityUncertain { t.Fatal(r) }
}
