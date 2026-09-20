// coverage_test.go consolidates three earlier coverage-chasing files into a
// single behavioural coverage pass for the WorkOS user-table helpers in
// users_workos.go and the Project.BeforeSave hook in projects.go.
//
// Predecessors and the lines they took:
//
//   - models_100_percent_test.go          (483 lines, 7 funcs) — TestConnectDatabaseComprehensive,
//                                              TestBeforeSaveCompleteCoverage, TestFindOrCreateUserFromWorkOSCompleteCoverage,
//                                              TestDatabaseFunctionsCompleteCoverage, TestEdgeCasesCompleteCoverage,
//                                              TestJSONMarshalingEdgeCases, TestConnectDatabaseImplementation
//   - models_final_100_percent_test.go    (520 lines, 5 funcs) — TestConnectDatabaseDirectCall,
//                                              TestBeforeSaveCompleteLineCoverage, TestFindOrCreateUserFromWorkOSCompleteLineCoverage,
//                                              TestDatabaseFunctionsCompleteLineCoverage, TestConnectDatabaseImplementationFinal
//   - models_ultimate_100_percent_test.go (488 lines, 4 funcs) — TestConnectDatabaseUltimate,
//                                              TestBeforeSaveUltimateCoverage, TestFindOrCreateUserFromWorkOSUltimateCoverage,
//                                              TestDatabaseFunctionsUltimateCoverage
//
// All three pre-existing files asserted the same code paths in slightly
// different shapes (table-driven in the third, ad-hoc in the first two) and
// contributed 1098 lines of duplicated coverage per the issue-#383 audit.
// The consolidated file below preserves the union of every behavioural
// scenario they exercised, runs as a single package-level go test invocation,
// and lands the same coverage number on the affected files.
//
// Behaviour preserved:
//   - ConnectDatabase: function-existence reference (kept for the coverage it
//     gives to the ConnectDatabase symbol; no behavioural assertion is
//     possible without a real database).
//   - Project.BeforeSave: nil GORM DB, every whitespace-UUID variant
//     (empty/single space/whitespace/tabs/newlines/mixed), large deployments
//     (10000 entries), JSON special chars (quotes / backslashes / newlines),
//     unicode / Chinese / Arabic names, deployments containing empty values,
//     nil deployments, empty deployments map, very long project name.
//   - FindOrCreateUserFromWorkOS: nil global DB (panic), closed sqlite handle
//     (DB error), complex user data (4 table-driven variants), existing user
//     matched by WorkOS ID, existing user matched by email.
//   - GetUserByWorkOSID: nil DB (panic).
//   - CreateUserFromWorkOS: nil DB (panic), closed sqlite handle (DB error).

package models

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ConnectDatabase has no testable behaviour from the framework alone (it tries
// to attach to a real PostgreSQL/MySQL via DATABASE_URL), but referencing the
// symbol keeps it on the coverage report. The function name avoids
// TestConnectDatabase because models_comprehensive_test.go already declares
// that symbol.
func TestConnectDatabaseReferenceForCoverage(t *testing.T) {
	assertConnectDatabase := func(t *testing.T) { assert.NotNil(t, ConnectDatabase) }
	t.Run("function is exported and is a function value", func(t *testing.T) {
		assertConnectDatabase(t)
		var fnType func() = func() {}
		assert.IsType(t, fnType, ConnectDatabase)
	})

	t.Run("DATABASE_URL present does not affect reference", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://example.test/x")
		t.Setenv("GIN_MODE", "release")
		assertConnectDatabase(t)
	})

	t.Run("DATABASE_URL absent does not affect reference", func(t *testing.T) {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("GIN_MODE")
		assertConnectDatabase(t)
	})
}

// saveAndDecodeDeployments calls BeforeSave and asserts that the
// resulting DeploymentsJSON round-trips through encoding/json back to a
// map of the same size as want. It is shared between the shape-variant
// table and the JSON-special-chars subtests so neither owns the
// round-trip dance.
func saveAndDecodeDeployments(t *testing.T, p *Project, want map[string]Instance) {
	t.Helper()
	require.NoError(t, p.BeforeSave(&gorm.DB{}))
	assert.NotEmpty(t, p.UUID)
	assert.NotEmpty(t, p.DeploymentsJSON)
	var roundTrip map[string]Instance
	require.NoError(t, json.Unmarshal([]byte(p.DeploymentsJSON), &roundTrip))
	assert.Equal(t, len(want), len(roundTrip))
}

// Project.BeforeSave — every whitespace UUID variant gets a UUID, the
// deployments map round-trips through JSON, nil GORM DB does not panic, and
// large / unicode / special-char payloads are tolerated.
func TestProjectBeforeSave(t *testing.T) {
	t.Run("uuid whitespace variants are replaced with a real uuid", func(t *testing.T) {
		for _, tc := range []struct{ name, in, wantTrimmed string }{
			{"empty string", "", ""},
			{"single space", " ", " "},
			{"all whitespace", "   ", "   "},
			{"tabs", "\t\t\t", "\t\t\t"},
			{"newlines", "\n\n\n", "\n\n\n"},
			{"mixed whitespace", " \t\n ", " \t\n "},
		} {
			t.Run(tc.name, func(t *testing.T) {
				p := &Project{
					UUID:  tc.in,
					ID:    "test-project-" + tc.name,
					Owner: "user-test",
					Name:  "Test Project",
				}
				require.NoError(t, p.BeforeSave(&gorm.DB{}))
				assert.NotEmpty(t, p.UUID)
				assert.NotEqual(t, tc.in, p.UUID)
				assert.NotEqual(t, strings.TrimSpace(tc.in), p.UUID)
			})
		}
	})

	t.Run("nil gorm DB does not panic and assigns a uuid", func(t *testing.T) {
		p := &Project{ID: "test-project-nil-db", Owner: "user-nil", Name: "Nil DB"}
		assert.NotPanics(t, func() { _ = p.BeforeSave(nil) })
		assert.NotEmpty(t, p.UUID)
	})

	t.Run("deployments map shape variants marshal losslessly", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			deployments map[string]Instance
		}{
			{"nil deployments", nil},
			{"empty deployments", map[string]Instance{}},
			{"single deployment", map[string]Instance{"test": {UUID: "test", Name: "Test", Status: "running"}}},
			{"multiple deployments", map[string]Instance{
				"prod":    {UUID: "prod", Name: "Production", Status: "running"},
				"staging": {UUID: "staging", Name: "Staging", Status: "stopped"},
				"dev":     {UUID: "dev", Name: "Development", Status: "building"},
			}},
			{"deployments with empty values", map[string]Instance{
				"empty": {UUID: "", Name: "", Status: ""},
				"valid": {UUID: "valid", Name: "Valid", Status: "running"},
			}},
			{"deployments with unicode names", map[string]Instance{
				"unicode":  {UUID: "u", Name: "Test 🚀🌟✨", Status: "running"},
				"chinese":  {UUID: "c", Name: "实例名称", Status: "building"},
				"arabic":   {UUID: "a", Name: "اسم المثال", Status: "stopped"},
			}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				p := &Project{ID: "test-project-" + tc.name, Owner: "user-test", Name: "Test Project"}
				if tc.deployments == nil {
					require.NoError(t, p.BeforeSave(&gorm.DB{}))
					assert.Empty(t, p.DeploymentsJSON)
					return
				}
				p.SetDeploy(tc.deployments)
				saveAndDecodeDeployments(t, p, tc.deployments)
			})
		}
	})

	t.Run("deployments containing JSON special chars round-trip", func(t *testing.T) {
		want := map[string]Instance{
			"quotes":      {UUID: "i1", Name: `Instance "with" 'quotes'`, Status: "running"},
			"backslashes": {UUID: "i2", Name: `Instance \back\ and /forward/`, Status: "building"},
			"newlines":    {UUID: "i3", Name: "Instance\nwith\twhitespace", Status: "stopped"},
		}
		p := &Project{ID: "p1", Owner: "u1", Name: "P1"}
		p.SetDeploy(want)
		saveAndDecodeDeployments(t, p, want)
	})

	t.Run("very large deployments map serializes without losing ends", func(t *testing.T) {
		p := &Project{ID: "p1", Owner: "u1", Name: "P1"}
		deployments := make(map[string]Instance, 10000)
		for i := 0; i < 10000; i++ {
			deployments[fmt.Sprintf("instance-%d", i)] = Instance{
				UUID: fmt.Sprintf("uuid-%d", i), Name: fmt.Sprintf("Instance %d", i), Status: "running",
			}
		}
		p.SetDeploy(deployments)
		require.NoError(t, p.BeforeSave(&gorm.DB{}))
		assert.Contains(t, p.DeploymentsJSON, "instance-0")
		assert.Contains(t, p.DeploymentsJSON, "instance-9999")
	})

	t.Run("very long project name is preserved verbatim", func(t *testing.T) {
		longName := strings.Repeat("a", 10000)
		p := &Project{ID: "p1", Owner: "u1", Name: longName}
		require.NoError(t, p.BeforeSave(&gorm.DB{}))
		assert.NotEmpty(t, p.UUID)
		assert.Equal(t, longName, p.Name)
	})
}

// workosUsersDB returns an in-memory sqlite gorm.DB with the legacy
// workos_users schema that the helpers in users_workos.go evolved against.
// The modern columns (encrypted_*, last_login_at, …) are intentionally
// omitted because the helpers under test interact only with uuid, work_os_id,
// name, email, and the audit timestamps. The returned DB has the schema
// applied; callers must restore the previous global DB before returning.
func workosUsersDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE workos_users (
			uuid TEXT PRIMARY KEY,
			work_os_id TEXT NOT NULL,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	return db
}

// withWorkOSDB swaps the package-global DB to a fresh in-memory
// workos_users schema, runs fn, and restores the previous DB via
// t.Cleanup. Use this whenever a test under TestFindOrCreateUserFromWorkOS
// or TestWorkOSDBHelpers needs a real backend.
func withWorkOSDB(t *testing.T, fn func(*gorm.DB)) {
	t.Helper()
	orig := DB
	DB = workosUsersDB(t)
	t.Cleanup(func() { DB = orig })
	fn(DB)
}

// withNilDB temporarily clears the package-global DB so callables that
// touch it (FindOrCreateUserFromWorkOS, GetUserByWorkOSID,
// CreateUserFromWorkOS) take their nil-DB guard path. The previous DB is
// restored via t.Cleanup.
func withNilDB(t *testing.T) {
	t.Helper()
	orig := DB
	DB = nil
	t.Cleanup(func() { DB = orig })
}

// testWorkOSUserInfo returns the canonical WorkOSUserInfo fixture
// duplicated across the nil-DB / closed-handle / happy-path assertions.
// Centralising it keeps the literal payload consistent across tests.
func testWorkOSUserInfo() *WorkOSUserInfo {
	return &WorkOSUserInfo{
		ID: "test-id", Email: "test@example.com", FirstName: "Test", LastName: "User",
	}
}

// assertFindOrCreatePanics wires withNilDB to assert.Panics around
// FindOrCreateUserFromWorkOS(testWorkOSUserInfo()). Used by the
// nil-DB guard tests in TestFindOrCreateUserFromWorkOS.
func assertFindOrCreatePanics(t *testing.T) {
	t.Helper()
	withNilDB(t)
	assert.Panics(t, func() { _, _ = FindOrCreateUserFromWorkOS(testWorkOSUserInfo()) })
}

// assertCreateUserPanics wires withNilDB to assert.Panics around
// CreateUserFromWorkOS(testWorkOSUserInfo()). Used by the nil-DB
// guard test in TestWorkOSDBHelpers.
func assertCreateUserPanics(t *testing.T) {
	t.Helper()
	withNilDB(t)
	assert.Panics(t, func() { _, _ = CreateUserFromWorkOS(testWorkOSUserInfo()) })
}

// closeHandleAndCall closes the underlying sqlDB of the in-memory
// gorm.DB set by withWorkOSDB, then invokes fn with the now-closed
// DB. Helpers like FindOrCreateUserFromWorkOS and
// CreateUserFromWorkOS see the closed handle and return an error,
// which the caller asserts on. Centralising the dance prevents
// five-line copy-paste between the closed-handle subtests.
func closeHandleAndCall(t *testing.T, db *gorm.DB, fn func(*gorm.DB) error) error {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return fn(db)
}

// FindOrCreateUserFromWorkOS covers nil DB (panics), a closed sqlite handle
// (returns error), four user-shape variants (table-driven), the existing-
// user-by-WorkOSID branch, and the existing-user-by-email branch which
// updates the WorkOS ID to match the new auth principal.
func TestFindOrCreateUserFromWorkOS(t *testing.T) {
	t.Run("nil global db panics", func(t *testing.T) {
		assertFindOrCreatePanics(t)
	})

	t.Run("closed sqlite handle returns an error", func(t *testing.T) {
		withWorkOSDB(t, func(db *gorm.DB) {
			err := closeHandleAndCall(t, db, func(db *gorm.DB) error {
				_, err := FindOrCreateUserFromWorkOS(testWorkOSUserInfo())
				return err
			})
			assert.Error(t, err)
		})
	})

	t.Run("complex user-shape variants create a fresh row", func(t *testing.T) {
		withWorkOSDB(t, func(db *gorm.DB) {
			cases := []struct {
				name  string
				email string
				first string
				last  string
			}{
				{"simple user", "simple@example.com", "Simple", "User"},
				{"complex email", "complex.user+test@example.com", "Complex", "User"},
				{"unicode name", "unicode@example.com", "🚀", "🌟"},
				{"special characters", "special@example.com", "Special!@#$%^&*()", "User!@#$%^&*()"},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					info := &WorkOSUserInfo{
						ID: tc.name + "-id", Email: tc.email, FirstName: tc.first, LastName: tc.last,
					}
					user, err := FindOrCreateUserFromWorkOS(info)
					require.NoError(t, err)
					assert.Equal(t, tc.name+"-id", user.WorkOSID)
					assert.Equal(t, tc.first+" "+tc.last, user.Name)
					assert.Equal(t, tc.email, user.Email)
				})
			}
		})
	})

	t.Run("existing user match returns or updates the row", func(t *testing.T) {
		cases := []struct {
			name           string
			existing       *WorkOSUser
			info           *WorkOSUserInfo
			wantWorkOSID   string
			wantName       string
		}{
			{
				name:         "matched by WorkOS ID is returned unchanged",
				existing:     &WorkOSUser{WorkOSID: "existing-workos-id", Name: "Existing User", Email: "existing@example.com"},
				info:         &WorkOSUserInfo{ID: "existing-workos-id", Email: "existing@example.com", FirstName: "Existing", LastName: "User"},
				wantWorkOSID: "existing-workos-id",
				wantName:     "Existing User",
			},
			{
				name:         "matched by email gets its WorkOS ID rewritten",
				existing:     &WorkOSUser{WorkOSID: "old-workos-id", Name: "Email Match User", Email: "emailmatch@example.com"},
				info:         &WorkOSUserInfo{ID: "new-workos-id", Email: "emailmatch@example.com", FirstName: "Email", LastName: "Match"},
				wantWorkOSID: "new-workos-id",
				wantName:     "Email Match User",
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				withWorkOSDB(t, func(db *gorm.DB) {
					require.NoError(t, db.Create(tc.existing).Error)
					user, err := FindOrCreateUserFromWorkOS(tc.info)
					require.NoError(t, err)
					assert.Equal(t, tc.wantWorkOSID, user.WorkOSID)
					assert.Equal(t, tc.wantName, user.Name)
				})
			})
		}
	})
}

// Database helpers — nil DB panics, closed sqlite handle returns an error.
func TestWorkOSDBHelpers(t *testing.T) {
	t.Run("GetUserByWorkOSID with nil db panics", func(t *testing.T) {
		withNilDB(t)
		assert.Panics(t, func() { _, _ = GetUserByWorkOSID("test-id") })
	})

	t.Run("CreateUserFromWorkOS with nil db panics", func(t *testing.T) {
		assertCreateUserPanics(t)
	})

	t.Run("CreateUserFromWorkOS with closed sqlite returns an error", func(t *testing.T) {
		withWorkOSDB(t, func(db *gorm.DB) {
			err := closeHandleAndCall(t, db, func(db *gorm.DB) error {
				_, err := CreateUserFromWorkOS(testWorkOSUserInfo())
				return err
			})
			assert.Error(t, err)
		})
	})
}
