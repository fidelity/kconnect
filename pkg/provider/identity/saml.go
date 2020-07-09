package identity

import "github.com/urfave/cli/v2"

type SAMLIdentityProvider struct {
}

func (p *SAMLIdentityProvider) Name() string {
	return "saml"
}

func (p *SAMLIdentityProvider) Flags() []cli.Flag {

	return []cli.Flag{
		&cli.StringFlag{
			Name:     "idp-endpoint",
			Usage:    "specify the SAML endpoint",
			Required: true,
		},
	}
}

func (p *SAMLIdentityProvider) Authenticate(flags *[]cli.Flag) (*Identity, error) {
	return &Identity{}, nil
}
