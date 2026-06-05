package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stayflow/stayflow-track/internal/modules/auth/domain"
	"github.com/stayflow/stayflow-track/internal/modules/auth/middleware"
)

func TestRequireRole_Allowed(t *testing.T) {
	claims := &domain.Claims{
		RoleName: "super_admin",
	}

	handler := middleware.RequireRole("super_admin", "property_admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := middleware.SetClaimsForTest(req.Context(), claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	claims := &domain.Claims{
		RoleName: "housekeeping",
	}

	handler := middleware.RequireRole("super_admin", "property_admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := middleware.SetClaimsForTest(req.Context(), claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestRequireRole_NoClaims(t *testing.T) {
	handler := middleware.RequireRole("super_admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}
