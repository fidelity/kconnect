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
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	awssp "github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp/aws"
	"github.com/fidelity/kconnect/pkg/provider"
)

func init() {
	if err := provider.RegisterClusterProviderPlugin("eks", newEKSProvider()); err != nil {
		zap.S().Fatalw("Failed to register EKS cluster provider plugin", "error", err)
	}
}

func newEKSProvider() *eksClusterProvider {
	return &eksClusterProvider{}
}

type eksClusteProviderConfig struct {
	provider.ClusterProviderConfig
	Region       *string `json:"region"`
	RegionFilter *string `json:"region-filter"`
	RoleArn      *string `json:"role-arn"`
	RoleFilter   *string `json:"role-filter"`
}

// EKSClusterProvider will discover EKS clusters in AWS
type eksClusterProvider struct {
	config    *eksClusteProviderConfig
	identity  *awssp.Identity
	eksClient eksiface.EKSAPI

	logger *zap.SugaredLogger
}

// Name returns the name of the provider
func (p *eksClusterProvider) Name() string {
	return "eks"
}

// ConfigurationItems returns the configuration items for this provider
func (p *eksClusterProvider) ConfigurationItems() config.ConfigurationSet {
	cs := aws.SharedConfig()

	cs.String("region-filter", "", "A filter to apply to the AWS regions list, e.g. 'us-' will only show US regions")                 //nolint: errcheck
	cs.String("role-arn", "", "ARN of the AWS role to be assumed")                                                                    //nolint: errcheck
	cs.String("role-filter", "", "A filter to apply to the roles list, e.g. 'EKS' will only show roles that contain EKS in the name") //nolint: errcheck

	return cs
}

// ConfigurationResolver returns the resolver to use for config with this provider
func (p *eksClusterProvider) ConfigurationResolver() provider.ConfigResolver {
	return &awsConfigResolver{}
}

func (p *eksClusterProvider) setup(ctx *provider.Context, identity provider.Identity) error {
	p.ensureLogger()
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

	p.logger.Debugw("creating AWS session", "region", *p.config.Region, "profile", awsID.ProfileName)
	session, err := aws.NewSession(*p.config.Region, awsID.ProfileName)
	if err != nil {
		return fmt.Errorf("getting aws session: %w", err)
	}
	p.eksClient = aws.NewEKSClient(session)

	return nil
}

func (p *eksClusterProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", "eks")
	}
}

// UsageExample will provide an example of the usage of this provider
func (p *eksClusterProvider) UsageExample() string {
	return `
  # Discover EKS clusters using SAML
  kconnect use eks --idp-protocol saml

  # Discover EKS clusters using SAML with a specific role
  kconnect use eks --idp-protocol saml --role-arn arn:aws:iam::000000000000:role/KubernetesAdmin

  # Discover an EKS cluster and add an alias to its connection history entry
  kconnect use eks --alias mycluster
`
}
