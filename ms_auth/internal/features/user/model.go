package user

import (
	"errors"
	"ms_auth/internal/core/domain/models"
	"ms_auth/internal/core/validator"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	models.BaseModel
	ID           uuid.UUID
	Email        string
	PasswordHash []byte
	Name         string
	Activated    bool
}

type CreateUserDTO struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
	Name     string `json:"name"`
}

type UserDTO struct {
	ID      *uuid.UUID `json:"id"`
	Email   *string    `json:"email"`
	Name    *string    `json:"name"`
	Version *int       `json:"version"`
}

func (m *User) ToDTO() UserDTO {
	return UserDTO{
		ID:      &m.ID,
		Email:   &m.Email,
		Name:    &m.Name,
		Version: &m.Version,
	}
}

func (d CreateUserDTO) ToModel() (*User, error) {
	passwordHash, err := d.HashPassword()
	if err != nil {
		return nil, err
	}
	return &User{
		Email:        d.Email,
		PasswordHash: passwordHash,
		Name:         d.Name,
	}, nil
}

func (d UserDTO) ToModel() *User {
	var model User

	if d.Email != nil {
		model.Email = *d.Email
	}

	if d.ID != nil {
		model.ID = *d.ID
	}

	if d.Name != nil {
		model.Name = *d.Name
	}

	if d.Version != nil {
		model.Version = *d.Version
	}

	return &model
}

func (d CreateUserDTO) HashPassword() ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(d.Password), 12)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (m *User) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(m.PasswordHash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (m *User) Validate(v *validator.Validator) {
	v.Check(m.Name != "", "nome", "must be provided")
	v.Check(len(m.Name) >= 3, "nome", "must be at least 3 characters long")
	v.Check(len(m.Name) <= 100, "nome", "must not be more than 100 characters long")
	v.Check(m.Email != "", "email", "must be provided")
	v.Check(validator.Matches(m.Email, validator.EmailRX), "email", "must be a valid email address")
}

func validatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")

	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	v.Check(hasLetter, "password", "must contain at least one letter")
	v.Check(hasDigit, "password", "must contain at least one digit")
}

func (u *User) GetID() uuid.UUID {
	return u.ID
}

func (u *User) GetIsAtivo() bool {
	return u.Activated
}

func (u *User) GetUsername() string { return u.Email }

func (u *User) IsAnonymous() bool {
	return false
}

func (u *User) GetRoles() []string {
	return nil
}
