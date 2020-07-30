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
	"fmt"

	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/plugins/identity/saml"
	"github.com/fidelity/kconnect/pkg/provider"
)

func init() {
	if err := provider.RegisterClusterProviderPlugin("eks", newEKSProvider()); err != nil {
		// TODO: handle fatal error
		logrus.Fatalf("Failed to register EKS cluster provider plugin: %v", err)
	}
}

func newEKSProvider() *eksClusterProvider {
	return &eksClusterProvider{}
}

type eksClusteProviderConfig struct {
	provider.ClusterProviderConfig

	Region     *string `flag:"region"`
	Profile    *string `flag:"profile"`
	RoleArn    *string `flag:"role-arn"`
	RoleFilter *string `flag:"role-filter"`
}

// EKSClusterProvider will discover EKS clusters in AWS
type eksClusterProvider struct {
	config    *eksClusteProviderConfig
	identity  *saml.AWSIdentity
	logger    *logrus.Entry
	eksClient eksiface.EKSAPI
}

// Name returns the name of the provider
func (p *eksClusterProvider) Name() string {
	return "eks"
}

// Flags returns the flags for this provider
func (p *eksClusterProvider) Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.String("region", "us-west-2", "AWS region to connect to. Defaults to us-west-2")
	fs.String("profile", "", "AWS profile to use")
	fs.String("role-arn", "", "ARN of the AWS role to be assumed")
	fs.String("role-filter", "*EKS*", "A filter to apply to the roles list")

	return fs
}

// FlagsResolver returns the resolver to use for flags with this provider
func (p *eksClusterProvider) FlagsResolver() provider.FlagsResolver {
	return &awsFlagsResolver{}
}

// Usage returns a description for use in the help/usage
func (p *eksClusterProvider) Usage() string {
	return "discover and connect to AWS EKS clusters"
}

func (p *eksClusterProvider) setup(ctx *provider.Context, identity provider.Identity, logger *logrus.Entry) error {
	cfg := &eksClusteProviderConfig{}
	if err := flags.Unmarshal(ctx.Command().Flags(), cfg); err != nil {
		return fmt.Errorf("unmarshalling flags into config: %w", err)
	}
	p.config = cfg

	awsID, ok := identity.(*saml.AWSIdentity)
	if !ok {
		return ErrNotAWSIdentity
	}
	p.identity = awsID

	logger.Debugf("creating AWS session with region %s and profile %s", *p.config.Region, awsID.ProfileName)
	session, err := aws.NewSession(*p.config.Region, awsID.ProfileName)
	if err != nil {
		return fmt.Errorf("getting aws session: %w", err)
	}
	p.eksClient = aws.NewEKSClient(session)

	p.logger = logger
	return nil
}
