package identity

import "errors"

var (
	ErrNotTokenIdentity = errors.New("not a token identity")
)

type TokenIdentity struct {
	token          string
	name           string
	idProviderName string
}

func NewTokenIdentity(name, token, idProviderName string) *TokenIdentity {
	return &TokenIdentity{
		token:          token,
		name:           name,
		idProviderName: idProviderName,
	}
}

func (t *TokenIdentity) Type() string {
	return "token"
}

func (t *TokenIdentity) Name() string {
	return t.name
}

func (t *TokenIdentity) IsExpired() bool {
	// TODO: handle properly
	return false
}

func (t *TokenIdentity) IdentityProviderName() string {
	return t.idProviderName
}

func (t *TokenIdentity) Token() string {
	return t.token
}
