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

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/sirupsen/logrus"

	"github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	awssp "github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp/aws"
	"github.com/fidelity/kconnect/pkg/provider"
)

func init() {
	if err := provider.RegisterClusterProviderPlugin("eks", newEKSProvider()); err != nil {
		// TODO: handle fatal error
		logrus.Fatalf("Failed to register EKS cluster provider plugin: %v", err)
	}
}

func newEKSProvider() *eksClusterProvider {
	logger := logrus.WithField("disovery-provider", "eks")
	return &eksClusterProvider{
		logger: logger,
	}
}

type eksClusteProviderConfig struct {
	provider.ClusterProviderConfig

	Region     *string `flag:"region" json:"region"`
	Profile    *string `flag:"profile" json:"profile"`
	RoleArn    *string `flag:"role-arn" json:"role-arn"`
	RoleFilter *string `flag:"role-filter" json:"role-filter"`
}

// EKSClusterProvider will discover EKS clusters in AWS
type eksClusterProvider struct {
	config    *eksClusteProviderConfig
	identity  *awssp.Identity
	logger    *logrus.Entry
	eksClient eksiface.EKSAPI
}

// Name returns the name of the provider
func (p *eksClusterProvider) Name() string {
	return "eks"
}

// ConfigurationItems returns the configuration items for this provider
func (p *eksClusterProvider) ConfigurationItems() config.ConfigurationSet {
	cs := config.NewConfigurationSet()
	cs.String("partition", endpoints.AwsPartition().ID(), "AWS partition to use") //nolint: errcheck
	cs.String("region", "", "AWS region to connect to")                           //nolint: errcheck
	cs.String("profile", "", "AWS profile to use")                                //nolint: errcheck
	cs.String("role-arn", "", "ARN of the AWS role to be assumed")                //nolint: errcheck
	cs.String("role-filter", "*EKS*", "A filter to apply to the roles list")      //nolint: errcheck

	cs.SetRequired("profile")   //nolint: errcheck
	cs.SetRequired("region")    //nolint: errcheck
	cs.SetRequired("partition") //nolint: errcheck

	return cs
}

// ConfigurationResolver returns the resolver to use for config with this provider
func (p *eksClusterProvider) ConfigurationResolver() provider.ConfigResolver {
	return &awsConfigResolver{
		logger: p.logger,
	}
}

// Usage returns a description for use in the help/usage
func (p *eksClusterProvider) Usage() string {
	return "discover and connect to AWS EKS clusters"
}

func (p *eksClusterProvider) setup(ctx *provider.Context, identity provider.Identity, logger *logrus.Entry) error {
	cfg := &eksClusteProviderConfig{}
	if err := config.Unmarshall(ctx.ConfigurationItems(), cfg); err != nil {
		return fmt.Errorf("unmarshalling config items into eksClusteProviderConfig: %w", err)
	}
	p.config = cfg

	awsID, ok := identity.(*awssp.Identity)
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
