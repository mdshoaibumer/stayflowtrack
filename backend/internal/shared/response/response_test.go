package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

func TestJSON_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"name": "test"}
	JSON(rr, http.StatusOK, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Error != nil {
		t.Error("expected no error")
	}
}

func TestJSON_ContentType(t *testing.T) {
	rr := httptest.NewRecorder()
	JSON(rr, http.StatusCreated, nil)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestJSONWithMeta(t *testing.T) {
	rr := httptest.NewRecorder()
	meta := &Meta{Page: 1, PerPage: 20, Total: 100, TotalPages: 5}
	JSONWithMeta(rr, http.StatusOK, []string{"item1"}, meta)

	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Meta == nil {
		t.Fatal("expected meta to be present")
	}
	if resp.Meta.Total != 100 {
		t.Errorf("expected total 100, got %d", resp.Meta.Total)
	}
	if resp.Meta.TotalPages != 5 {
		t.Errorf("expected 5 pages, got %d", resp.Meta.TotalPages)
	}
}

func TestErr_AppError(t *testing.T) {
	rr := httptest.NewRecorder()
	appErr := apperrors.NotFound("guest", "abc-123")
	Err(rr, appErr)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error == nil {
		t.Fatal("expected error to be present")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestErr_GenericError(t *testing.T) {
	rr := httptest.NewRecorder()
	err := errors.New("something went wrong")
	Err(rr, err)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}

	var resp APIResponse
	if decErr := json.Unmarshal(rr.Body.Bytes(), &resp); decErr != nil {
		t.Fatalf("failed to decode: %v", decErr)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
}

func TestErr_BadRequest(t *testing.T) {
	rr := httptest.NewRecorder()
	appErr := apperrors.BadRequest("invalid date format")
	Err(rr, appErr)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestErr_Validation(t *testing.T) {
	rr := httptest.NewRecorder()
	details := []map[string]string{{"field": "email", "error": "invalid"}}
	appErr := apperrors.Validation("validation failed", details)
	Err(rr, appErr)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error.Details == nil {
		t.Error("expected validation details")
	}
}
