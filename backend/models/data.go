package models

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// documentedSQLiteDefault is the default DSN published in INSTALL.md,
// DEPLOYMENT.md and docker-compose.yml. Keep it in sync with those files.
const documentedSQLiteDefault = "file:./byteport.db"

var DB *gorm.DB

// driverAndDialectorFor picks the GORM driver implied by the DSN and returns
// both the driver name (for logging) and the dialector.
//
// A postgres DSN is either a URL ("postgres://…", "postgresql://…") or the
// key=value form used by libpq, which always carries "host=". A "sqlite:" /
// "sqlite://" prefix (docker-compose.yml uses "sqlite:///data/byteport.db") is
// stripped, because SQLite treats the remainder as a plain file path. Anything
// else (notably the documented "file:./byteport.db" default) is a SQLite path.
//
// This replaces a hardcoded postgres.Open, which contradicted the documented
// SQLite default and made the documented configuration impossible to run.
func driverAndDialectorFor(dsn string) (string, gorm.Dialector) {
	switch {
	case strings.HasPrefix(dsn, "postgres://"),
		strings.HasPrefix(dsn, "postgresql://"),
		strings.Contains(dsn, "host="):
		return "PostgreSQL", postgres.Open(dsn)
	case strings.HasPrefix(dsn, "sqlite://"):
		return "SQLite", sqlite.Open(normalisedSQLitePath(dsn))
	case strings.HasPrefix(dsn, "sqlite:"):
		return "SQLite", sqlite.Open(normalisedSQLitePath(dsn))
	default:
		return "SQLite", sqlite.Open(dsn)
	}
}

// normalisedSQLitePath turns a "sqlite://"-prefixed DSN into the file path the
// SQLite driver expects: "sqlite:///data/byteport.db" becomes
// "/data/byteport.db", and "sqlite://byteport.db" becomes "byteport.db".
func normalisedSQLitePath(dsn string) string {
	path := strings.TrimPrefix(dsn, "sqlite://")
	if path == dsn {
		path = strings.TrimPrefix(dsn, "sqlite:")
	}
	// "sqlite://" leaves a leading "//" for relative paths.
	path = strings.TrimPrefix(path, "//")
	if path == "" {
		return dsn
	}
	return path
}

// databaseDSN resolves the DSN to connect with, defaulting to the path the
// documentation publishes when DATABASE_URL is unset.
func databaseDSN() (dsn string, documentedDefault bool) {
	if dsn = strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn, false
	}
	return documentedSQLiteDefault, true
}

// connectDatabase opens the database described by dsn and migrates the schema.
//
// It returns errors rather than exiting so callers (and tests) can decide how
// to report a failure; ConnectDatabase is the process-level wrapper that turns
// a failure into a fatal error.
func connectDatabase(dsn string) (*gorm.DB, error) {
	driver, dialector := driverAndDialectorFor(dsn)

	// Configure GORM logger; release and test runs log warnings only.
	logLevel := logger.Info
	switch os.Getenv("GIN_MODE") {
	case "release", "test":
		logLevel = logger.Warn
	}

	database, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Disable foreign key constraints for SQLite compatibility during migration
		// Re-enable in production
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database (%s): %w", driver, dsn, err)
	}

	// Configure connection pool
	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping %s database (%s): %w", driver, dsn, err)
	}

	fmt.Printf("Successfully connected to %s database (%s)\n", driver, dsn)

	// AutoMigrate models in the correct order
	// Note: For production, use migration files instead
	err = database.AutoMigrate(
		&User{},
		&Repository{},
		&Project{},
		&Instance{},
		&AWSResource{},
		&Deployment{},
		&Host{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate %s database: %w", driver, err)
	}

	fmt.Println("Database migration completed successfully")
	return database, nil
}

func ConnectDatabase() {
	dsn, documentedDefault := databaseDSN()
	if documentedDefault {
		log.Printf("DATABASE_URL not set, using the documented SQLite default %s", dsn)
	}

	database, err := connectDatabase(dsn)
	if err != nil {
		log.Fatalf("%v", err)
	}

	DB = database
}
