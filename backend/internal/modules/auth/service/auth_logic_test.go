package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/stayflow/stayflow-track/internal/config"
)

func newTestJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		AccessSecret:      "test-access-secret-32chars-long!",
		RefreshSecret:     "test-refresh-secret-32chars-lng!",
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 168 * time.Hour,
		Issuer:            "stayflow-track-test",
	}
}

func TestValidateAccessToken_ValidToken(t *testing.T) {
	jwtCfg := newTestJWTConfig()
	svc := New(nil, jwtCfg)

	userID := uuid.New()
	tenantID := uuid.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       userID.String(),
		"tenant_id": tenantID.String(),
		"role":      "super_admin",
		"email":     "admin@hotel.com",
		"iss":       "stayflow-track-test",
		"exp":       time.Now().Add(15 * time.Minute).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(jwtCfg.AccessSecret))

	claims, err := svc.ValidateAccessToken(tokenString)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("user_id: expected %s, got %s", userID, claims.UserID)
	}
	if claims.TenantID != tenantID {
		t.Errorf("tenant_id: expected %s, got %s", tenantID, claims.TenantID)
	}
	if claims.RoleName != "super_admin" {
		t.Errorf("role: expected super_admin, got %s", claims.RoleName)
	}
	if claims.Email != "admin@hotel.com" {
		t.Errorf("email: expected admin@hotel.com, got %s", claims.Email)
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	jwtCfg := newTestJWTConfig()
	svc := New(nil, jwtCfg)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       uuid.New().String(),
		"tenant_id": uuid.New().String(),
		"role":      "receptionist",
		"email":     "test@test.com",
		"exp":       time.Now().Add(-1 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(jwtCfg.AccessSecret))

	_, err := svc.ValidateAccessToken(tokenString)
	if err == nil {
		t.Fatal("should reject expired token")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	jwtCfg := newTestJWTConfig()
	svc := New(nil, jwtCfg)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       uuid.New().String(),
		"tenant_id": uuid.New().String(),
		"role":      "super_admin",
		"email":     "test@test.com",
		"exp":       time.Now().Add(15 * time.Minute).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("wrong-secret-key-that-is-invalid"))

	_, err := svc.ValidateAccessToken(tokenString)
	if err == nil {
		t.Fatal("should reject token signed with wrong key")
	}
}

func TestValidateAccessToken_NoneAlgorithm(t *testing.T) {
	jwtCfg := newTestJWTConfig()
	svc := New(nil, jwtCfg)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub":       uuid.New().String(),
		"tenant_id": uuid.New().String(),
		"role":      "super_admin",
		"email":     "attacker@evil.com",
		"exp":       time.Now().Add(15 * time.Minute).Unix(),
	})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := svc.ValidateAccessToken(tokenString)
	if err == nil {
		t.Fatal("CRITICAL: accepted token with 'none' algorithm")
	}
}

func TestValidateAccessToken_InvalidFormat(t *testing.T) {
	jwtCfg := newTestJWTConfig()
	svc := New(nil, jwtCfg)

	_, err := svc.ValidateAccessToken("not.a.jwt")
	if err == nil {
		t.Fatal("should reject invalid token format")
	}

	_, err = svc.ValidateAccessToken("")
	if err == nil {
		t.Fatal("should reject empty token")
	}
}

func TestValidateAccessToken_MissingRequiredClaims(t *testing.T) {
	jwtCfg := newTestJWTConfig()
	svc := New(nil, jwtCfg)

	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{"missing sub", jwt.MapClaims{
			"tenant_id": uuid.New().String(), "role": "admin", "email": "a@b.com",
			"exp": time.Now().Add(15 * time.Minute).Unix(),
		}},
		{"missing tenant_id", jwt.MapClaims{
			"sub": uuid.New().String(), "role": "admin", "email": "a@b.com",
			"exp": time.Now().Add(15 * time.Minute).Unix(),
		}},
		{"missing role", jwt.MapClaims{
			"sub": uuid.New().String(), "tenant_id": uuid.New().String(), "email": "a@b.com",
			"exp": time.Now().Add(15 * time.Minute).Unix(),
		}},
		{"missing email", jwt.MapClaims{
			"sub": uuid.New().String(), "tenant_id": uuid.New().String(), "role": "admin",
			"exp": time.Now().Add(15 * time.Minute).Unix(),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tt.claims)
			tokenString, _ := token.SignedString([]byte(jwtCfg.AccessSecret))

			_, err := svc.ValidateAccessToken(tokenString)
			if err == nil {
				t.Errorf("should reject token with %s", tt.name)
			}
		})
	}
}

func TestValidateAccessToken_AllRoles(t *testing.T) {
	jwtCfg := newTestJWTConfig()
	svc := New(nil, jwtCfg)

	roles := []string{"super_admin", "property_admin", "receptionist", "housekeeping"}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub":       uuid.New().String(),
				"tenant_id": uuid.New().String(),
				"role":      role,
				"email":     role + "@hotel.com",
				"iss":       "stayflow-track-test",
				"exp":       time.Now().Add(15 * time.Minute).Unix(),
			})
			tokenString, _ := token.SignedString([]byte(jwtCfg.AccessSecret))

			claims, err := svc.ValidateAccessToken(tokenString)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if claims.RoleName != role {
				t.Errorf("expected role %s, got %s", role, claims.RoleName)
			}
		})
	}
}

func TestPasswordComplexity(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"strong - Admin@123!", "Admin@123!", true},
		{"strong - P@ssw0rd", "P@ssw0rd", true},
		{"strong - MyH0tel!2026", "MyH0tel!2026", true},
		{"strong - A1b!cdef", "A1b!cdef", true},
		{"no uppercase", "admin@123!", false},
		{"no lowercase", "ADMIN@123!", false},
		{"no digit", "Admin@Pass!", false},
		{"no special char", "AdminPass1", false},
		{"only lowercase", "abcdefgh", false},
		{"only digits", "12345678", false},
		{"only special", "!@#$%^&*", false},
		{"unicode rupee as special", "Admin1₹pass", true},
		{"space is not special char", "Admin1 pass", false}, // space is NOT punct/symbol in Go
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePasswordComplexity(tt.password)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("expected error for invalid password")
			}
		})
	}
}

func TestSplitFullName(t *testing.T) {
	tests := []struct {
		input     string
		firstName string
		lastName  string
	}{
		{"John Doe", "John", "Doe"},
		{"Alice", "Alice", "."},
		{"", "User", "."},
		{"  ", "User", "."},
		{"Dr. Ram Kumar Singh", "Dr.", "Ram Kumar Singh"},
		{"Mary Jane Watson", "Mary", "Jane Watson"},
		{"   Leading Spaces", "Leading", "Spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			first, last := splitFullName(tt.input)
			if first != tt.firstName {
				t.Errorf("firstName: expected %q, got %q", tt.firstName, first)
			}
			if last != tt.lastName {
				t.Errorf("lastName: expected %q, got %q", tt.lastName, last)
			}
		})
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Grand Hotel", "grand-hotel"},
		{"My Property", "my-property"},
		{"Hotel @#$ Name", "hotel--name"},
		{"simple", "simple"},
		{"With-Dash", "with-dash"},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := generateSlug(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRoleLevel(t *testing.T) {
	tests := []struct {
		role  string
		level int
	}{
		{"super_admin", 100},
		{"property_admin", 50},
		{"receptionist", 20},
		{"housekeeping", 10},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			if roleLevel(tt.role) != tt.level {
				t.Errorf("expected level %d for role %s, got %d", tt.level, tt.role, roleLevel(tt.role))
			}
		})
	}
}

func TestCanCreateRole(t *testing.T) {
	tests := []struct {
		caller string
		target string
		can    bool
	}{
		{"super_admin", "property_admin", true},
		{"super_admin", "receptionist", true},
		{"super_admin", "housekeeping", true},
		{"property_admin", "receptionist", true},
		{"property_admin", "housekeeping", true},
		{"property_admin", "property_admin", false}, // equal level
		{"property_admin", "super_admin", false},    // lower creating higher
		{"receptionist", "housekeeping", true},
		{"receptionist", "receptionist", false},
		{"housekeeping", "receptionist", false},
		{"housekeeping", "housekeeping", false},
	}

	for _, tt := range tests {
		t.Run(tt.caller+"_creates_"+tt.target, func(t *testing.T) {
			result := canCreateRole(tt.caller, tt.target)
			if result != tt.can {
				t.Errorf("expected %v for %s creating %s", tt.can, tt.caller, tt.target)
			}
		})
	}
}

func TestGenerateRandomToken(t *testing.T) {
	t.Run("generates non-empty token", func(t *testing.T) {
		token, err := generateRandomToken(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Error("token should not be empty")
		}
		// 32 bytes hex-encoded = 64 chars
		if len(token) != 64 {
			t.Errorf("expected 64 chars (32 bytes hex), got %d", len(token))
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		tokens := make(map[string]bool)
		for i := 0; i < 100; i++ {
			token, err := generateRandomToken(32)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tokens[token] {
				t.Fatal("generated duplicate token")
			}
			tokens[token] = true
		}
	})
}
