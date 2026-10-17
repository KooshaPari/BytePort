package models

import "testing"

func TestLegacyImportDoesNotFabricateVerifiedIdentity(t *testing.T) {
	row := LegacyDeploymentImport{
		ProjectID: "project-1",
		Provider: "nvms",
		ExternalID: "sandbox-1",
		Verification: LegacyVerificationUnverified,
	}
	if row.SourceSnapshot != nil || row.ManifestRevision != nil || row.BuildArtifact != nil {
		t.Fatal("legacy import fabricated mature verification identity")
	}
	if row.Verification != LegacyVerificationUnverified { t.Fatal(row.Verification) }
}
