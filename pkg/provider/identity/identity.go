package identity

import (
	"context"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
	provcfg "github.com/fidelity/kconnect/pkg/provider/config"
)

// Provider represents the interface used to implement an identity provider
// plugin. It provides authentication functionality.
type Provider interface {
	provider.Plugin
	provider.PluginPreReqs
	provcfg.Resolver

	// Authenticate will authenticate a user and return details of their identity.
	Authenticate(ctx context.Context, input *AuthenticateInput) (*AuthenticateOutput, error)
}

type ProviderCreatorFun func(input *provider.PluginCreationInput) (Provider, error)

type AuthenticateInput struct {
	ConfigSet config.ConfigurationSet
}

type AuthenticateOutput struct {
	Identity provider.Identity
}
