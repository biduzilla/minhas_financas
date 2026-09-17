package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"ms_auth/internal/core/config"
	"os"
	"shared/auth/domain"
	sharedsecurity "shared/auth/security"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	*sharedsecurity.JwtService
	config     config.Config
	privateKey *rsa.PrivateKey
}

func NewService(
	config config.Config,
) (*JwtService, error) {
	service := &JwtService{
		config: config,
	}

	if err := service.loadKeys(); err != nil {
		return nil, fmt.Errorf("failed to load RSA keys: %w", err)
	}

	sharedSvc, err := sharedsecurity.NewService(config.Base)
	if err != nil {
		return nil, err
	}

	service.JwtService = sharedSvc

	return service, nil
}

func (s *JwtService) loadKeys() error {
	privateKey, err := loadRSAPrivateKey(s.config.Base.Security.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}
	s.privateKey = privateKey

	return nil
}

func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("private key file not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}

	return rsaKey, nil
}

func (s *JwtService) CreateToken(
	user domain.UserDetails,
	tokenType sharedsecurity.TokenType,
) (string, error) {
	var expiration time.Duration

	switch tokenType {
	case sharedsecurity.TokenTypeAccess:
		expiration = sharedsecurity.AccessTokenExpiration
	case sharedsecurity.TokenTypeRefresh:
		expiration = sharedsecurity.RefreshTokenExpiration
	default:
		expiration = 0
	}

	now := time.Now()
	claims := sharedsecurity.TokenClaims{
		UserID:   user.GetID(),
		Username: user.GetUsername(),
		IsAtivo:  user.GetIsAtivo(),
		Type:     tokenType,
		Roles:    user.GetRoles(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    sharedsecurity.TokenIssuer,
			Audience:  jwt.ClaimStrings{sharedsecurity.TokenAudience},
			Subject:   user.GetID().String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}
