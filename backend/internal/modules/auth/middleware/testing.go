package middleware

import (
	"context"

	"github.com/stayflow/stayflow-track/internal/modules/auth/domain"
)

// SetClaimsForTest is a test helper that sets claims in context.
// Only available in _test files via this export.
func SetClaimsForTest(ctx context.Context, claims *domain.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}
