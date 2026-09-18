package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDriverAndDialectorForSelectsDriver pins the DSN-to-driver mapping. Before
// this existed, data.go hardcoded postgres.Open, so the documented SQLite
// default could never be opened at all.
func TestDriverAndDialectorForSelectsDriver(t *testing.T) {
	cases := []struct {
		name string
		dsn  string
		want string
	}{
		{"postgres URL", "postgres://user:pw@localhost:5432/byteport", "PostgreSQL"},
		{"postgresql URL", "postgresql://user:pw@localhost:5432/byteport", "PostgreSQL"},
		{"libpq key=value", "host=localhost user=postgres dbname=byteport_dev port=5432 sslmode=disable", "PostgreSQL"},
		{"documented default", "file:./byteport.db", "SQLite"},
		{"bare path", "./byteport.db", "SQLite"},
		{"sqlite scheme (docker-compose.yml)", "sqlite:///data/byteport.db", "SQLite"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			driver, dialector := driverAndDialectorFor(tc.dsn)
			if driver != tc.want {
				t.Fatalf("driver = %q, want %q", driver, tc.want)
			}
			if dialector == nil {
				t.Fatal("dialector is nil")
			}
			if !strings.Contains(strings.ToLower(dialector.Name()), strings.ToLower(map[string]string{"PostgreSQL": "postgres", "SQLite": "sqlite"}[tc.want])) {
				t.Fatalf("dialector.Name() = %q, want a %s dialector", dialector.Name(), tc.want)
			}
		})
	}
}

// TestNormalisedSQLitePath keeps the docker-compose.yml DSN usable: SQLite
// expects a file path, not a URL.
func TestNormalisedSQLitePath(t *testing.T) {
	cases := map[string]string{
		"sqlite:///data/byteport.db": "/data/byteport.db",
		"sqlite://byteport.db":       "byteport.db",
		"sqlite:byteport.db":         "byteport.db",
	}
	for dsn, want := range cases {
		if got := normalisedSQLitePath(dsn); got != want {
			t.Fatalf("normalisedSQLitePath(%q) = %q, want %q", dsn, got, want)
		}
	}
}

// TestDatabaseDSNDefaultsToDocumentedSQLitePath covers the documented default:
// with DATABASE_URL unset the server must fall back to the path INSTALL.md,
// DEPLOYMENT.md and docker-compose.yml publish.
func TestDatabaseDSNDefaultsToDocumentedSQLitePath(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	dsn, documentedDefault := databaseDSN()
	if dsn != documentedSQLiteDefault {
		t.Fatalf("dsn = %q, want %q", dsn, documentedSQLiteDefault)
	}
	if !documentedDefault {
		t.Fatal("documentedDefault = false, want true so the caller can log the fallback")
	}

	t.Setenv("DATABASE_URL", "postgres://example/db")
	if dsn, documentedDefault = databaseDSN(); dsn != "postgres://example/db" || documentedDefault {
		t.Fatalf("databaseDSN() = (%q, %v), want the environment value and documentedDefault=false", dsn, documentedDefault)
	}
}

// TestSQLiteAutoMigrateIsPortable is the regression test for the failure that
// made the documented SQLite default unusable:
//
//	near "(": syntax error   on  CREATE TABLE users (... uuid DEFAULT gen_random_uuid() ...)
//
// The schema must migrate on SQLite as well as PostgreSQL, and a created row
// must receive a generated UUID from the application rather than the database.
func TestSQLiteAutoMigrateIsPortable(t *testing.T) {
	t.Setenv("GIN_MODE", "test")
	dsn := filepath.Join(t.TempDir(), "byteport-test.db")

	db, err := connectDatabase(dsn)
	if err != nil {
		t.Fatalf("connectDatabase(%q) failed: %v", dsn, err)
	}

	user := User{Name: "Portable", Email: "portable@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.UUID == "" {
		t.Fatal("created user has an empty UUID: no application-generated primary key")
	}
	if _, err := uuid.Parse(user.UUID); err != nil {
		t.Fatalf("created user UUID %q is not a UUID: %v", user.UUID, err)
	}

	// An explicitly supplied UUID must be preserved, not overwritten.
	explicit := uuid.New().String()
	second := User{UUID: explicit, Name: "Explicit", Email: "explicit@example.com", Password: "hash"}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("create user with explicit UUID: %v", err)
	}
	if second.UUID != explicit {
		t.Fatalf("UUID = %q, want the supplied %q", second.UUID, explicit)
	}

	var stored User
	if err := db.Where("email = ?", "portable@example.com").First(&stored).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if stored.UUID != user.UUID {
		t.Fatalf("stored UUID = %q, want %q", stored.UUID, user.UUID)
	}
}

// TestSQLiteMigratesEveryTableIndividually migrates each model that carries a
// UUID primary key on its own, so a regression in one table cannot be masked by
// another. All six tables below used to fail with `near "(": syntax error`.
//
// WorkOSUser is included even though ConnectDatabase does not migrate it: the
// table is queried by GetUserByWorkOSID/FindOrCreateUserFromWorkOS, so the
// model must at least be migratable on both drivers. Its absence from
// ConnectDatabase's AutoMigrate list is a separate, pre-existing gap.
func TestSQLiteMigratesEveryTableIndividually(t *testing.T) {
	models := []struct {
		name  string
		model any
	}{
		{"User", &User{}},
		{"WorkOSUser", &WorkOSUser{}},
		{"Project", &Project{}},
		{"Instance", &Instance{}},
		{"Deployment", &Deployment{}},
		{"Host", &Host{}},
	}

	for _, tc := range models {
		t.Run(tc.name, func(t *testing.T) {
			dsn := filepath.Join(t.TempDir(), strings.ToLower(tc.name)+".db")
			dialector := sqlite.Open(dsn)
			db, err := gorm.Open(dialector, &gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: true,
			})
			if err != nil {
				t.Fatalf("open sqlite: %v", err)
			}
			if err := db.AutoMigrate(tc.model); err != nil {
				t.Fatalf("AutoMigrate(%s) on SQLite failed: %v", tc.name, err)
			}
		})
	}
}

// TestApplicationGeneratedUUIDsFireOnCreate exercises every model whose UUID
// used to come from the database-side gen_random_uuid() default, so the
// BeforeCreate hooks in primary_keys.go are covered rather than assumed.
func TestApplicationGeneratedUUIDsFireOnCreate(t *testing.T) {
	t.Setenv("GIN_MODE", "test")
	dsn := filepath.Join(t.TempDir(), "uuid-hooks.db")

	db, err := connectDatabase(dsn)
	if err != nil {
		t.Fatalf("connectDatabase(%q) failed: %v", dsn, err)
	}

	owner := User{Name: "Owner", Email: "owner@example.com", Password: "hash"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}

	cases := []struct {
		name  string
		model any
		uuid  func(any) string
	}{
		{"User", &User{Name: "Second", Email: "second@example.com", Password: "hash"},
			func(m any) string { return m.(*User).UUID }},
		{"WorkOSUser", &WorkOSUser{WorkOSID: "workos-1", Name: "WorkOS", Email: "workos@example.com"},
			func(m any) string { return m.(*WorkOSUser).UUID }},
		{"Project", &Project{ID: "proj-1", Owner: owner.UUID, Name: "Project"},
			func(m any) string { return m.(*Project).UUID }},
		{"Instance", &Instance{Owner: owner.UUID, Name: "Instance", Status: "running", ResUUID: "res-1"},
			func(m any) string { return m.(*Instance).UUID }},
		{"Deployment", &Deployment{Name: "Deployment", Owner: owner.UUID},
			func(m any) string { return m.(*Deployment).UUID }},
		{"Host", &Host{Owner: owner.UUID, Name: "Host", HostURL: "http://localhost:9999", APIKey: "key"},
			func(m any) string { return m.(*Host).UUID }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// ConnectDatabase migrates the deployment tables only, so migrate
			// the model under test here. (WorkOSUser is not in that list; see
			// the note on TestSQLiteMigratesEveryTableIndividually.)
			if err := db.AutoMigrate(tc.model); err != nil {
				t.Fatalf("migrate %s: %v", tc.name, err)
			}
			if err := db.Create(tc.model).Error; err != nil {
				t.Fatalf("create %s: %v", tc.name, err)
			}
			generated := tc.uuid(tc.model)
			if generated == "" {
				t.Fatalf("%s was created without a UUID", tc.name)
			}
			if _, err := uuid.Parse(generated); err != nil {
				t.Fatalf("%s UUID %q is not a UUID: %v", tc.name, generated, err)
			}
		})
	}
}

// TestPostgresDSNStillMigrates runs only when DATABASE_URL points at a
// PostgreSQL server, so the PostgreSQL path is verified on demand (see the
// repository docs for the local DSN) without making the unit suite require one.
func TestPostgresDSNStillMigrates(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" || !strings.Contains(dsn, "host=") && !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
		t.Skip("DATABASE_URL is not a PostgreSQL DSN; skipping PostgreSQL migration check")
	}

	driver, _ := driverAndDialectorFor(dsn)
	if driver != "PostgreSQL" {
		t.Fatalf("driver = %q, want PostgreSQL", driver)
	}

	db, err := connectDatabase(dsn)
	if err != nil {
		t.Fatalf("connectDatabase(postgres) failed: %v", err)
	}

	user := User{Name: "Postgres", Email: uuid.New().String() + "@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user on postgres: %v", err)
	}
	if _, err := uuid.Parse(user.UUID); err != nil {
		t.Fatalf("postgres user UUID %q is not a UUID: %v", user.UUID, err)
	}
}
