package plugins

import (
	// Initialize all the identity, discovery and resolver plugins
	_ "github.com/fidelity/kconnect/pkg/plugins/discovery"
	_ "github.com/fidelity/kconnect/pkg/plugins/identity"
)
