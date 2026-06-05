package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/stayflow/stayflow-track/internal/config"
	"github.com/stayflow/stayflow-track/internal/modules/auth/domain"
	"github.com/stayflow/stayflow-track/internal/modules/auth/repository"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Service struct {
	repo   *repository.Repository
	jwtCfg config.JWTConfig
}

func New(repo *repository.Repository, jwtCfg config.JWTConfig) *Service {
	return &Service{repo: repo, jwtCfg: jwtCfg}
}

type RegisterTenantInput struct {
	TenantName string `json:"tenant_name" validate:"required,min=2,max=100"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8,max=72"`
	FirstName  string `json:"first_name" validate:"required,min=1,max=100"`
	LastName   string `json:"last_name" validate:"required,min=1,max=100"`
	Phone      string `json:"phone" validate:"omitempty,max=20"`
}

type CreateUserInput struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string `json:"phone" validate:"omitempty,max=20"`
	RoleName  string `json:"role_name" validate:"required,oneof=property_admin receptionist housekeeping"`
}

type LoginInput struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	TenantSlug string `json:"tenant_slug" validate:"omitempty"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type PasswordResetRequestInput struct {
	Email string `json:"email" validate:"required,email"`
}

type PasswordResetConfirmInput struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type RegisterTenantResult struct {
	Tenant *domain.Tenant    `json:"tenant"`
	User   *domain.User      `json:"user"`
	Tokens *domain.TokenPair `json:"tokens"`
}

func (s *Service) RegisterTenant(ctx context.Context, input RegisterTenantInput) (*RegisterTenantResult, error) {
	slug := generateSlug(input.TenantName)

	existing, err := s.repo.GetTenantBySlug(ctx, slug)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if existing != nil {
		return nil, apperrors.Conflict("a tenant with this name already exists")
	}

	existingUser, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if existingUser != nil {
		return nil, apperrors.Conflict("a user with this email already exists")
	}

	tenant := &domain.Tenant{
		Name:  input.TenantName,
		Slug:  slug,
		Email: input.Email,
		Phone: input.Phone,
	}

	if err := s.repo.CreateTenant(ctx, tenant); err != nil {
		return nil, apperrors.Internal(err)
	}

	role, err := s.repo.GetRoleByName(ctx, "super_admin")
	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("hash password: %w", err))
	}

	user := &domain.User{
		TenantID:     tenant.ID,
		RoleID:       role.ID,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Phone:        input.Phone,
		RoleName:     role.Name,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, apperrors.Internal(err)
	}

	tokens, err := s.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &RegisterTenantResult{
		Tenant: tenant,
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *Service) CreateUser(ctx context.Context, tenantID uuid.UUID, input CreateUserInput) (*domain.User, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if existingUser != nil {
		return nil, apperrors.Conflict("a user with this email already exists")
	}

	role, err := s.repo.GetRoleByName(ctx, input.RoleName)
	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("hash password: %w", err))
	}

	user := &domain.User{
		TenantID:     tenantID,
		RoleID:       role.ID,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Phone:        input.Phone,
		RoleName:     role.Name,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, apperrors.Internal(err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*domain.TokenPair, error) {
	var user *domain.User
	var err error

	if input.TenantSlug != "" {
		// Tenant-scoped login (prevents cross-tenant access)
		user, err = s.repo.GetUserByEmailAndTenant(ctx, input.Email, input.TenantSlug)
	} else {
		// Global lookup (backward compatible — will fail if email exists in multiple tenants)
		user, err = s.repo.GetUserByEmail(ctx, input.Email)
	}
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		return nil, apperrors.Unauthorized("invalid email or password")
	}

	if !user.IsActive {
		return nil, apperrors.Unauthorized("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, apperrors.Unauthorized("invalid email or password")
	}

	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, apperrors.Internal(err)
	}

	tokens, err := s.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Service) RefreshToken(ctx context.Context, input RefreshInput) (*domain.TokenPair, error) {
	tokenHash := repository.HashToken(input.RefreshToken)

	rt, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	// Revoke the old token (rotation)
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, apperrors.Internal(err)
	}

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, apperrors.Unauthorized("account is deactivated")
	}

	tokens, err := s.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.RevokeAllUserTokens(ctx, userID); err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, input PasswordResetRequestInput) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return "", apperrors.Internal(err)
	}
	if user == nil {
		// Don't reveal whether email exists
		return "", nil
	}

	token, err := generateRandomToken(32)
	if err != nil {
		return "", apperrors.Internal(err)
	}

	// Store hashed token — raw token is sent to user via email
	tokenHash := repository.HashToken(token)
	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.repo.SetPasswordResetToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return "", apperrors.Internal(err)
	}

	return token, nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, input PasswordResetConfirmInput) error {
	// Hash the incoming token to compare against stored hash
	tokenHash := repository.HashToken(input.Token)
	user, err := s.repo.GetUserByResetToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("hash password: %w", err))
	}

	if err := s.repo.UpdatePassword(ctx, user.ID, string(passwordHash)); err != nil {
		return apperrors.Internal(err)
	}

	// Revoke all existing tokens after password reset
	if err := s.repo.RevokeAllUserTokens(ctx, user.ID); err != nil {
		return apperrors.Internal(err)
	}

	return nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*domain.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtCfg.AccessSecret), nil
	})

	if err != nil {
		return nil, apperrors.Unauthorized("invalid access token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, apperrors.Unauthorized("invalid access token")
	}

	userID, err := uuid.Parse(mapClaims["sub"].(string))
	if err != nil {
		return nil, apperrors.Unauthorized("invalid token claims")
	}

	tenantID, err := uuid.Parse(mapClaims["tenant_id"].(string))
	if err != nil {
		return nil, apperrors.Unauthorized("invalid token claims")
	}

	return &domain.Claims{
		UserID:   userID,
		TenantID: tenantID,
		RoleName: mapClaims["role"].(string),
		Email:    mapClaims["email"].(string),
	}, nil
}

func (s *Service) GenerateTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(s.jwtCfg.AccessExpiration)

	accessClaims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"tenant_id": user.TenantID.String(),
		"role":      user.RoleName,
		"email":     user.Email,
		"iss":       s.jwtCfg.Issuer,
		"iat":       now.Unix(),
		"exp":       accessExp.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtCfg.AccessSecret))
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("sign access token: %w", err))
	}

	refreshTokenRaw, err := generateRandomToken(32)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	refreshTokenHash := repository.HashToken(refreshTokenRaw)
	refreshExp := now.Add(s.jwtCfg.RefreshExpiration)

	_, err = s.repo.CreateRefreshToken(ctx, user.ID, refreshTokenHash, refreshExp)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	return &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenRaw,
		ExpiresIn:    int64(s.jwtCfg.AccessExpiration.Seconds()),
	}, nil
}

func generateSlug(name string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		} else if r == ' ' || r == '-' {
			builder.WriteRune('-')
		}
	}
	return builder.String()
}

func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
