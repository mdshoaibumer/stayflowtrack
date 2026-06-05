-- name: CreateTenant :one
INSERT INTO tenants (name, slug, email, phone, settings)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: UpdateTenantStatus :exec
UPDATE tenants SET status = $2 WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (tenant_id, role_id, email, password_hash, first_name, last_name, phone)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByID :one
SELECT u.*, r.name as role_name, r.permissions as role_permissions
FROM users u
JOIN roles r ON u.role_id = r.id
WHERE u.id = $1;

-- name: GetUserByEmail :one
SELECT u.*, r.name as role_name, r.permissions as role_permissions
FROM users u
JOIN roles r ON u.role_id = r.id
WHERE u.email = $1 AND u.tenant_id = $2;

-- name: GetUserByEmailGlobal :one
SELECT u.*, r.name as role_name, r.permissions as role_permissions
FROM users u
JOIN roles r ON u.role_id = r.id
WHERE u.email = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = NOW() WHERE id = $1;

-- name: SetPasswordResetToken :exec
UPDATE users SET password_reset_token = $2, password_reset_expires_at = $3 WHERE id = $1;

-- name: GetUserByResetToken :one
SELECT * FROM users WHERE password_reset_token = $1 AND password_reset_expires_at > NOW();

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, password_reset_token = NULL, password_reset_expires_at = NULL WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users SET is_active = false WHERE id = $1 AND tenant_id = $2;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked = false AND expires_at > NOW();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens SET revoked = true WHERE user_id = $1;

-- name: DeleteExpiredTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = true;
