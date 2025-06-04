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
	"os"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/common"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
	"github.com/fidelity/kconnect/pkg/utils"
)

const (
	ProviderName = "eks"
	UsageExample = `
  # Discover EKS clusters using SAML
  {{.CommandPath}} use eks --idp-protocol saml

  # Discover EKS clusters using SAML with a specific role
  {{.CommandPath}} use eks --idp-protocol saml --role-arn arn:aws:iam::000000000000:role/KubernetesAdmin

  # Discover an EKS cluster and add an alias to its connection history entry
  {{.CommandPath}} use eks --alias mycluster
  `
)

func init() {
	if err := registry.RegisterDiscoveryPlugin(&registry.DiscoveryPluginRegistration{
		PluginRegistration: registry.PluginRegistration{
			Name:                   ProviderName,
			UsageExample:           UsageExample,
			ConfigurationItemsFunc: ConfigurationItems,
		},
		CreateFunc:                 New,
		SupportedIdentityProviders: []string{"aws-iam", "saml"},
	}); err != nil {
		zap.S().Fatalw("Failed to register EKS discovery plugin", "error", err)
	}
}

// New will create a new EKS discovery plugin
func New(input *provider.PluginCreationInput) (discovery.Provider, error) {
	return &eksClusterProvider{
		logger:      input.Logger,
		interactive: input.IsInteractive,
	}, nil
}

type eksClusterProviderConfig struct {
	common.ClusterProviderConfig
	AssumeRoleARN *string `json:"assume-role-arn"`
	Region        *string `json:"region"`
	RegionFilter  *string `json:"region-filter"`
	RoleArn       *string `json:"role-arn"`
	RoleFilter    *string `json:"role-filter"`
}

// EKSClusterProvider will discover EKS clusters in AWS
type eksClusterProvider struct {
	config    *eksClusterProviderConfig
	identity  *aws.Identity
	eksClient *eks.Client

	interactive bool
	logger      *zap.SugaredLogger
}

// Name returns the name of the provider
func (p *eksClusterProvider) Name() string {
	return ProviderName
}

func (p *eksClusterProvider) setup(cs config.ConfigurationSet, userID identity.Identity) error {
	cfg := &eksClusterProviderConfig{}
	if err := config.Unmarshall(cs, cfg); err != nil {
		return fmt.Errorf("unmarshalling config items into eksClusterProviderConfig: %w", err)
	}
	p.config = cfg

	awsID, ok := userID.(*aws.Identity)
	if !ok {
		return ErrNotAWSIdentity
	}
	p.identity = awsID

	p.logger.Debugw("creating AWS session", "region", *p.config.Region)
	sess, err := aws.NewSession(p.identity.Region, p.identity.ProfileName, p.identity.AWSAccessKey, p.identity.AWSSecretKey, p.identity.AWSSessionToken, p.identity.AWSSharedCredentialsFile)
	if err != nil {
		return fmt.Errorf("creating aws session: %w", err)
	}
	p.eksClient = aws.NewEKSClient(*sess)

	return nil
}

func (p *eksClusterProvider) ListPreReqs() []*provider.PreReq {
	return []*provider.PreReq{}
}

func (p *eksClusterProvider) CheckPreReqs() error {
	return utils.CheckAWSIAMAuthPrereq()
}

// ConfigurationItems returns the configuration items for this provider
func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	cs := aws.SharedConfig()

	cs.String("aws-shared-credentials-file", os.Getenv("AWS_SHARED_CREDENTIALS_FILE"), "Location to store AWS credentials file")
	cs.String("assume-role-arn", "", "ARN of the AWS role to be assumed")
	cs.String("region-filter", "", "A regex filter to apply to the AWS regions list, e.g. '^us-|^eu-' will only show US and eu regions") //nolint: errcheck
	cs.String("role-arn", "", "ARN of the AWS role to be logged in with")                                                                //nolint: errcheck
	cs.String("role-filter", "", "A filter to apply to the roles list, e.g. 'EKS' will only show roles that contain EKS in the name")    //nolint: errcheck

	return cs, nil
}
