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
	"github.com/fidelity/kconnect/pkg/provider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func init() {
	if err := provider.RegisterClusterProviderPlugin("eks", newEKSProvider()); err != nil {
		// TODO: handle fatal error
		log.Fatalf("Failed to register EKS cluster provider plugin: %v", err)
	}
}

func newEKSProvider() *eksClusterProvider {
	return &eksClusterProvider{}
}

// EKSClusterProvider will discover EKS clusters in AWS
type eksClusterProvider struct {
	region     *string
	profile    *string
	roleArn    *string
	roleFilter *string
	cluster    *string

	flags *pflag.FlagSet
}

// Name returns the name of the provider
func (p *eksClusterProvider) Name() string {
	return "eks"
}

// Flags returns the flags for this provider
func (p *eksClusterProvider) Flags() *pflag.FlagSet {
	if p.flags == nil {
		p.flags = &pflag.FlagSet{}
		p.region = p.flags.String("region", "us-west-2", "AWS region to connect to. Defaults to us-west-2")
		p.profile = p.flags.String("profile", "", "AWS profile to use")
		p.roleArn = p.flags.String("role-arn", "", "ARN of the AWS role to be assumed")
		p.roleFilter = p.flags.String("role-filter", "*EKS*", "A filter to apply to the roles list")
		p.cluster = p.flags.StringP("cluster", "c", "", "Name of a EKS cluster to use.")
	}

	return p.flags
}

// FlagsResolver returns the resolver to use for flags with this provider
func (p *eksClusterProvider) FlagsResolver() provider.FlagsResolver {
	return &awsFlagsResolver{}
}

// Usage returns a description for use in the help/usage
func (p *eksClusterProvider) Usage() string {
	return "discover and connect to AWS EKS clusters"
}
