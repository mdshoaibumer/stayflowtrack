package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stayflow/stayflow-track/internal/modules/auth/domain"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
	settings, _ := json.Marshal(tenant.Settings)

	err := r.pool.QueryRow(ctx,
		`INSERT INTO tenants (name, slug, email, phone, settings)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		tenant.Name, tenant.Slug, tenant.Email, tenant.Phone, settings,
	).Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create tenant: %w", err)
	}
	return nil
}

func (r *Repository) GetTenantByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var t domain.Tenant
	var settings []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, email, phone, status, settings, created_at, updated_at
		 FROM tenants WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Slug, &t.Email, &t.Phone, &t.Status, &settings, &t.CreatedAt, &t.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("tenant", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}

	_ = json.Unmarshal(settings, &t.Settings)
	return &t, nil
}

func (r *Repository) GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var t domain.Tenant
	var settings []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, email, phone, status, settings, created_at, updated_at
		 FROM tenants WHERE slug = $1`, slug,
	).Scan(&t.ID, &t.Name, &t.Slug, &t.Email, &t.Phone, &t.Status, &settings, &t.CreatedAt, &t.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tenant by slug: %w", err)
	}

	_ = json.Unmarshal(settings, &t.Settings)
	return &t, nil
}

func (r *Repository) CreateUser(ctx context.Context, user *domain.User) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (tenant_id, role_id, email, password_hash, first_name, last_name, phone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, is_active, created_at, updated_at`,
		user.TenantID, user.RoleID, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Phone,
	).Scan(&user.ID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	var perms []byte

	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.tenant_id, u.role_id, u.email, u.password_hash,
		        u.first_name, u.last_name, u.phone, u.is_active, u.last_login_at,
		        u.created_at, u.updated_at, r.name, r.permissions
		 FROM users u JOIN roles r ON u.role_id = r.id
		 WHERE u.id = $1`, id,
	).Scan(&u.ID, &u.TenantID, &u.RoleID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Phone, &u.IsActive, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.RoleName, &perms)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("user", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	_ = json.Unmarshal(perms, &u.RolePermissions)
	return &u, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	var perms []byte

	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.tenant_id, u.role_id, u.email, u.password_hash,
		        u.first_name, u.last_name, u.phone, u.is_active, u.last_login_at,
		        u.created_at, u.updated_at, r.name, r.permissions
		 FROM users u JOIN roles r ON u.role_id = r.id
		 WHERE u.email = $1`, email,
	).Scan(&u.ID, &u.TenantID, &u.RoleID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Phone, &u.IsActive, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.RoleName, &perms)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	_ = json.Unmarshal(perms, &u.RolePermissions)
	return &u, nil
}

// GetUserByEmailAndTenant looks up a user scoped to a specific tenant slug.
// This prevents cross-tenant authentication.
func (r *Repository) GetUserByEmailAndTenant(ctx context.Context, email, tenantSlug string) (*domain.User, error) {
	var u domain.User
	var perms []byte

	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.tenant_id, u.role_id, u.email, u.password_hash,
		        u.first_name, u.last_name, u.phone, u.is_active, u.last_login_at,
		        u.created_at, u.updated_at, r.name, r.permissions
		 FROM users u
		 JOIN roles r ON u.role_id = r.id
		 JOIN tenants t ON u.tenant_id = t.id
		 WHERE u.email = $1 AND t.slug = $2`, email, tenantSlug,
	).Scan(&u.ID, &u.TenantID, &u.RoleID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Phone, &u.IsActive, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.RoleName, &perms)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email and tenant: %w", err)
	}

	_ = json.Unmarshal(perms, &u.RolePermissions)
	return &u, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID)
	return err
}

func (r *Repository) SetPasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_reset_token = $2, password_reset_expires_at = $3 WHERE id = $1`,
		userID, token, expiresAt)
	return err
}

func (r *Repository) GetUserByResetToken(ctx context.Context, token string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, role_id, email, first_name, last_name
		 FROM users WHERE password_reset_token = $1 AND password_reset_expires_at > NOW()`, token,
	).Scan(&u.ID, &u.TenantID, &u.RoleID, &u.Email, &u.FirstName, &u.LastName)

	if err == pgx.ErrNoRows {
		return nil, apperrors.Unauthorized("invalid or expired reset token")
	}
	if err != nil {
		return nil, fmt.Errorf("get user by reset token: %w", err)
	}
	return &u, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, password_reset_token = NULL, password_reset_expires_at = NULL WHERE id = $1`,
		userID, passwordHash)
	return err
}

func (r *Repository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	var perms []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, permissions, created_at FROM roles WHERE name = $1`, name,
	).Scan(&role.ID, &role.Name, &role.Description, &perms, &role.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.NotFound("role", name)
	}
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	_ = json.Unmarshal(perms, &role.Permissions)
	return &role, nil
}

func (r *Repository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.pool.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3) RETURNING id, user_id, token_hash, expires_at, revoked, created_at`,
		userID, tokenHash, expiresAt,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}
	return &rt, nil
}

func (r *Repository) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked, created_at
		 FROM refresh_tokens WHERE token_hash = $1 AND revoked = false AND expires_at > NOW()`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, apperrors.Unauthorized("invalid refresh token")
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return &rt, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *Repository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`, userID)
	return err
}

// HashToken creates a SHA-256 hash of a token string.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
