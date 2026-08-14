package auth

import (
	"ms_auth/internal/core/domain/models"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	models.BaseModel
	ID        uuid.UUID
	TokenHash string
	UserID    uuid.UUID
	ExpiresAt time.Time
	Family    uuid.UUID
	Revoked   bool
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
