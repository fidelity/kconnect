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
package aws

import (
	"github.com/fidelity/kconnect/pkg/providers"
	"github.com/spf13/pflag"
)

// EKSClusterProvider will discover EKS clusters in AWS
type EKSClusterProvider struct {
	//Discover(identity identity.Identity, options map[string]string) (map[credentials][]clusters, error)
}

// Name returns the name of the provider
func (p *EKSClusterProvider) Name() string {
	return "eks"
}

// Flags returns the flags for this provider
func (p *EKSClusterProvider) Flags() *pflag.FlagSet {
	return nil
}

// FlagsResolver returns the resolver to use for flags with this provider
func (p *EKSClusterProvider) FlagsResolver() providers.FlagsResolver {
	return &FlagsResolver{}
}

// Usage returns a description for use in the help/usage
func (p *EKSClusterProvider) Usage() string {
	return "discover and connect to AWS EKS clusters"
}
