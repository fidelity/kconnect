package resolver

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/urfave/cli/v2"
)

type AzureFlagsResolver struct {
}

func (r *AzureFlagsResolver) Resolve(identity identity.Identity, flags []cli.Flag) error {
	fmt.Println("In AzureGlags.Resolver.Resolve")

	return nil
}
