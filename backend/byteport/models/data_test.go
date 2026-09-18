package models

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSQLiteFilePathForStripsSchemes keeps the documented DSN shapes usable:
// SQLite expects a file path, and docker-compose.yml passes a "sqlite://" URL.
func TestSQLiteFilePathForStripsSchemes(t *testing.T) {
	cases := map[string]string{
		"file:./byteport.db":         "./byteport.db",
		"sqlite:///data/byteport.db": "/data/byteport.db",
		"sqlite://byteport.db":       "byteport.db",
		"sqlite:byteport.db":         "byteport.db",
		"./byteport.db":              "./byteport.db",
	}
	for dsn, want := range cases {
		if got := sqliteFilePathFor(dsn); got != want {
			t.Fatalf("sqliteFilePathFor(%q) = %q, want %q", dsn, got, want)
		}
	}
}

// TestSQLiteDSNForKeepsDriverUsableSchemes covers what is actually handed to
// gorm.io/driver/sqlite: "file:" is understood by the driver, other schemes
// are not.
func TestSQLiteDSNForKeepsDriverUsableSchemes(t *testing.T) {
	cases := map[string]string{
		"file:./byteport.db":         "file:./byteport.db",
		"sqlite:///data/byteport.db": "/data/byteport.db",
		"./byteport.db":              "./byteport.db",
	}
	for dsn, want := range cases {
		if got := sqliteDSNFor(dsn); got != want {
			t.Fatalf("sqliteDSNFor(%q) = %q, want %q", dsn, got, want)
		}
	}
}

// TestResolveDSNPrefersTheEnvironmentAndKeepsLegacyData covers the resolution
// order this server previously lacked: DATABASE_URL was ignored outright and
// data.go always opened "../database.db".
func TestResolveDSNPrefersTheEnvironmentAndKeepsLegacyData(t *testing.T) {
	t.Run("DATABASE_URL wins", func(t *testing.T) {
		t.Chdir(t.TempDir())
		t.Setenv("DATABASE_URL", "file:./custom.db")

		dsn, documentedDefault := resolveDSN()
		if dsn != "file:./custom.db" || documentedDefault {
			t.Fatalf("resolveDSN() = (%q, %v), want the environment value with documentedDefault=false", dsn, documentedDefault)
		}
	})

	t.Run("documented default when nothing exists", func(t *testing.T) {
		t.Chdir(t.TempDir())
		t.Setenv("DATABASE_URL", "")

		dsn, documentedDefault := resolveDSN()
		if dsn != documentedSQLiteDefault || !documentedDefault {
			t.Fatalf("resolveDSN() = (%q, %v), want (%q, true)", dsn, documentedDefault, documentedSQLiteDefault)
		}
	})

	t.Run("legacy database is kept when the documented file is absent", func(t *testing.T) {
		root := t.TempDir()
		subdir := filepath.Join(root, "byteport")
		if err := os.Mkdir(subdir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "database.db"), []byte("legacy"), 0o600); err != nil {
			t.Fatalf("seed legacy db: %v", err)
		}
		t.Chdir(subdir)
		t.Setenv("DATABASE_URL", "")

		dsn, documentedDefault := resolveDSN()
		if dsn != legacySQLitePath || !documentedDefault {
			t.Fatalf("resolveDSN() = (%q, %v), want (%q, true)", dsn, documentedDefault, legacySQLitePath)
		}
	})

	t.Run("documented file wins once it exists", func(t *testing.T) {
		root := t.TempDir()
		subdir := filepath.Join(root, "byteport")
		if err := os.Mkdir(subdir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "database.db"), []byte("legacy"), 0o600); err != nil {
			t.Fatalf("seed legacy db: %v", err)
		}
		if err := os.WriteFile(filepath.Join(subdir, "byteport.db"), []byte("documented"), 0o600); err != nil {
			t.Fatalf("seed documented db: %v", err)
		}
		t.Chdir(subdir)
		t.Setenv("DATABASE_URL", "")

		dsn, documentedDefault := resolveDSN()
		if dsn != documentedSQLiteDefault || !documentedDefault {
			t.Fatalf("resolveDSN() = (%q, %v), want (%q, true)", dsn, documentedDefault, documentedSQLiteDefault)
		}
	})
}

// TestConnectDatabaseMigratesTheDocumentedDefault migrates a fresh database at
// the documented default path, which is what a first run of this server does.
func TestConnectDatabaseMigratesTheDocumentedDefault(t *testing.T) {
	t.Chdir(t.TempDir())

	dsn := documentedSQLiteDefault
	db, err := connectDatabase(dsn)
	if err != nil {
		t.Fatalf("connectDatabase(%q) failed: %v", dsn, err)
	}

	for _, table := range []string{"users", "projects", "instances", "git_secrets"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("table %q was not migrated", table)
		}
	}

	// The primary key is text on this server, so no database-side UUID default
	// is involved; a row must still round-trip.
	user := User{UUID: "11111111-2222-3333-4444-555555555555", Name: "Local", Email: "local@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	var stored User
	if err := db.Where("email = ?", "local@example.com").First(&stored).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if stored.UUID != user.UUID {
		t.Fatalf("stored UUID = %q, want %q", stored.UUID, user.UUID)
	}

	if _, err := os.Stat(sqliteFilePathFor(dsn)); err != nil {
		t.Fatalf("documented default file was not created: %v", err)
	}
}

// TestConnectDatabaseRejectsPostgresDSN states the boundary plainly: this
// server has no PostgreSQL driver, so it must fail with an explanation instead
// of opening a file called "postgres://...".
func TestConnectDatabaseRejectsPostgresDSN(t *testing.T) {
	postgresDSNs := []string{
		"postgres://user:pw@localhost:5432/byteport",
		"postgresql://user:pw@localhost:5432/byteport",
		"host=localhost user=postgres dbname=byteport_dev port=5432 sslmode=disable",
	}
	for _, dsn := range postgresDSNs {
		if _, err := connectDatabase(dsn); err == nil {
			t.Fatalf("connectDatabase(%q) succeeded, want an explicit SQLite-only error", dsn)
		}
	}
}
