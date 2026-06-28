package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCSRFProtection_SafeMethodsPassThrough(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	methods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/test", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", method, rr.Code)
			}
		})
	}
}

func TestCSRFProtection_SetsTokenCookieOnGET(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "_csrf" {
			found = true
			if c.Value == "" {
				t.Error("CSRF cookie should have a value")
			}
			if c.HttpOnly {
				t.Error("CSRF cookie should NOT be HttpOnly (JS needs to read it)")
			}
			if !c.Secure {
				t.Error("CSRF cookie should be Secure")
			}
			if c.SameSite != http.SameSiteStrictMode {
				t.Error("CSRF cookie should have SameSite=Strict")
			}
		}
	}
	if !found {
		t.Error("expected _csrf cookie to be set")
	}
}

func TestCSRFProtection_BlocksPOSTWithoutToken(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader("{}"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestCSRFProtection_BlocksMismatchedTokens(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader("{}"))
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: "token-aaa"})
	req.Header.Set("X-CSRF-Token", "token-bbb")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for mismatched tokens, got %d", rr.Code)
	}
}

func TestCSRFProtection_AllowsMatchingTokens(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := "valid-csrf-token-for-testing-12345"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader("{}"))
	req.AddCookie(&http.Cookie{Name: "_csrf", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for matching tokens, got %d", rr.Code)
	}
}

func TestCSRFProtection_SkipsWithAuthorizationHeader(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer some-jwt-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 when Authorization header present, got %d", rr.Code)
	}
}

func TestCSRFProtection_SkipsWebhookPaths(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	webhookPaths := []string{
		"/api/v1/webhooks/razorpay",
		"/api/v1/webhooks/notifications",
	}

	for _, path := range webhookPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("{}"))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("expected 200 for webhook path %s, got %d", path, rr.Code)
			}
		})
	}
}

func TestCSRFProtection_SkippedInDevMode(t *testing.T) {
	os.Setenv("APP_ENV", "development")
	defer os.Unsetenv("APP_ENV")

	handler := CSRFProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader("{}"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 in dev mode (no CSRF enforcement), got %d", rr.Code)
	}
}

func TestValidateCSRFToken_ConstantTimeComparison(t *testing.T) {
	tests := []struct {
		name     string
		cookie   string
		header   string
		expected bool
	}{
		{"matching tokens", "abc123", "abc123", true},
		{"different tokens", "abc123", "xyz456", false},
		{"empty cookie", "", "abc123", false},
		{"empty header", "abc123", "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateCSRFToken(tt.cookie, tt.header)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGenerateCSRFToken_NotEmpty(t *testing.T) {
	token := generateCSRFToken()
	if token == "" {
		t.Error("generated CSRF token should not be empty")
	}
	if len(token) < 32 {
		t.Errorf("token should be at least 32 chars, got %d", len(token))
	}
}

func TestGenerateCSRFToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := generateCSRFToken()
		if tokens[token] {
			t.Fatal("generated duplicate CSRF token")
		}
		tokens[token] = true
	}
}

func TestIsSafeMethod(t *testing.T) {
	safe := []string{"GET", "HEAD", "OPTIONS"}
	unsafe := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, m := range safe {
		if !isSafeMethod(m) {
			t.Errorf("%s should be safe", m)
		}
	}
	for _, m := range unsafe {
		if isSafeMethod(m) {
			t.Errorf("%s should NOT be safe", m)
		}
	}
}

func TestIsWebhookPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/v1/webhooks/razorpay", true},
		{"/api/v1/webhooks/notifications", true},
		{"/api/v1/auth/login", false},
		{"/api/v1/reservations", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if isWebhookPath(tt.path) != tt.expected {
				t.Errorf("expected %v for path %s", tt.expected, tt.path)
			}
		})
	}
}
