package models

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Application-generated primary keys.
//
// The UUID primary keys used to be declared
// `type:uuid;primaryKey;default:gen_random_uuid()`. That is PostgreSQL-only
// DDL: SQLite rejects `DEFAULT gen_random_uuid()` with
// `near "(": syntax error`, so the documented SQLite default (INSTALL.md,
// DEPLOYMENT.md and docker-compose.yml all say `file:./byteport.db`) could
// never migrate and AutoMigrate aborted on the first table.
//
// Generating the value in Go instead keeps a single portable schema:
//
//   - the column type stays `type:uuid` (native on PostgreSQL, and accepted by
//     SQLite, which allows free-form column type names), and
//   - the value is filled by a BeforeCreate hook that runs on every driver.
//
// Project already generated its UUID this way (see Project.BeforeSave); these
// hooks bring the remaining tables in line.
//
// One caveat worth stating plainly: with the database-side default gone, rows
// inserted by raw SQL rather than through GORM no longer receive a UUID.

// newUUID returns a random RFC 4122 UUID string.
func newUUID() string {
	return uuid.New().String()
}

// BeforeCreate fills User.UUID when the caller did not set one.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(u.UUID) == "" {
		u.UUID = newUUID()
	}
	return nil
}

// BeforeCreate fills WorkOSUser.UUID when the caller did not set one.
func (u *WorkOSUser) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(u.UUID) == "" {
		u.UUID = newUUID()
	}
	return nil
}

// BeforeCreate fills Instance.UUID when the caller did not set one.
func (i *Instance) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(i.UUID) == "" {
		i.UUID = newUUID()
	}
	return nil
}

// BeforeCreate fills Deployment.UUID when the caller did not set one.
func (d *Deployment) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(d.UUID) == "" {
		d.UUID = newUUID()
	}
	return nil
}

// BeforeCreate fills Host.UUID when the caller did not set one.
func (h *Host) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(h.UUID) == "" {
		h.UUID = newUUID()
	}
	return nil
}
