package deployment

import (
	"errors"
	"strings"
	"testing"
)

// =============================================================================
// Shared test assertion helpers for the deployment package.
//
// The four files in this package (create_deployment_test.go, get_list_test.go,
// terminate_update_test.go, and errors_test.go) historically each repeated the
// same ~5-line `errors.As → assert status → assert code → assert message`
// boilerplate. This file consolidates those patterns into single-call helpers
// so each test's body stays linear (which keeps SonarCloud go:S3776 cognitive
// complexity below 15).
//
// All helpers are t.Helper()s so failures report the caller's line.
// =============================================================================

// extractAppError returns the *ApplicationError wrapped inside err, or nil if
// err is not (or does not wrap) an ApplicationError. Callers should normally
// pair this with requireAppError when the error is the test's primary signal.
func extractAppError(err error) *ApplicationError {
	var appErr *ApplicationError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// requireAppError fails the test if err is not (or does not wrap) an
// ApplicationError. Returns the unwrapped ApplicationError so callers can keep
// asserting against fields without re-running errors.As.
func requireAppError(t *testing.T, err error) *ApplicationError {
	t.Helper()
	appErr := extractAppError(err)
	if appErr == nil {
		t.Fatalf("Expected ApplicationError, got: %T (%v)", err, err)
	}
	return appErr
}

// assertErrStatus fails if the wrapped ApplicationError's StatusCode does not
// match want. Skips when err does not wrap an ApplicationError so this helper
// can be used in tests that primarily assert success but want to double-check
// a status code path.
func assertErrStatus(t *testing.T, err error, want int) {
	t.Helper()
	appErr := extractAppError(err)
	if appErr == nil {
		return
	}
	if appErr.StatusCode != want {
		t.Errorf("StatusCode = %d, want %d", appErr.StatusCode, want)
	}
}

// assertErrCode fails if the wrapped ApplicationError's Code does not match
// want. Skips when err does not wrap an ApplicationError.
func assertErrCode(t *testing.T, err error, want string) {
	t.Helper()
	appErr := extractAppError(err)
	if appErr == nil {
		return
	}
	if appErr.Code != want {
		t.Errorf("Code = %q, want %q", appErr.Code, want)
	}
}

// assertErrMessage fails if the wrapped ApplicationError's Message does not
// match want. If contains is true, the assertion passes when Message contains
// want as a substring. Skips when err does not wrap an ApplicationError.
func assertErrMessage(t *testing.T, err error, want string, contains bool) {
	t.Helper()
	appErr := extractAppError(err)
	if appErr == nil {
		return
	}
	if contains {
		if !strings.Contains(appErr.Message, want) {
			t.Errorf("Message = %q, want it to contain %q", appErr.Message, want)
		}
		return
	}
	if appErr.Message != want {
		t.Errorf("Message = %q, want %q", appErr.Message, want)
	}
}

// assertErrWraps fails if the wrapped ApplicationError's wrapped cause does
// not match want. Skips when want is nil or when err does not wrap an
// ApplicationError.
func assertErrWraps(t *testing.T, err error, wantWrapped error) {
	t.Helper()
	if wantWrapped == nil {
		return
	}
	appErr := extractAppError(err)
	if appErr == nil {
		return
	}
	if appErr.Err != wantWrapped {
		t.Errorf("wrapped error = %v, want %v", appErr.Err, wantWrapped)
	}
}

// assertErrCodeUnique fails if code is already present in seen. Updates seen
// so the next call sees this code as taken. Use to assert that a family of
// constructor helpers produces distinct codes.
func assertErrCodeUnique(t *testing.T, seen map[string]bool, code string) {
	t.Helper()
	if seen[code] {
		t.Errorf("duplicate error code %q across constructors", code)
	}
	seen[code] = true
}

// assertAppError is the most common helper in this package: it unwraps err to
// an *ApplicationError, fails the test if it isn't one, and asserts the
// expected (status, code) tuple in a single call.
//
// Use this for the canonical "use case returned a validation/not-found/etc.
// error" assertion. For message/wrap assertions, layer assertErrMessage and
// assertErrWraps on top.
func assertAppError(t *testing.T, err error, wantStatus int, wantCode string) *ApplicationError {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected ApplicationError, got nil")
	}
	appErr := requireAppError(t, err)
	assertErrStatus(t, err, wantStatus)
	assertErrCode(t, err, wantCode)
	return appErr
}

// assertAppErrStatus is the status-only counterpart to assertAppError. It is
// useful for tests that only need to assert on HTTP status code (e.g. the
// terminate/update path where each branch maps to a unique status) and
// shouldn't constrain the error code.
func assertAppErrStatus(t *testing.T, err error, wantStatus int) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected ApplicationError with status %d, got nil", wantStatus)
	}
	appErr := requireAppError(t, err)
	if appErr.StatusCode != wantStatus {
		t.Errorf("StatusCode = %d, want %d", appErr.StatusCode, wantStatus)
	}
}
