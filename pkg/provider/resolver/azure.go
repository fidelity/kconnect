package resolver

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/manifoldco/promptui"
	"github.com/spf13/pflag"
)

type AzureFlagsResolver struct {
}

func (r *AzureFlagsResolver) Resolve(identity identity.Identity, flags *pflag.FlagSet) error {
	fmt.Println("In AzureFlagsResolver.Resolve\n")

	//TODO get client from identity

	flag := flags.Lookup("resource-group")
	if flag == nil {
		return fmt.Errorf("no resource-group flag defined")
	}
	if flag.Value == nil {
		if err := r.resolveResourceGroup(flag, flags); err != nil {
			return fmt.Errorf("failed to resolve resource group: %w", err)
		}
	}

	return nil
}

func (r *AzureFlagsResolver) resolveResourceGroup(flag *pflag.Flag, flags *pflag.FlagSet) error {
	//TODO: azure client will be accessible
	//TODO: query the azure API

	// NOTE: testing creating a very long list
	resourceGroups := []string{}
	for i := 0; i < 100; i++ {
		resourceGroupName := fmt.Sprintf("reg-test-%d", i)
		resourceGroups = append(resourceGroups, resourceGroupName)
	}

	prompt := promptui.Select{
		Label: "Resource group",
		Items: resourceGroups,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("failed prompting for resource group: %w", err)
	}
	value := stringValue(result)

	flag.Value = &value

	return nil
}
