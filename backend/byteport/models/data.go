package models

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// documentedSQLiteDefault is the DSN published for this server in INSTALL.md,
// DEPLOYMENT.md and docker-compose.yml. Keep it in sync with those files.
const documentedSQLiteDefault = "file:./byteport.db"

// legacySQLitePath is where this server kept its database before DATABASE_URL
// was honoured. The process is started from backend/byteport (see the `start`
// script and setup-windows.ps1), so this resolves to backend/database.db.
const legacySQLitePath = "../database.db"

var DB *gorm.DB

// sqliteFilePathFor returns the filesystem path a DATABASE_URL value points at,
// with any scheme removed: "file:./byteport.db" and
// "sqlite:///data/byteport.db" become "./byteport.db" and "/data/byteport.db".
func sqliteFilePathFor(dsn string) string {
	path := strings.TrimPrefix(dsn, "file:")
	path = strings.TrimPrefix(path, "sqlite://")
	if path == dsn {
		path = strings.TrimPrefix(dsn, "sqlite:")
	}
	path = strings.TrimPrefix(path, "//")
	if path == "" {
		return dsn
	}
	return path
}

// sqliteDSNFor returns the DSN to hand the SQLite driver. The "file:" scheme is
// understood by the driver and is passed through unchanged; a "sqlite://"
// prefix (used by docker-compose.yml) is stripped, because the driver would
// otherwise treat the whole string as a filename.
func sqliteDSNFor(dsn string) string {
	if strings.HasPrefix(dsn, "file:") {
		return dsn
	}
	return sqliteFilePathFor(dsn)
}

// isPostgresDSN reports whether the DSN selects a PostgreSQL server.
func isPostgresDSN(dsn string) bool {
	return strings.HasPrefix(dsn, "postgres://") ||
		strings.HasPrefix(dsn, "postgresql://") ||
		strings.Contains(dsn, "host=")
}

// resolveDSN returns the DSN to connect with and whether it came from the
// documented default rather than from DATABASE_URL.
//
// DATABASE_URL was previously ignored by this server entirely: data.go opened a
// hard-coded "../database.db", so the documented variable had no effect on the
// server INSTALL.md/DEPLOYMENT.md/docker-compose.yml describe. When the
// documented default file does not exist yet but a legacy database does, the
// legacy database is kept so an existing install does not silently start
// against an empty one; the caller logs which path was chosen.
func resolveDSN() (dsn string, documentedDefault bool) {
	if value := strings.TrimSpace(os.Getenv("DATABASE_URL")); value != "" {
		return value, false
	}

	if _, err := os.Stat(sqliteFilePathFor(documentedSQLiteDefault)); err != nil {
		if _, legacyErr := os.Stat(legacySQLitePath); legacyErr == nil {
			return legacySQLitePath, true
		}
	}
	return documentedSQLiteDefault, true
}

// descriptionOf renders an absolute path for logging, so the relative DSNs do
// not hide which file is actually in use.
func descriptionOf(dsn string) string {
	absolute, err := filepath.Abs(sqliteFilePathFor(dsn))
	if err != nil {
		return dsn
	}
	return absolute
}

// connectDatabase opens the SQLite database named by dsn and migrates the
// schema. It returns errors rather than exiting so callers (and tests) can
// decide how to report a failure.
func connectDatabase(dsn string) (*gorm.DB, error) {
	// This module only links a SQLite driver and its schema uses portable
	// column types; a PostgreSQL DSN would otherwise be opened as a file whose
	// name is the DSN, which fails far from the cause.
	if isPostgresDSN(dsn) {
		return nil, fmt.Errorf("DATABASE_URL=%q selects a PostgreSQL server, but this server (backend/byteport) is SQLite-only; "+
			"leave DATABASE_URL unset to use the documented default %s, or use the API server in backend/", dsn, documentedSQLiteDefault)
	}

	database, err := gorm.Open(sqlite.Open(sqliteDSNFor(dsn)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database %s: %w", descriptionOf(dsn), err)
	}

	// AutoMigrate models in the correct order
	if err := database.AutoMigrate(&User{}, &Project{}, &Instance{}, &GitSecret{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate SQLite database %s: %w", descriptionOf(dsn), err)
	}

	return database, nil
}

// ConnectDatabase opens the SQLite database named by DATABASE_URL (falling back
// to the documented default) and migrates the schema.
func ConnectDatabase() {
	dsn, documentedDefault := resolveDSN()

	database, err := connectDatabase(dsn)
	if err != nil {
		log.Fatalf("%v", err)
	}
	DB = database

	if documentedDefault {
		if dsn == legacySQLitePath {
			fmt.Printf("DATABASE_URL not set; keeping the existing SQLite database %s (set DATABASE_URL=%s to switch to the documented default)\n",
				descriptionOf(dsn), documentedSQLiteDefault)
		} else {
			fmt.Printf("DATABASE_URL not set, using the documented SQLite default %s\n", descriptionOf(dsn))
		}
	}
	fmt.Printf("Successfully connected to SQLite database (%s)\n", descriptionOf(dsn))
}
