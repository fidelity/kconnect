package config

import (
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider/identity"
)

// Resolver is an interface that indicates that the plugin can resolve and validate configuration
type Resolver interface {

	// Validate is used to validate the config items and return any errors
	Validate(config config.ConfigurationSet) error

	// Resolve will resolve the values for the supplied config items. It will interactively
	// resolve the values by asking the user for selections.
	Resolve(config config.ConfigurationSet, identity identity.Identity) error
}
