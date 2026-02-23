package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	appauth "ava-sales/internal/auth"
	"ava-sales/internal/models"
	"ava-sales/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	tx               repository.TxManager
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type MeResponse struct {
	ID    uuid.UUID       `json:"id"`
	Email string          `json:"email"`
	Role  models.ActorRole `json:"role"`
}

func NewAuthService(tx repository.TxManager, jwtSecret string, accessTokenTTLMinutes int, refreshTokenTTLDays int) *AuthService {
	return &AuthService{
		tx:              tx,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  time.Duration(accessTokenTTLMinutes) * time.Minute,
		refreshTokenTTL: time.Duration(refreshTokenTTLDays) * 24 * time.Hour,
	}
}

func (s *AuthService) Login(ctx context.Context, identity string, password string) (*TokenResponse, error) {
	identity = strings.TrimSpace(identity)
	if identity == "" || strings.TrimSpace(password) == "" {
		return nil, models.NewAPIError("VALIDATION_ERROR", "email/username and password are required", http.StatusBadRequest, nil)
	}

	user, err := s.tx.Repo().GetAppUserByEmailOrUsername(ctx, identity)
	if err != nil {
		return nil, models.NewAPIError("UNAUTHORIZED", "invalid credentials", http.StatusUnauthorized, nil)
	}
	if !user.IsActive {
		return nil, models.NewAPIError("UNAUTHORIZED", "invalid credentials", http.StatusUnauthorized, nil)
	}
	if strings.TrimSpace(user.PasswordHash) == "" {
		return nil, models.NewAPIError("UNAUTHORIZED", "invalid credentials", http.StatusUnauthorized, nil)
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); compareErr != nil {
		return nil, models.NewAPIError("UNAUTHORIZED", "invalid credentials", http.StatusUnauthorized, nil)
	}

	return s.issueTokenPair(ctx, user, nil)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, models.NewAPIError("VALIDATION_ERROR", "refresh_token is required", http.StatusBadRequest, nil)
	}

	var out *TokenResponse
	err := s.tx.WithTx(ctx, func(repo repository.Repository) error {
		hash := appauth.HashToken(refreshToken)
		stored, getErr := repo.GetRefreshTokenByHash(ctx, hash)
		if getErr != nil {
			return models.NewAPIError("UNAUTHORIZED", "invalid refresh token", http.StatusUnauthorized, nil)
		}
		if stored.RevokedAt != nil || time.Now().UTC().After(stored.ExpiresAt) {
			return models.NewAPIError("UNAUTHORIZED", "invalid refresh token", http.StatusUnauthorized, nil)
		}

		user, userErr := repo.GetAppUserByID(ctx, stored.UserID)
		if userErr != nil || !user.IsActive {
			return models.NewAPIError("UNAUTHORIZED", "invalid refresh token", http.StatusUnauthorized, nil)
		}

		if revokeErr := repo.RevokeRefreshToken(ctx, stored.ID); revokeErr != nil {
			return revokeErr
		}

		pair, issueErr := s.issueTokenPairWithRepo(ctx, repo, user)
		if issueErr != nil {
			return issueErr
		}
		out = pair
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*MeResponse, error) {
	user, err := s.tx.Repo().GetAppUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(user.Email) == "" {
		return nil, models.NewAPIError("UNAUTHORIZED", "user profile incomplete", http.StatusUnauthorized, nil)
	}
	return &MeResponse{ID: user.ID, Email: user.Email, Role: user.Role}, nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *models.AppUser, repo repository.Repository) (*TokenResponse, error) {
	if repo != nil {
		return s.issueTokenPairWithRepo(ctx, repo, user)
	}

	var out *TokenResponse
	err := s.tx.WithTx(ctx, func(txRepo repository.Repository) error {
		pair, issueErr := s.issueTokenPairWithRepo(ctx, txRepo, user)
		if issueErr != nil {
			return issueErr
		}
		out = pair
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AuthService) issueTokenPairWithRepo(ctx context.Context, repo repository.Repository, user *models.AppUser) (*TokenResponse, error) {
	accessToken, expiresIn, err := appauth.GenerateAccessToken(s.jwtSecret, user.ID.String(), string(user.Role), s.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := appauth.NewOpaqueToken()
	if err != nil {
		return nil, err
	}
	tokenHash := appauth.HashToken(refreshToken)
	expiresAt := time.Now().UTC().Add(s.refreshTokenTTL)

	if err := repo.CreateRefreshToken(ctx, &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}
