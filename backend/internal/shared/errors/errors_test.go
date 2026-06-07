package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNotFound(t *testing.T) {
	err := NotFound("guest", "abc-123")
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected 404, got %d", err.HTTPStatus)
	}
	if err.Code != CodeNotFound {
		t.Errorf("expected NOT_FOUND, got %s", err.Code)
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestConflict(t *testing.T) {
	err := Conflict("reservation already exists for this period")
	if err.HTTPStatus != http.StatusConflict {
		t.Errorf("expected 409, got %d", err.HTTPStatus)
	}
	if err.Code != CodeConflict {
		t.Errorf("expected CONFLICT, got %s", err.Code)
	}
}

func TestValidation(t *testing.T) {
	details := map[string]string{"name": "required"}
	err := Validation("validation failed", details)
	if err.HTTPStatus != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", err.HTTPStatus)
	}
	if err.Code != CodeValidation {
		t.Errorf("expected VALIDATION_ERROR, got %s", err.Code)
	}
	if err.Details == nil {
		t.Error("expected details to be set")
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("invalid token")
	if err.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", err.HTTPStatus)
	}
	if err.Code != CodeUnauthorized {
		t.Errorf("expected UNAUTHORIZED, got %s", err.Code)
	}
}

func TestForbidden(t *testing.T) {
	err := Forbidden("insufficient permissions")
	if err.HTTPStatus != http.StatusForbidden {
		t.Errorf("expected 403, got %d", err.HTTPStatus)
	}
	if err.Code != CodeForbidden {
		t.Errorf("expected FORBIDDEN, got %s", err.Code)
	}
}

func TestInternal(t *testing.T) {
	cause := errors.New("database connection lost")
	err := Internal(cause)
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", err.HTTPStatus)
	}
	if err.Code != CodeInternal {
		t.Errorf("expected INTERNAL_ERROR, got %s", err.Code)
	}
	if err.Err != cause {
		t.Error("wrapped error should be preserved")
	}
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("invalid date format")
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", err.HTTPStatus)
	}
	if err.Code != CodeBadRequest {
		t.Errorf("expected BAD_REQUEST, got %s", err.Code)
	}
}

func TestNew_CustomError(t *testing.T) {
	err := New(http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
	if err.HTTPStatus != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", err.HTTPStatus)
	}
	if err.Code != "RATE_LIMITED" {
		t.Errorf("expected RATE_LIMITED, got %s", err.Code)
	}
}

func TestAppError_Error_WithWrapped(t *testing.T) {
	cause := errors.New("pg: connection refused")
	err := &AppError{
		Code:       CodeInternal,
		Message:    "database error",
		HTTPStatus: 500,
		Err:        cause,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error string should not be empty")
	}
	// Should contain both code and wrapped error
	if !errors.Is(err, cause) {
		t.Error("Unwrap should return the wrapped error")
	}
}

func TestAppError_Error_WithoutWrapped(t *testing.T) {
	err := BadRequest("invalid input")
	msg := err.Error()
	if msg == "" {
		t.Error("error string should not be empty")
	}
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	err := BadRequest("test")
	if err.Unwrap() != nil {
		t.Error("Unwrap should return nil when no wrapped error")
	}
}

func TestErrorsAs_AppError(t *testing.T) {
	err := NotFound("property", "xyz")
	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Error("should be able to extract AppError via errors.As")
	}
	if appErr.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected 404, got %d", appErr.HTTPStatus)
	}
}
