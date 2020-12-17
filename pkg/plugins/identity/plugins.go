/*
Copyright 2020 The kconnect Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package discovery

import (
	// Initialize the identity plugins
	_ "github.com/fidelity/kconnect/pkg/plugins/identity/aws/iam"
	_ "github.com/fidelity/kconnect/pkg/plugins/identity/azure/aad"
	_ "github.com/fidelity/kconnect/pkg/plugins/identity/azure/env"
	_ "github.com/fidelity/kconnect/pkg/plugins/identity/rancher/activedirectory"
	_ "github.com/fidelity/kconnect/pkg/plugins/identity/saml"
	_ "github.com/fidelity/kconnect/pkg/plugins/identity/static/token"
)
