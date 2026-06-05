package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents application-level error codes.
type ErrorCode string

const (
	CodeNotFound       ErrorCode = "NOT_FOUND"
	CodeConflict       ErrorCode = "CONFLICT"
	CodeValidation     ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	CodeForbidden      ErrorCode = "FORBIDDEN"
	CodeInternal       ErrorCode = "INTERNAL_ERROR"
	CodeBadRequest     ErrorCode = "BAD_REQUEST"
	CodeTenantRequired ErrorCode = "TENANT_REQUIRED"
)

// AppError represents a structured application error.
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    any       `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Err        error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NotFound(resource, id string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s with id '%s' not found", resource, id),
		HTTPStatus: http.StatusNotFound,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

func Validation(message string, details any) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

func Internal(err error) *AppError {
	return &AppError{
		Code:       CodeInternal,
		Message:    "an internal error occurred",
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

func BadRequest(message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// New creates an AppError with a custom HTTP status, code, and message.
func New(httpStatus int, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}
