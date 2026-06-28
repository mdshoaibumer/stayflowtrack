package config

import (
	"testing"
)

func setEnvs(t *testing.T, envs map[string]string) {
	t.Helper()
	for k, v := range envs {
		t.Setenv(k, v)
	}
}

func minimalEnvs() map[string]string {
	return map[string]string{
		"DB_PASSWORD":        "test-password",
		"JWT_ACCESS_SECRET":  "test-access-secret-dev",
		"JWT_REFRESH_SECRET": "test-refresh-secret-dev",
	}
}

func TestLoad_MinimalConfig(t *testing.T) {
	setEnvs(t, minimalEnvs())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.App.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.App.Port)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected default host localhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("expected default MaxOpenConns 25, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Notifications.Provider != "log" {
		t.Errorf("expected default notification provider 'log', got %s", cfg.Notifications.Provider)
	}
}

func TestLoad_MissingDBPassword(t *testing.T) {
	t.Setenv("JWT_ACCESS_SECRET", "test-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh")
	t.Setenv("DB_PASSWORD", "")

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing DB_PASSWORD")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("DB_PASSWORD", "test-password")
	t.Setenv("JWT_ACCESS_SECRET", "")
	t.Setenv("JWT_REFRESH_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing JWT secrets")
	}
}

func TestLoad_ProductionValidation_WeakSecret(t *testing.T) {
	setEnvs(t, map[string]string{
		"APP_ENV":            "production",
		"DB_PASSWORD":        "a-production-db-credential-minimum-32chars-long-ok",
		"DB_SSL_MODE":        "require",
		"JWT_ACCESS_SECRET":  "short",
		"JWT_REFRESH_SECRET": "short",
	})

	_, err := Load()
	if err == nil {
		t.Error("expected error for short JWT secrets in production")
	}
}

func TestLoad_ProductionValidation_SameSecrets(t *testing.T) {
	secret := "this-is-a-very-long-secret-that-passes-the-32-char-check"
	setEnvs(t, map[string]string{
		"APP_ENV":            "production",
		"DB_PASSWORD":        "a-production-db-credential-minimum-32chars-long-ok",
		"DB_SSL_MODE":        "require",
		"JWT_ACCESS_SECRET":  secret,
		"JWT_REFRESH_SECRET": secret,
	})

	_, err := Load()
	if err == nil {
		t.Error("expected error when access and refresh secrets are the same")
	}
}

func TestLoad_ProductionValidation_UnsafeDefault(t *testing.T) {
	setEnvs(t, map[string]string{
		"APP_ENV":            "production",
		"DB_PASSWORD":        "a-production-db-credential-minimum-32chars-long-ok",
		"DB_SSL_MODE":        "require",
		"JWT_ACCESS_SECRET":  "dev-access-secret-change-in-production-min-32-chars",
		"JWT_REFRESH_SECRET": "prod-refresh-secret-something-different-and-long-enough",
	})

	_, err := Load()
	if err == nil {
		t.Error("expected error for unsafe 'dev' keyword in secret")
	}
}

func TestLoad_ProductionValidation_Valid(t *testing.T) {
	setEnvs(t, map[string]string{
		"APP_ENV":            "production",
		"DB_PASSWORD":        "a-production-db-credential-minimum-32chars-long-ok",
		"DB_SSL_MODE":        "require",
		"DB_APP_USER":        "stayflow_app",
		"DB_APP_PASSWORD":    "app-role-password-for-rls-testing",
		"JWT_ACCESS_SECRET":  "production-access-secret-that-is-long-enough-and-secure",
		"JWT_REFRESH_SECRET": "production-refresh-secret-that-is-also-long-and-different",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error for valid production config: %v", err)
	}
	if cfg.App.Env != "production" {
		t.Errorf("expected production env, got %s", cfg.App.Env)
	}
}

func TestLoad_CustomPorts(t *testing.T) {
	setEnvs(t, map[string]string{
		"DB_PASSWORD":        "test-password",
		"JWT_ACCESS_SECRET":  "test-access-secret-dev",
		"JWT_REFRESH_SECRET": "test-refresh-secret-dev",
		"APP_PORT":           "9090",
		"DB_PORT":            "5433",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.App.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.App.Port)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("expected DB port 5433, got %d", cfg.Database.Port)
	}
}

func TestLoad_InvalidDBPort(t *testing.T) {
	setEnvs(t, map[string]string{
		"DB_PASSWORD":        "test-password",
		"JWT_ACCESS_SECRET":  "test-secret",
		"JWT_REFRESH_SECRET": "test-refresh",
		"DB_PORT":            "not-a-number",
	})

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid DB_PORT")
	}
}

func TestLoad_NotificationConfig(t *testing.T) {
	setEnvs(t, map[string]string{
		"DB_PASSWORD":                 "test-password",
		"JWT_ACCESS_SECRET":           "test-access-secret-dev",
		"JWT_REFRESH_SECRET":          "test-refresh-secret-dev",
		"NOTIFICATION_PROVIDER":       "gupshup",
		"GUPSHUP_API_KEY":             "my-api-key",
		"GUPSHUP_APP_NAME":            "my-app",
		"NOTIFICATION_WEBHOOK_SECRET": "webhook-secret",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Notifications.Provider != "gupshup" {
		t.Errorf("expected provider gupshup, got %s", cfg.Notifications.Provider)
	}
	if cfg.Notifications.GupshupAPIKey != "my-api-key" {
		t.Errorf("expected API key 'my-api-key', got %s", cfg.Notifications.GupshupAPIKey)
	}
	if cfg.Notifications.GupshupApp != "my-app" {
		t.Errorf("expected app 'my-app', got %s", cfg.Notifications.GupshupApp)
	}
	if cfg.Notifications.WebhookSecret != "webhook-secret" {
		t.Errorf("expected webhook secret, got %s", cfg.Notifications.WebhookSecret)
	}
}

func TestLoad_StorageConfig(t *testing.T) {
	setEnvs(t, map[string]string{
		"DB_PASSWORD":            "test-password",
		"JWT_ACCESS_SECRET":      "test-access-secret-dev",
		"JWT_REFRESH_SECRET":     "test-refresh-secret-dev",
		"STORAGE_PROVIDER":       "s3",
		"STORAGE_ENDPOINT":       "http://minio:9000",
		"STORAGE_BUCKET":         "test-bucket",
		"STORAGE_USE_PATH_STYLE": "true",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Storage.Provider != "s3" {
		t.Errorf("expected provider s3, got %s", cfg.Storage.Provider)
	}
	if cfg.Storage.Bucket != "test-bucket" {
		t.Errorf("expected bucket test-bucket, got %s", cfg.Storage.Bucket)
	}
	if !cfg.Storage.UsePathStyle {
		t.Error("expected UsePathStyle=true")
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "myuser",
		Password: "mypass",
		Name:     "mydb",
		SSLMode:  "disable",
	}

	expected := "postgres://myuser:mypass@localhost:5432/mydb?sslmode=disable"
	if cfg.DSN() != expected {
		t.Errorf("DSN() = %q, want %q", cfg.DSN(), expected)
	}
}

func TestLoad_CORSMultipleOrigins(t *testing.T) {
	setEnvs(t, map[string]string{
		"DB_PASSWORD":          "test-password",
		"JWT_ACCESS_SECRET":    "test-access-secret-dev",
		"JWT_REFRESH_SECRET":   "test-refresh-secret-dev",
		"CORS_ALLOWED_ORIGINS": "http://localhost:3000,https://app.example.com",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Errorf("expected 2 origins, got %d", len(cfg.CORS.AllowedOrigins))
	}
}

func TestLoad_ProductionValidation_SSLDisabled(t *testing.T) {
	setEnvs(t, map[string]string{
		"APP_ENV":            "production",
		"DB_PASSWORD":        "a-production-db-credential-minimum-32chars-long-ok",
		"DB_SSL_MODE":        "disable",
		"DB_APP_USER":        "stayflow_app",
		"DB_APP_PASSWORD":    "app-role-password-for-rls-testing",
		"JWT_ACCESS_SECRET":  "production-access-secret-that-is-long-enough-and-secure",
		"JWT_REFRESH_SECRET": "production-refresh-secret-that-is-also-long-and-different",
	})

	_, err := Load()
	if err == nil {
		t.Error("expected error for DB_SSL_MODE=disable in production")
	}
}

func TestLoad_ProductionValidation_MissingAppUser(t *testing.T) {
	setEnvs(t, map[string]string{
		"APP_ENV":            "production",
		"DB_PASSWORD":        "a-production-db-credential-minimum-32chars-long-ok",
		"DB_SSL_MODE":        "require",
		"DB_APP_USER":        "",
		"DB_APP_PASSWORD":    "",
		"JWT_ACCESS_SECRET":  "production-access-secret-that-is-long-enough-and-secure",
		"JWT_REFRESH_SECRET": "production-refresh-secret-that-is-also-long-and-different",
	})

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing DB_APP_USER in production")
	}
}

func TestLoad_ProductionValidation_UnsafeDBPassword(t *testing.T) {
	setEnvs(t, map[string]string{
		"APP_ENV":            "production",
		"DB_PASSWORD":        "please-change-this-password-its-too-long-enough",
		"DB_SSL_MODE":        "require",
		"DB_APP_USER":        "stayflow_app",
		"DB_APP_PASSWORD":    "app-role-password-for-rls-testing",
		"JWT_ACCESS_SECRET":  "production-access-secret-that-is-long-enough-and-secure",
		"JWT_REFRESH_SECRET": "production-refresh-secret-that-is-also-long-and-different",
	})

	_, err := Load()
	if err == nil {
		t.Error("expected error for DB_PASSWORD containing 'change'")
	}
}

func TestLoad_ProductionValidation_ShortDBPassword(t *testing.T) {
	setEnvs(t, map[string]string{
		"APP_ENV":            "production",
		"DB_PASSWORD":        "short-pass",
		"DB_SSL_MODE":        "require",
		"DB_APP_USER":        "stayflow_app",
		"DB_APP_PASSWORD":    "app-role-password-for-rls-testing",
		"JWT_ACCESS_SECRET":  "production-access-secret-that-is-long-enough-and-secure",
		"JWT_REFRESH_SECRET": "production-refresh-secret-that-is-also-long-and-different",
	})

	_, err := Load()
	if err == nil {
		t.Error("expected error for DB_PASSWORD < 32 chars in production")
	}
}

func TestLoad_EmailConfig(t *testing.T) {
	setEnvs(t, map[string]string{
		"DB_PASSWORD":        "test-password",
		"JWT_ACCESS_SECRET":  "test-access-secret-dev",
		"JWT_REFRESH_SECRET": "test-refresh-secret-dev",
		"SMTP_ENABLED":       "true",
		"SMTP_HOST":          "smtp.example.com",
		"SMTP_PORT":          "465",
		"SMTP_USER":          "user@example.com",
		"SMTP_PASSWORD":      "smtp-pass",
		"SMTP_FROM_ADDRESS":  "noreply@example.com",
		"SMTP_FROM_NAME":     "My Hotel",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if !cfg.Email.Enabled {
		t.Error("expected email enabled=true")
	}
	if cfg.Email.SMTPHost != "smtp.example.com" {
		t.Errorf("expected smtp.example.com, got %s", cfg.Email.SMTPHost)
	}
	if cfg.Email.SMTPPort != 465 {
		t.Errorf("expected port 465, got %d", cfg.Email.SMTPPort)
	}
	if cfg.Email.FromAddress != "noreply@example.com" {
		t.Errorf("expected noreply@example.com, got %s", cfg.Email.FromAddress)
	}
}

func TestLoad_RazorpayConfig(t *testing.T) {
	setEnvs(t, map[string]string{
		"DB_PASSWORD":             "test-password",
		"JWT_ACCESS_SECRET":       "test-access-secret-dev",
		"JWT_REFRESH_SECRET":      "test-refresh-secret-dev",
		"RAZORPAY_KEY_ID":         "rzp_test_123",
		"RAZORPAY_KEY_SECRET":     "secret_456",
		"RAZORPAY_WEBHOOK_SECRET": "whsec_789",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Razorpay.KeyID != "rzp_test_123" {
		t.Errorf("expected razorpay key, got %s", cfg.Razorpay.KeyID)
	}
	if cfg.Razorpay.KeySecret != "secret_456" {
		t.Errorf("expected razorpay secret")
	}
	if cfg.Razorpay.WebhookSecret != "whsec_789" {
		t.Errorf("expected webhook secret")
	}
}

func TestLoad_JWTExpirations(t *testing.T) {
	setEnvs(t, map[string]string{
		"DB_PASSWORD":            "test-password",
		"JWT_ACCESS_SECRET":      "test-access-secret-dev",
		"JWT_REFRESH_SECRET":     "test-refresh-secret-dev",
		"JWT_ACCESS_EXPIRATION":  "30m",
		"JWT_REFRESH_EXPIRATION": "720h",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.JWT.AccessExpiration.Minutes() != 30 {
		t.Errorf("expected 30m access expiration, got %v", cfg.JWT.AccessExpiration)
	}
	if cfg.JWT.RefreshExpiration.Hours() != 720 {
		t.Errorf("expected 720h refresh expiration, got %v", cfg.JWT.RefreshExpiration)
	}
}
