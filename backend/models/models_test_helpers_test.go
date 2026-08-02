package models

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func assertFindOrCreatePanics(t *testing.T, info *WorkOSUserInfo) {
	t.Helper()
	assert.Panics(t, func() { _, _ = FindOrCreateUserFromWorkOS(info) })
}

func assertGetUserPanics(t *testing.T, workosID string) {
	t.Helper()
	assert.Panics(t, func() { _, _ = GetUserByWorkOSID(workosID) })
}

func assertCreateUserPanics(t *testing.T, info *WorkOSUserInfo) {
	t.Helper()
	assert.Panics(t, func() { _, _ = CreateUserFromWorkOS(info) })
}

func assertBeforeSaveDoesNotPanic(t *testing.T, project *Project) {
	t.Helper()
	assert.NotPanics(t, func() { _ = project.BeforeSave(nil) })
}

func closeTestDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	if err == nil {
		require.NoError(t, sqlDB.Close())
	}
}

func setTestEnvironment(t *testing.T, key, value string) func() {
	t.Helper()
	require.NoError(t, os.Setenv(key, value))
	return func() { require.NoError(t, os.Unsetenv(key)) }
}
