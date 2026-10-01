package errors_test

import (
	"errors"
	"net/http"
	"testing"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

func TestApplicationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *domainErrors.ApplicationError
		expected string
	}{
		{
			name:     "code, message, and cause",
			appErr:   domainErrors.NewApplicationErrorWithMessage("DB-7000", http.StatusServiceUnavailable, "database down", errors.New("connection refused")),
			expected: "DB-7000: database down: connection refused",
		},
		{
			name:     "code and cause only",
			appErr:   domainErrors.NewApplicationError("DB-7000", http.StatusServiceUnavailable, errors.New("connection refused")),
			expected: "DB-7000: connection refused",
		},
		{
			name:     "code and message only",
			appErr:   domainErrors.NewApplicationErrorWithMessage("AUTH-100", http.StatusUnauthorized, "unauthorized"),
			expected: "AUTH-100: unauthorized",
		},
		{
			name:     "code only",
			appErr:   domainErrors.NewApplicationError("DB-7000", http.StatusServiceUnavailable),
			expected: "DB-7000",
		},
		{
			name: "message and cause without code",
			appErr: &domainErrors.ApplicationError{
				Message: "database down",
				Cause:   errors.New("connection refused"),
			},
			expected: "database down: connection refused",
		},
		{
			name: "message only",
			appErr: &domainErrors.ApplicationError{
				Message: "something went wrong",
			},
			expected: "something went wrong",
		},
		{
			name: "cause only",
			appErr: &domainErrors.ApplicationError{
				Cause: errors.New("root cause"),
			},
			expected: "root cause",
		},
		{
			name:     "empty error",
			appErr:   &domainErrors.ApplicationError{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appErr.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestApplicationError_Unwrap(t *testing.T) {
	cause := errors.New("underlying cause")
	appErr := domainErrors.NewApplicationError("SRV-5000", http.StatusInternalServerError, cause)

	if !errors.Is(appErr, cause) {
		t.Errorf("expected errors.Is to find cause, got false")
	}
	if appErr.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", appErr.Unwrap(), cause)
	}
}

func TestAsApplicationError(t *testing.T) {
	appErr := domainErrors.NewApplicationError("DB-7000", http.StatusServiceUnavailable)
	extracted, ok := domainErrors.AsApplicationError(appErr)
	if !ok || extracted == nil {
		t.Fatalf("expected AsApplicationError to return true and non-nil")
	}
	if extracted.Code != "DB-7000" {
		t.Errorf("extracted Code = %q, want DB-7000", extracted.Code)
	}

	var regularErr = errors.New("normal error")
	_, ok = domainErrors.AsApplicationError(regularErr)
	if ok {
		t.Errorf("expected AsApplicationError to return false for normal error")
	}
}
