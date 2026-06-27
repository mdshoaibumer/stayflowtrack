package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
	"github.com/stayflow/stayflow-track/internal/shared/response"
)

const (
	csrfTokenLength = 32
	csrfCookieName  = "_csrf"
	csrfHeaderName  = "X-CSRF-Token"
)

// CSRFProtection provides Double Submit Cookie pattern CSRF protection.
// It sets a CSRF token in a cookie and validates that state-changing requests
// include the same token in the X-CSRF-Token header.
func CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for safe methods (GET, HEAD, OPTIONS)
		if isSafeMethod(r.Method) {
			ensureCSRFCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF check in non-production environments for testing
		if os.Getenv("APP_ENV") != "production" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF for Bearer token authenticated API requests.
		// APIs authenticated via Authorization header are not vulnerable to CSRF
		// because the browser cannot automatically attach the header cross-origin.
		if hasAuthorizationHeader(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF for webhook endpoints (they use signature verification)
		if isWebhookPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Validate CSRF token
		cookieToken := getCSRFCookie(r)
		headerToken := r.Header.Get(csrfHeaderName)

		if cookieToken == "" || headerToken == "" {
			response.Err(w, apperrors.New(http.StatusForbidden, "CSRF_MISSING", "CSRF token missing"))
			return
		}

		if !validateCSRFToken(cookieToken, headerToken) {
			response.Err(w, apperrors.New(http.StatusForbidden, "CSRF_INVALID", "CSRF token invalid"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func hasAuthorizationHeader(r *http.Request) bool {
	return r.Header.Get("Authorization") != ""
}

func isWebhookPath(path string) bool {
	return strings.Contains(path, "/webhooks/")
}

func ensureCSRFCookie(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie(csrfCookieName); err != nil {
		token := generateCSRFToken()
		http.SetCookie(w, &http.Cookie{
			Name:     csrfCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: false, // Must be readable by JavaScript
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400, // 24 hours
		})
	}
}

func getCSRFCookie(r *http.Request) string {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func generateCSRFToken() string {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		// Fallback should never happen, but if it does, fail closed
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func validateCSRFToken(cookieToken, headerToken string) bool {
	if len(cookieToken) == 0 || len(headerToken) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) == 1
}
