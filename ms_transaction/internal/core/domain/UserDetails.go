package domain

import "github.com/google/uuid"

type UserDetails interface {
	GetID() uuid.UUID
	GetUsername() string
	GetIsAtivo() bool
	IsAnonymous() bool
	GetRoles() []string
}

type anonymousUser struct{}

func (anonymousUser) GetID() uuid.UUID    { return uuid.Nil }
func (anonymousUser) GetUsername() string { return "" }
func (anonymousUser) GetIsAtivo() bool    { return false }
func (anonymousUser) IsAnonymous() bool   { return true }
func (anonymousUser) GetRoles() []string  { return nil }

var AnonymousUser UserDetails = anonymousUser{}

type authenticatedUser struct {
	id       uuid.UUID
	username string
	isAtivo  bool
	roles    []string
}

func (u *authenticatedUser) GetID() uuid.UUID    { return u.id }
func (u *authenticatedUser) GetUsername() string { return u.username }
func (u *authenticatedUser) GetIsAtivo() bool    { return u.isAtivo }
func (u *authenticatedUser) IsAnonymous() bool   { return false }
func (u *authenticatedUser) GetRoles() []string  { return u.roles }

func NewAuthenticatedUser(
	id uuid.UUID,
	username string,
	isAtivo bool,
	roles []string,
) UserDetails {
	return &authenticatedUser{
		id:       id,
		username: username,
		isAtivo:  isAtivo,
		roles:    roles,
	}
}
