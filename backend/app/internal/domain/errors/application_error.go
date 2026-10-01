package errors

import (
	stderrors "errors"
	"strings"
)

// ApplicationError represents an application error with an HTTP status.
type ApplicationError struct {
	Code       string
	Message    string
	StatusCode int
	Cause      error
}

func (e *ApplicationError) Error() string {
	parts := make([]string, 0, 3)
	if e.Code != "" {
		parts = append(parts, e.Code)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.Cause != nil && e.Cause.Error() != "" {
		parts = append(parts, e.Cause.Error())
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ": ")
}

// Unwrap returns the underlying error, when present.
func (e *ApplicationError) Unwrap() error {
	return e.Cause
}

// NewApplicationError creates an ApplicationError and records its source.
func NewApplicationError(code string, statusCode int, cause ...error) *ApplicationError {
	appErr := &ApplicationError{
		Code:       code,
		StatusCode: statusCode,
	}

	if len(cause) > 0 && cause[0] != nil {
		appErr.Cause = cause[0]
	}

	return appErr
}

// NewApplicationErrorWithMessage creates an ApplicationError with a public message.
func NewApplicationErrorWithMessage(
	code string,
	statusCode int,
	message string,
	cause ...error,
) *ApplicationError {
	appErr := NewApplicationError(code, statusCode, cause...)
	appErr.Message = message
	return appErr
}

// AsApplicationError unwraps err into an ApplicationError when possible.
func AsApplicationError(err error) (*ApplicationError, bool) {
	var appErr *ApplicationError
	ok := stderrors.As(err, &appErr)
	return appErr, ok
}
