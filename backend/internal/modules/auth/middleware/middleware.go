package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/stayflow/stayflow-track/internal/modules/auth/domain"
	"github.com/stayflow/stayflow-track/internal/modules/auth/service"
	"github.com/stayflow/stayflow-track/internal/platform/database"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
	"github.com/stayflow/stayflow-track/internal/shared/response"
)

type contextKey string

const (
	claimsKey contextKey = "auth_claims"
	tenantKey contextKey = "tenant_id"
)

type AuthMiddleware struct {
	authService      *service.Service
	log              zerolog.Logger
	platformTenantID uuid.UUID // cached at startup, zero value if not configured
}

func New(authService *service.Service, log zerolog.Logger) *AuthMiddleware {
	m := &AuthMiddleware{authService: authService, log: log}
	// Cache platform tenant ID at startup — avoid reading ENV on every request
	if ptid := os.Getenv("PLATFORM_TENANT_ID"); ptid != "" {
		if parsed, err := uuid.Parse(ptid); err == nil {
			m.platformTenantID = parsed
		}
	}
	return m
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Err(w, apperrors.Unauthorized("missing authorization header"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Err(w, apperrors.Unauthorized("invalid authorization header format"))
			return
		}

		claims, err := m.authService.ValidateAccessToken(parts[1])
		if err != nil {
			response.Err(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TenantContext ensures the tenant_id from JWT is set in context.
func TenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r.Context())
		if claims == nil {
			response.Err(w, apperrors.Unauthorized("missing authentication context"))
			return
		}

		ctx := context.WithValue(r.Context(), tenantKey, claims.TenantID)
		ctx = database.WithTenantID(ctx, claims.TenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole creates middleware that restricts access to specific roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				response.Err(w, apperrors.Unauthorized("missing authentication context"))
				return
			}

			for _, role := range roles {
				if claims.RoleName == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.Err(w, apperrors.Forbidden("insufficient permissions"))
		})
	}
}

// GetClaims retrieves the authenticated user's claims from context.
func GetClaims(ctx context.Context) *domain.Claims {
	claims, _ := ctx.Value(claimsKey).(*domain.Claims)
	return claims
}

// GetTenantID retrieves the tenant ID from context.
func GetTenantID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(tenantKey).(uuid.UUID)
	return id
}

// RequirePlatformAdmin restricts access to the designated platform admin tenant.
// The platform tenant is identified by the PLATFORM_TENANT_ID environment variable (cached at startup).
// This prevents tenant-level super_admins from accessing platform-wide admin APIs.
func (m *AuthMiddleware) RequirePlatformAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r.Context())
		if claims == nil {
			response.Err(w, apperrors.Unauthorized("missing authentication context"))
			return
		}

		if m.platformTenantID == uuid.Nil {
			response.Err(w, apperrors.Forbidden("platform admin not configured"))
			return
		}

		if claims.TenantID != m.platformTenantID || claims.RoleName != "super_admin" {
			response.Err(w, apperrors.Forbidden("platform admin access required"))
			return
		}

		next.ServeHTTP(w, r.WithContext(database.WithPlatformAdmin(r.Context())))
	})
}
