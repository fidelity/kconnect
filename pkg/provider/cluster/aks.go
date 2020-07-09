package cluster

import (
	"github.com/fidelity/kconnect/pkg/provider/resolver"
	"github.com/urfave/cli/v2"
)

type AKSClusterProvider struct {
	resourceGroup *string
}

func (p *AKSClusterProvider) Name() string {
	return "aks"
}

func (p *AKSClusterProvider) FlagsResolver() resolver.FlagsResolver {
	return &resolver.AzureFlagsResolver{}
}

func (p *AKSClusterProvider) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "resource-group",
			Usage:       "The Azure resource group to use",
			Required:    true,
			Destination: p.resourceGroup,
		},
	}
}

//func (p *AKSClusterProvider) Discover(identity identity.Identity, options map[string]string) (map[credentials][]clusters, error) {
//}
