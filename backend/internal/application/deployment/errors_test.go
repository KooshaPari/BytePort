package deployment

import (
	"errors"
	"strings"
	"testing"
)

// =============================================================================
// ApplicationError.Error / Unwrap
// =============================================================================

func TestApplicationError_ErrorString(t *testing.T) {
	t.Run("with wrapped error contains all parts", func(t *testing.T) {
		innerErr := errors.New("database error")
		appErr := &ApplicationError{
			Code:       "TEST_CODE",
			Message:    "test message",
			StatusCode: 500,
			Err:        innerErr,
		}

		got := appErr.Error()
		for _, want := range []string{"TEST_CODE", "test message", "database error"} {
			if !strings.Contains(got, want) {
				t.Errorf("Error() = %q, want it to contain %q", got, want)
			}
		}
	})

	t.Run("without wrapped error is 'CODE: message'", func(t *testing.T) {
		appErr := &ApplicationError{
			Code:       "TEST_CODE",
			Message:    "test message",
			StatusCode: 400,
		}

		if got, want := appErr.Error(), "TEST_CODE: test message"; got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestApplicationError_Unwrap(t *testing.T) {
	t.Run("returns wrapped error", func(t *testing.T) {
		innerErr := errors.New("inner error")
		appErr := &ApplicationError{Err: innerErr}

		if got := appErr.Unwrap(); got != innerErr {
			t.Errorf("Unwrap() = %v, want %v", got, innerErr)
		}
	})

	t.Run("returns nil when no wrapped error", func(t *testing.T) {
		appErr := &ApplicationError{}

		if got := appErr.Unwrap(); got != nil {
			t.Errorf("Unwrap() = %v, want nil", got)
		}
	})
}

// =============================================================================
// NewXxxError constructors (table-driven)
// =============================================================================

func TestNewErrorConstructors(t *testing.T) {
	innerErr := errors.New("database connection failed")

	cases := []struct {
		name         string
		build        func() *ApplicationError
		wantCode     string
		wantStatus   int
		wantMsg      string
		wantContains bool // if true, only check message contains (not equality)
		wantWrapped  error
	}{
		{
			name:       "Validation",
			build:      func() *ApplicationError { return NewValidationError("invalid input") },
			wantCode:   "VALIDATION_ERROR",
			wantStatus: StatusBadRequest,
			wantMsg:    "invalid input",
		},
		{
			name:         "NotFound",
			build:        func() *ApplicationError { return NewNotFoundError("deployment") },
			wantCode:     "NOT_FOUND",
			wantStatus:   StatusNotFound,
			wantMsg:      "deployment",
			wantContains: true,
		},
		{
			name:       "Unauthorized",
			build:      func() *ApplicationError { return NewUnauthorizedError("authentication required") },
			wantCode:   "UNAUTHORIZED",
			wantStatus: StatusUnauthorized,
			wantMsg:    "authentication required",
		},
		{
			name:       "Forbidden",
			build:      func() *ApplicationError { return NewForbiddenError("access denied") },
			wantCode:   "FORBIDDEN",
			wantStatus: StatusForbidden,
			wantMsg:    "access denied",
		},
		{
			name:       "Conflict",
			build:      func() *ApplicationError { return NewConflictError("resource already exists") },
			wantCode:   "CONFLICT",
			wantStatus: StatusConflict,
			wantMsg:    "resource already exists",
		},
		{
			name:        "Internal",
			build:       func() *ApplicationError { return NewInternalError("internal server error", innerErr) },
			wantCode:    "INTERNAL_ERROR",
			wantStatus:  StatusInternalServerError,
			wantMsg:     "internal server error",
			wantWrapped: innerErr,
		},
	}

	seenCodes := make(map[string]bool)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.build()
			if err == nil {
				t.Fatal("constructor returned nil")
			}
			if err.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", err.Code, tc.wantCode)
			}
			if err.StatusCode != tc.wantStatus {
				t.Errorf("StatusCode = %d, want %d", err.StatusCode, tc.wantStatus)
			}
			if tc.wantContains {
				if !strings.Contains(err.Message, tc.wantMsg) {
					t.Errorf("Message = %q, want it to contain %q", err.Message, tc.wantMsg)
				}
			} else if err.Message != tc.wantMsg {
				t.Errorf("Message = %q, want %q", err.Message, tc.wantMsg)
			}
			if tc.wantWrapped != nil && err.Err != tc.wantWrapped {
				t.Errorf("wrapped error = %v, want %v", err.Err, tc.wantWrapped)
			}
			if seenCodes[err.Code] {
				t.Errorf("duplicate error code %q across constructors", err.Code)
			}
			seenCodes[err.Code] = true
		})
	}
}

// =============================================================================
// errors.Is / errors.As integration
// =============================================================================

func TestErrorUnwrapping(t *testing.T) {
	innerErr := errors.New("database error")
	appErr := NewInternalError("failed operation", innerErr)

	t.Run("errors.Is finds wrapped error", func(t *testing.T) {
		if !errors.Is(appErr, innerErr) {
			t.Error("expected errors.Is to find wrapped error")
		}
	})

	t.Run("errors.As extracts ApplicationError", func(t *testing.T) {
		var asAppErr *ApplicationError
		if !errors.As(appErr, &asAppErr) {
			t.Fatal("expected errors.As to work with ApplicationError")
		}
		if asAppErr.Code != "INTERNAL_ERROR" {
			t.Errorf("Code = %q, want %q", asAppErr.Code, "INTERNAL_ERROR")
		}
	})
}
