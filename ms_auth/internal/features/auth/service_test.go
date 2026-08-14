package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"ms_auth/internal/core/domain"
	"ms_auth/internal/core/domain/apiError"
	"ms_auth/internal/core/security"
	userPkg "ms_auth/internal/features/user"
)

type mockJWTService struct {
	createTokenFn   func(user domain.UserDetails, tokenType security.TokenType) (string, error)
	validateTokenFn func(tokenString string, expectedType security.TokenType) (*security.TokenClaims, error)
	getAccessExpFn  func() time.Duration
	getRefreshExpFn func() time.Duration
}

func (m *mockJWTService) CreateToken(user domain.UserDetails, tokenType security.TokenType) (string, error) {
	if m.createTokenFn != nil {
		return m.createTokenFn(user, tokenType)
	}
	return "", nil
}

func (m *mockJWTService) ValidateToken(tokenString string, expectedType security.TokenType) (*security.TokenClaims, error) {
	if m.validateTokenFn != nil {
		return m.validateTokenFn(tokenString, expectedType)
	}
	return nil, errors.New("invalid token")
}

func (m *mockJWTService) GetAccessTokenExpiration() time.Duration {
	if m.getAccessExpFn != nil {
		return m.getAccessExpFn()
	}
	return time.Hour
}

func (m *mockJWTService) GetRefreshTokenExpiration() time.Duration {
	if m.getRefreshExpFn != nil {
		return m.getRefreshExpFn()
	}
	return 24 * time.Hour
}

type mockUserService struct {
	findByEmailFn func(ctx context.Context, email string) (*userPkg.User, error)
}

func (m *mockUserService) FindByEmail(ctx context.Context, email string) (*userPkg.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, apiError.ErrRecordNotFound
}

type mockRefreshTokenRepo struct {
	findByTokenHashFn   func(ctx context.Context, tokenHash string) (*RefreshToken, error)
	insertFn            func(ctx context.Context, model *RefreshToken) error
	updateFn            func(ctx context.Context, model *RefreshToken) error
	revokeAllByFamilyFn func(ctx context.Context, family uuid.UUID) error
	deleteByIdFn        func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRefreshTokenRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	if m.findByTokenHashFn != nil {
		return m.findByTokenHashFn(ctx, tokenHash)
	}
	return nil, apiError.ErrRecordNotFound
}

func (m *mockRefreshTokenRepo) Insert(ctx context.Context, model *RefreshToken) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, model)
	}
	return nil
}

func (m *mockRefreshTokenRepo) Update(ctx context.Context, model *RefreshToken) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, model)
	}
	return nil
}

func (m *mockRefreshTokenRepo) RevokeAllByFamily(ctx context.Context, family uuid.UUID) error {
	if m.revokeAllByFamilyFn != nil {
		return m.revokeAllByFamilyFn(ctx, family)
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteById(ctx context.Context, id uuid.UUID) error {
	if m.deleteByIdFn != nil {
		return m.deleteByIdFn(ctx, id)
	}
	return nil
}

type mockTxManager struct {
	err error
}

func (m *mockTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.err != nil {
		return m.err
	}
	return fn(ctx)
}

func newValidAuthUser() *userPkg.User {
	password := "Senha123A"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return &userPkg.User{
		ID:           uuid.New(),
		Email:        "joao@empresa.com",
		Name:         "João",
		Activated:    true,
		PasswordHash: hash,
	}
}

func newStoredRefreshToken(revoked bool) *RefreshToken {
	return &RefreshToken{
		ID:        uuid.New(),
		TokenHash: "hash-do-token",
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(time.Hour),
		Family:    uuid.New(),
		Revoked:   revoked,
	}
}

func TestAuthenticate_Success(t *testing.T) {
	validUser := newValidAuthUser()
	repoInsertCalled := false
	repo := &mockRefreshTokenRepo{
		insertFn: func(ctx context.Context, model *RefreshToken) error {
			repoInsertCalled = true
			return nil
		},
	}
	jwt := &mockJWTService{
		createTokenFn: func(user domain.UserDetails, tokenType security.TokenType) (string, error) {
			if tokenType == security.TokenTypeAccess {
				return "access-token", nil
			}
			return "refresh-token", nil
		},
		getAccessExpFn:  func() time.Duration { return time.Hour },
		getRefreshExpFn: func() time.Duration { return 24 * time.Hour },
	}
	users := &mockUserService{
		findByEmailFn: func(ctx context.Context, email string) (*userPkg.User, error) {
			return validUser, nil
		},
	}

	svc := NewService(jwt, users, repo, &mockTxManager{})

	resp, err := svc.Authenticate(context.Background(), validUser.Email, "Senha123A")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.AccessToken != "access-token" || resp.RefreshToken != "refresh-token" {
		t.Errorf("Unexpected tokens: access=%q refresh=%q", resp.AccessToken, resp.RefreshToken)
	}
	if !repoInsertCalled {
		t.Error("Insert should be called to save refresh token")
	}
}

func TestAuthenticate_InvalidCredentials_UserNotFound(t *testing.T) {
	repoInsertCalled := false
	repo := &mockRefreshTokenRepo{
		insertFn: func(ctx context.Context, model *RefreshToken) error {
			repoInsertCalled = true
			return nil
		},
	}
	jwt := &mockJWTService{}
	users := &mockUserService{
		findByEmailFn: func(ctx context.Context, email string) (*userPkg.User, error) {
			return nil, apiError.ErrRecordNotFound
		},
	}

	svc := NewService(jwt, users, repo, &mockTxManager{})

	_, err := svc.Authenticate(context.Background(), "naoexiste@empresa.com", "qualquer123A")

	if !errors.Is(err, apiError.ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
	if repoInsertCalled {
		t.Error("Insert should NOT be called for invalid credentials")
	}
}

func TestAuthenticate_InactiveAccount(t *testing.T) {
	user := newValidAuthUser()
	user.Activated = false

	repo := &mockRefreshTokenRepo{}
	jwt := &mockJWTService{}
	users := &mockUserService{
		findByEmailFn: func(ctx context.Context, email string) (*userPkg.User, error) {
			return user, nil
		},
	}

	svc := NewService(jwt, users, repo, &mockTxManager{})

	_, err := svc.Authenticate(context.Background(), user.Email, "Senha123A")

	if !errors.Is(err, apiError.ErrInactiveAccount) {
		t.Errorf("Expected ErrInactiveAccount, got %v", err)
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	user := newValidAuthUser()

	repo := &mockRefreshTokenRepo{}
	jwt := &mockJWTService{}
	users := &mockUserService{
		findByEmailFn: func(ctx context.Context, email string) (*userPkg.User, error) {
			return user, nil
		},
	}

	svc := NewService(jwt, users, repo, &mockTxManager{})

	_, err := svc.Authenticate(context.Background(), user.Email, "SenhaErrada1")

	if !errors.Is(err, apiError.ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	user := newValidAuthUser()
	stored := newStoredRefreshToken(false)

	updateCalled := false
	insertCalled := false
	repo := &mockRefreshTokenRepo{
		findByTokenHashFn: func(ctx context.Context, tokenHash string) (*RefreshToken, error) {
			return stored, nil
		},
		updateFn: func(ctx context.Context, model *RefreshToken) error {
			updateCalled = true
			return nil
		},
		insertFn: func(ctx context.Context, model *RefreshToken) error {
			insertCalled = true
			return nil
		},
	}
	jwt := &mockJWTService{
		validateTokenFn: func(tokenString string, expectedType security.TokenType) (*security.TokenClaims, error) {
			return &security.TokenClaims{
				Username: user.Email,
				IsAtivo:  true,
				Type:     security.TokenTypeRefresh,
			}, nil
		},
		createTokenFn: func(user domain.UserDetails, tokenType security.TokenType) (string, error) {
			if tokenType == security.TokenTypeAccess {
				return "new-access-token", nil
			}
			return "new-refresh-token", nil
		},
		getAccessExpFn:  func() time.Duration { return time.Hour },
		getRefreshExpFn: func() time.Duration { return 24 * time.Hour },
	}
	users := &mockUserService{
		findByEmailFn: func(ctx context.Context, email string) (*userPkg.User, error) {
			return user, nil
		},
	}

	svc := NewService(jwt, users, repo, &mockTxManager{})

	resp, err := svc.RefreshToken(context.Background(), "refresh-token-antigo")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.AccessToken != "new-access-token" || resp.RefreshToken != "new-refresh-token" {
		t.Errorf("Unexpected tokens: access=%q refresh=%q", resp.AccessToken, resp.RefreshToken)
	}
	if !updateCalled {
		t.Error("Update should be called to revoke old token")
	}
	if !insertCalled {
		t.Error("Insert should be called for new refresh token")
	}
}

func TestRefreshToken_RevokedToken_RevokesFamily(t *testing.T) {
	user := newValidAuthUser()
	stored := newStoredRefreshToken(true)

	revokeAllCalled := false
	updateCalled := false
	insertCalled := false
	repo := &mockRefreshTokenRepo{
		findByTokenHashFn: func(ctx context.Context, tokenHash string) (*RefreshToken, error) {
			return stored, nil
		},
		revokeAllByFamilyFn: func(ctx context.Context, family uuid.UUID) error {
			revokeAllCalled = true
			return nil
		},
		updateFn: func(ctx context.Context, model *RefreshToken) error {
			updateCalled = true
			return nil
		},
		insertFn: func(ctx context.Context, model *RefreshToken) error {
			insertCalled = true
			return nil
		},
	}
	jwt := &mockJWTService{
		validateTokenFn: func(tokenString string, expectedType security.TokenType) (*security.TokenClaims, error) {
			return &security.TokenClaims{Username: user.Email, IsAtivo: true, Type: security.TokenTypeRefresh}, nil
		},
	}
	users := &mockUserService{
		findByEmailFn: func(ctx context.Context, email string) (*userPkg.User, error) {
			return user, nil
		},
	}

	svc := NewService(jwt, users, repo, &mockTxManager{})

	_, err := svc.RefreshToken(context.Background(), "refresh-token-revogado")

	if err == nil {
		t.Fatal("Expected error for revoked token")
	}
	if !revokeAllCalled {
		t.Error("RevokeAllByFamily should be called when token revoked is reused")
	}
	if updateCalled || insertCalled {
		t.Error("Update/Insert should NOT be called when token is revoked")
	}
}

func TestLogout_Success(t *testing.T) {
	stored := newStoredRefreshToken(false)
	updateCalled := false
	revokeAllCalled := false
	repo := &mockRefreshTokenRepo{
		findByTokenHashFn: func(ctx context.Context, tokenHash string) (*RefreshToken, error) {
			return stored, nil
		},
		updateFn: func(ctx context.Context, model *RefreshToken) error {
			updateCalled = true
			return nil
		},
		revokeAllByFamilyFn: func(ctx context.Context, family uuid.UUID) error {
			revokeAllCalled = true
			return nil
		},
	}
	jwt := &mockJWTService{}
	users := &mockUserService{}

	svc := NewService(jwt, users, repo, &mockTxManager{})

	err := svc.Logout(context.Background(), "refresh-token-valido")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !updateCalled {
		t.Error("Update should be called to revoke token")
	}
	if !revokeAllCalled {
		t.Error("RevokeAllByFamily should be called")
	}
}
