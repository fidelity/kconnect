package cluster

import (
	"github.com/fidelity/kconnect/pkg/provider/resolver"
	"github.com/spf13/pflag"
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

func (p *AKSClusterProvider) Flags() *pflag.FlagSet {
	set := &pflag.FlagSet{}

	set.AddFlag(&pflag.Flag{
		Name:      "resource-group",
		Shorthand: "r",
		Usage:     "The Azure resource group  to use",
		DefValue:  "",
	})

	return set
}

//func (p *AKSClusterProvider) Discover(identity identity.Identity, options map[string]string) (map[credentials][]clusters, error) {
//}
