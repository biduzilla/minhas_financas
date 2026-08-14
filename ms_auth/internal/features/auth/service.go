package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"ms_auth/internal/core/domain"
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/core/security"
	"ms_auth/internal/core/transaction"
	"ms_auth/internal/features/user"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	jwtService       jwtService
	userService      userService
	refreshTokenRepo repository
	tx               transaction.Manager
}

type jwtService interface {
	CreateToken(user domain.UserDetails, tokenType security.TokenType) (string, error)
	ValidateToken(tokenString string, expectedType security.TokenType) (*security.TokenClaims, error)
	GetAccessTokenExpiration() time.Duration
	GetRefreshTokenExpiration() time.Duration
}

type userService interface {
	FindByEmail(ctx context.Context, email string) (*user.User, error)
}

func NewService(
	jwt jwtService,
	users userService,
	refreshRepo repository,
	tx transaction.Manager,
) *AuthService {
	return &AuthService{
		jwtService:       jwt,
		userService:      users,
		refreshTokenRepo: refreshRepo,
		tx:               tx,
	}
}

func (s *AuthService) Authenticate(
	ctx context.Context,
	email, password string,
) (*TokenResponse, error) {
	user, err := s.userService.FindByEmail(ctx, email)
	if err != nil {
		return nil, apiError.ErrInvalidCredentials
	}

	if !user.Activated {
		return nil, apiError.ErrInactiveAccount
	}

	match, err := user.Matches(password)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, apiError.ErrInvalidCredentials
	}

	accessToken, err := s.jwtService.CreateToken(user, security.TokenTypeAccess)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwtService.CreateToken(user, security.TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	family := uuid.New()
	if err := s.saveRefreshToken(ctx, user.ID, refreshToken, family); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtService.GetAccessTokenExpiration().Seconds()),
	}, nil
}

func (s *AuthService) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (*TokenResponse, error) {
	claims, err := s.jwtService.ValidateToken(refreshToken, security.TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	user, err := s.userService.FindByEmail(ctx, claims.Username)
	if err != nil {
		return nil, apiError.ErrInvalidCredentials
	}

	tokenHash := hashToken(refreshToken)

	var accessToken, newRefreshToken string

	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		storedToken, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
		if err != nil {
			return err
		}

		if storedToken.Revoked {
			_ = s.refreshTokenRepo.RevokeAllByFamily(ctx, storedToken.Family)
			return apiError.NewHTTPError("refresh token has been revoked", http.StatusBadRequest, nil)
		}

		if storedToken.ExpiresAt.Before(time.Now()) {
			return apiError.NewHTTPError("refresh token expired", http.StatusBadRequest, nil)
		}

		storedToken.Revoked = true
		if err := s.refreshTokenRepo.Update(ctx, storedToken); err != nil {
			return err
		}

		accessToken, err = s.jwtService.CreateToken(user, security.TokenTypeAccess)
		if err != nil {
			return err
		}
		newRefreshToken, err = s.jwtService.CreateToken(user, security.TokenTypeRefresh)
		if err != nil {
			return err
		}

		if err := s.saveRefreshToken(ctx, user.ID, newRefreshToken, storedToken.Family); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtService.GetAccessTokenExpiration().Seconds()),
	}, nil
}

func (s *AuthService) Logout(
	ctx context.Context,
	refreshToken string,
) error {
	tokenHash := hashToken(refreshToken)

	return s.tx.RunInTx(ctx, func(ctx context.Context) error {
		stored, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
		if err != nil {
			return err
		}

		if stored.Revoked {
			return nil
		}

		stored.Revoked = true
		if err := s.refreshTokenRepo.Update(ctx, stored); err != nil {
			return err
		}

		return s.refreshTokenRepo.RevokeAllByFamily(ctx, stored.Family)
	})
}

func (s *AuthService) saveRefreshToken(
	ctx context.Context,
	userID uuid.UUID,
	token string,
	family uuid.UUID,
) error {
	now := time.Now()
	entity := &RefreshToken{
		TokenHash: hashToken(token),
		UserID:    userID,
		ExpiresAt: now.Add(s.jwtService.GetRefreshTokenExpiration()),
		Family:    family,
		Revoked:   false,
	}

	return s.refreshTokenRepo.Insert(ctx, entity)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
