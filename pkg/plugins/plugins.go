package plugins

import (
	// Initialize all the identity and discovery plugins
	_ "github.com/fidelity/kconnect/pkg/plugins/discovery"
	_ "github.com/fidelity/kconnect/pkg/plugins/identity"
)
