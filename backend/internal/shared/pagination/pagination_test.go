package pagination_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stayflow/stayflow-track/internal/shared/pagination"
)

func TestParse_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	params := pagination.Parse(req)

	if params.Page != 1 {
		t.Errorf("expected page 1, got %d", params.Page)
	}
	if params.PerPage != 20 {
		t.Errorf("expected per_page 20, got %d", params.PerPage)
	}
	if params.Offset != 0 {
		t.Errorf("expected offset 0, got %d", params.Offset)
	}
}

func TestParse_CustomValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?page=3&per_page=50", nil)
	params := pagination.Parse(req)

	if params.Page != 3 {
		t.Errorf("expected page 3, got %d", params.Page)
	}
	if params.PerPage != 50 {
		t.Errorf("expected per_page 50, got %d", params.PerPage)
	}
	if params.Offset != 100 {
		t.Errorf("expected offset 100, got %d", params.Offset)
	}
}

func TestParse_MaxPerPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?per_page=500", nil)
	params := pagination.Parse(req)

	if params.PerPage != 100 {
		t.Errorf("expected per_page capped at 100, got %d", params.PerPage)
	}
}

func TestParse_InvalidValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?page=-1&per_page=0", nil)
	params := pagination.Parse(req)

	if params.Page != 1 {
		t.Errorf("expected page 1 for invalid input, got %d", params.Page)
	}
	if params.PerPage != 20 {
		t.Errorf("expected per_page 20 for invalid input, got %d", params.PerPage)
	}
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		total    int64
		perPage  int
		expected int64
	}{
		{0, 20, 0},
		{1, 20, 1},
		{20, 20, 1},
		{21, 20, 2},
		{100, 20, 5},
		{101, 20, 6},
	}

	for _, tt := range tests {
		result := pagination.TotalPages(tt.total, tt.perPage)
		if result != tt.expected {
			t.Errorf("TotalPages(%d, %d) = %d, expected %d", tt.total, tt.perPage, result, tt.expected)
		}
	}
}
