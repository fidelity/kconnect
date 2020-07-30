package discovery

import (
	// Initialize the identity plugins
	_ "github.com/fidelity/kconnect/pkg/plugins/identity/empty"
	_ "github.com/fidelity/kconnect/pkg/plugins/identity/saml"
)
