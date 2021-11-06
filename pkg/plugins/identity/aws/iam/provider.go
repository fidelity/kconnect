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

package iam

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	kaws "github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
)

const (
	ProviderName = "aws-iam"
)

var (
	ErrProfileWithAccessKey    = errors.New("cannot use profile with access-key")
	ErrProfileWithSecretKey    = errors.New("cannot use profile with secret-key")
	ErrAccessAndSecretRequired = errors.New("access-key and secret-key are both required")
)

func init() {
	if err := registry.RegisterIdentityPlugin(&registry.IdentityPluginRegistration{
		PluginRegistration: registry.PluginRegistration{
			Name:                   ProviderName,
			UsageExample:           "",
			ConfigurationItemsFunc: ConfigurationItems,
		},
		CreateFunc: New,
	}); err != nil {
		zap.S().Fatalw("Failed to register AWS IAM identity plugin", "error", err)
	}
}

// New will create a new AWS IAM identity provider
func New(input *provider.PluginCreationInput) (identity.Provider, error) {
	return &iamIdentityProvider{
		logger:      input.Logger,
		interactive: input.IsInteractice,
	}, nil
}

type iamIdentityProvider struct {
	logger      *zap.SugaredLogger
	interactive bool
}

type providerConfig struct {
	Profile      string `json:"profile"`
	AccessKey    string `json:"access-key"`
	SecretKey    string `json:"secret-key"`
	SessionToken string `json:"session-token"`
	Region       string `json:"region"`
	Partition    string `json:"partition"`
}

func (p *iamIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will authenticate a user and return details of their identity.
func (p *iamIdentityProvider) Authenticate(ctx context.Context, input *identity.AuthenticateInput) (*identity.AuthenticateOutput, error) {
	p.logger.Info("using aws iam for authentication")

	cfg := &providerConfig{}
	if err := config.Unmarshall(input.ConfigSet, cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into providerConfig: %w", err)
	}

	sess, err := kaws.NewSession(cfg.Region, cfg.Profile, cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("creating aws session: %w", err)
	}

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, fmt.Errorf("getting credentials: %w", err)
	}
	p.logger.Debugw("found aws iam credentials", "provider", creds.ProviderName)

	id := &kaws.Identity{
		ProfileName:     cfg.Profile,
		AWSAccessKey:    creds.AccessKeyID,
		AWSSecretKey:    creds.SecretAccessKey,
		AWSSessionToken: creds.SessionToken,
		Region:          cfg.Region,
	}

	return &identity.AuthenticateOutput{
		Identity: id,
	}, nil
}

// Validate is used to validate the config items and return any errors
func (p *iamIdentityProvider) Validate(cfg config.ConfigurationSet) error {
	hasProfile := cfg.ExistsWithValue(kaws.ProfileConfigItem)
	hasAccessKey := cfg.ExistsWithValue(kaws.AccessKeyConfigItem)
	hasSecretKey := cfg.ExistsWithValue(kaws.SecretKeyConfigItem)

	if hasProfile && hasAccessKey {
		return ErrProfileWithAccessKey
	}
	if hasProfile && hasSecretKey {
		return ErrProfileWithSecretKey
	}
	if hasAccessKey && !hasSecretKey {
		return ErrAccessAndSecretRequired
	}
	if !hasAccessKey && hasSecretKey {
		return ErrAccessAndSecretRequired
	}

	return nil
}

// Resolve will resolve the values for the supplied config items. It will interactively
// resolve the values by asking the user for selections.
func (p *iamIdentityProvider) Resolve(config config.ConfigurationSet, identity provider.Identity) error {
	return nil
}

func (p *iamIdentityProvider) ListPreReqs() []*provider.PreReq {
	return []*provider.PreReq{}
}

func (p *iamIdentityProvider) CheckPreReqs() error {
	return nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()

	kaws.AddRegionConfig(cs)
	kaws.AddPartitionConfig(cs)
	kaws.AddIAMConfigs(cs)

	return cs, nil
}
