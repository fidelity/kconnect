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
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	awssp "github.com/fidelity/kconnect/pkg/plugins/identity/saml/sp/aws"
	"github.com/fidelity/kconnect/pkg/provider"
)

func init() {
	if err := provider.RegisterClusterProviderPlugin("eks", newEKSProvider()); err != nil {
		// TODO: handle fatal error
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
	Profile      *string `json:"profile"`
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
	cs := config.NewConfigurationSet()
	cs.String("partition", endpoints.AwsPartition().ID(), "AWS partition to use")                                                     //nolint: errcheck
	cs.String("region", "", "AWS region to connect to")                                                                               //nolint: errcheck
	cs.String("region-filter", "", "A filter to apply to the AWS regions list, e.g. 'us-' will only show US regions")                 //nolint: errcheck
	cs.String("profile", "", "AWS profile to use")                                                                                    //nolint: errcheck
	cs.String("role-arn", "", "ARN of the AWS role to be assumed")                                                                    //nolint: errcheck
	cs.String("role-filter", "", "A filter to apply to the roles list, e.g. 'EKS' will only show roles that contain EKS in the name") //nolint: errcheck

	cs.SetRequired("profile")   //nolint: errcheck
	cs.SetRequired("region")    //nolint: errcheck
	cs.SetRequired("partition") //nolint: errcheck

	cs.SetHidden("profile") //nolint: errcheck

	return cs
}

// ConfigurationResolver returns the resolver to use for config with this provider
func (p *eksClusterProvider) ConfigurationResolver() provider.ConfigResolver {
	return &awsConfigResolver{}
}

// Usage returns a description for use in the help/usage
func (p *eksClusterProvider) Usage() string {
	return "discover and connect to AWS EKS clusters"
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
