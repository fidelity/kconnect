package identity

import (
	"github.com/spf13/pflag"
	"github.com/urfave/cli/v2"
)

type SAMLIdentityProvider struct {
}

func (p *SAMLIdentityProvider) Name() string {
	return "saml"
}

func (p *SAMLIdentityProvider) Flags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("idp-endpoint", "", "The SAML endpoint")

	return flagSet
}

func (p *SAMLIdentityProvider) Authenticate(flags *[]cli.Flag) (*Identity, error) {
	return &Identity{}, nil
}
