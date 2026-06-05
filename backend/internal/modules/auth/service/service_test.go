package service_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/stayflow/stayflow-track/internal/config"
	"github.com/stayflow/stayflow-track/internal/modules/auth/service"
)

func TestValidateAccessToken(t *testing.T) {
	jwtCfg := config.JWTConfig{
		AccessSecret:      "test-access-secret-32-characters",
		RefreshSecret:     "test-refresh-secret-32-characters",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 168 * time.Hour,
		Issuer:            "stayflow-track-test",
	}

	svc := service.New(nil, jwtCfg)

	t.Run("valid token", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":       userID.String(),
			"tenant_id": tenantID.String(),
			"role":      "super_admin",
			"email":     "test@example.com",
			"iss":       "stayflow-track-test",
			"iat":       time.Now().Unix(),
			"exp":       time.Now().Add(15 * time.Minute).Unix(),
		})

		tokenString, err := token.SignedString([]byte(jwtCfg.AccessSecret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		claims, err := svc.ValidateAccessToken(tokenString)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("expected user_id %s, got %s", userID, claims.UserID)
		}
		if claims.TenantID != tenantID {
			t.Errorf("expected tenant_id %s, got %s", tenantID, claims.TenantID)
		}
		if claims.RoleName != "super_admin" {
			t.Errorf("expected role super_admin, got %s", claims.RoleName)
		}
		if claims.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", claims.Email)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":       uuid.New().String(),
			"tenant_id": uuid.New().String(),
			"role":      "super_admin",
			"email":     "test@example.com",
			"exp":       time.Now().Add(-1 * time.Hour).Unix(),
		})

		tokenString, _ := token.SignedString([]byte(jwtCfg.AccessSecret))
		_, err := svc.ValidateAccessToken(tokenString)
		if err == nil {
			t.Fatal("expected error for expired token")
		}
	})

	t.Run("wrong signing key", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":       uuid.New().String(),
			"tenant_id": uuid.New().String(),
			"role":      "super_admin",
			"email":     "test@example.com",
			"exp":       time.Now().Add(15 * time.Minute).Unix(),
		})

		tokenString, _ := token.SignedString([]byte("wrong-secret"))
		_, err := svc.ValidateAccessToken(tokenString)
		if err == nil {
			t.Fatal("expected error for wrong signing key")
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := svc.ValidateAccessToken("not-a-valid-jwt")
		if err == nil {
			t.Fatal("expected error for invalid token format")
		}
	})
}
