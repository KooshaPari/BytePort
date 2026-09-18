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

var DB *gorm.DB

// dialectorFor picks the GORM driver implied by the DSN.
//
// A postgres DSN is either a URL ("postgres://…", "postgresql://…") or the
// key=value form used by libpq, which always carries "host=". Anything else
// (notably the documented "file:./byteport.db" default) is a SQLite path.
//
// This replaces a hardcoded postgres.Open, which contradicted the documented
// SQLite default and made the documented configuration impossible to run.
func dialectorFor(dsn string) gorm.Dialector {
	switch {
	case strings.HasPrefix(dsn, "postgres://"),
		strings.HasPrefix(dsn, "postgresql://"),
		strings.Contains(dsn, "host="):
		return postgres.Open(dsn)
	default:
		return sqlite.Open(dsn)
	}
}

func ConnectDatabase() {
	// Get database URL from environment variable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Documented default: INSTALL.md, DEPLOYMENT.md and docker-compose.yml
		// all state `file:./byteport.db`. This previously fell back to a
		// hardcoded external PostgreSQL instance, so the documented default
		// could never work.
		dsn = "file:./byteport.db"
		log.Println("DATABASE_URL not set, using SQLite at ./byteport.db")
	}

	// Configure GORM logger
	logLevel := logger.Info
	if os.Getenv("GIN_MODE") == "release" {
		logLevel = logger.Warn
	}

	// Open database connection with the driver implied by the DSN.
	database, err := gorm.Open(dialectorFor(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Disable foreign key constraints for SQLite compatibility during migration
		// Re-enable in production
		DisableForeignKeyConstraintWhenMigrating: false,
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL database")

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
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	fmt.Println("Database migration completed successfully")

	DB = database
}
