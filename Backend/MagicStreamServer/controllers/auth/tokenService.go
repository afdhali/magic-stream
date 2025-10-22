package authservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/config"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/models"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/repositories"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrRevokedToken = errors.New("token has been revoked")
)

type TokenService struct {
	cfg                      *config.Config
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewTokenService(cfg *config.Config, refreshTokenRepo repositories.RefreshTokenRepository) *TokenService {
	return &TokenService{
		cfg:              cfg,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// GenerateTokenPair creates new access + refresh tokens
func (ts *TokenService) GenerateTokenPair(userID string) (*models.TokenPair, error) {
	// === Access Token ===
	accessExp := time.Now().Add(time.Minute * time.Duration(ts.cfg.AccessTokenExpireMin))
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": accessExp.Unix(),
		"typ": "access",
	})
	accessStr, err := accessToken.SignedString([]byte(ts.cfg.JWTAccessSecret))
	if err != nil {
		return nil, err
	}

	// === Refresh Token (JWT string) ===
	refreshExp := time.Now().Add(time.Hour * time.Duration(ts.cfg.RefreshTokenExpireHr))
	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": refreshExp.Unix(),
		"typ": "refresh",
	})
	refreshStr, err := refreshJWT.SignedString([]byte(ts.cfg.JWTRefreshSecret))
	if err != nil {
		return nil, err
	}

	// === Store refresh token in DB ===
	refreshTokenDoc := models.RefreshToken{
		UserID:    userID,
		Token:     refreshStr,
		ExpiresAt: refreshExp,
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	ctx := context.TODO()
	err = ts.refreshTokenRepo.Create(ctx, &refreshTokenDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token in DB: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

// ValidateAccessToken validates a JWT access token and returns the user ID.
func (ts *TokenService) ValidateAccessToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(ts.cfg.JWTAccessSecret), nil
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	typ, ok := claims["typ"].(string)
	if !ok || typ != "access" {
		return "", ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// RevokeRefreshTokens revokes all active refresh tokens for a given user.
func (ts *TokenService) RevokeRefreshTokens(userID string) error {
	ctx := context.TODO()
	return ts.refreshTokenRepo.RevokeUserTokens(ctx, userID)
}

// UseRefreshToken validates a refresh token and issues a new token pair.
func (ts *TokenService) UseRefreshToken(refreshToken string) (*models.TokenPair, error) {
	// 1. Validate JWT signature and claims
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(ts.cfg.JWTRefreshSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if typ, ok := claims["typ"].(string); !ok || typ != "refresh" {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return nil, ErrInvalidToken
	}

	// 2. Look up token in DB
	ctx := context.TODO()
	stored, err := ts.refreshTokenRepo.FindByToken(ctx, refreshToken, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrRefreshTokenNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("database error while fetching refresh token: %w", err)
	}

	// 3. Check revocation and expiry
	if stored.Revoked {
		return nil, ErrRevokedToken
	}

	if time.Now().After(stored.ExpiresAt) {
		// Optional: mark as revoked if expired
		_ = ts.refreshTokenRepo.RevokeToken(ctx, stored.ID.Hex())
		return nil, ErrInvalidToken
	}

	// 4. Revoke current token (single-use policy)
	_ = ts.refreshTokenRepo.RevokeToken(ctx, stored.ID.Hex())

	// 5. Issue new token pair
	return ts.GenerateTokenPair(userID)
}

// CleanupExpiredRefreshTokens removes all expired refresh tokens from the database.
func (ts *TokenService) CleanupExpiredRefreshTokens() error {
	ctx := context.TODO()
	return ts.refreshTokenRepo.CleanupExpired(ctx)
}