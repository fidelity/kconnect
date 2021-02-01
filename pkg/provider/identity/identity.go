package identity

import (
	"context"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
)

// Provider represents the interface used to implement an identity provider
// plugin. It provides authentication functionality.
type Provider interface {
	provider.Plugin

	// Authenticate will authenticate a user and return details of their identity.
	Authenticate(ctx context.Context, input *AuthenticateInput) (*AuthenticateOutput, error)
}

type ProviderCreatorFun func(input *provider.PluginCreationInput) (Provider, error)

type AuthenticateInput struct {
	ConfigSet config.ConfigurationSet
}

type AuthenticateOutput struct {
	Identity Identity
}

// Identity represents a users identity for use with discovery.
// NOTE: details of this need finalising
type Identity interface {
	Type() string
	Name() string
	IsExpired() bool
	IdentityProviderName() string
}

// Store represents an way to store and retrieve credentials
type Store interface {
	CredsExists() (bool, error)
	Save(identity Identity) error
	Load() (Identity, error)
	Expired() bool
}
