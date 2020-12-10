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
	"errors"
	"fmt"

	"go.uber.org/zap"

	kaws "github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/provider"
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
	if err := provider.RegisterIdentityProviderPlugin(ProviderName, newProvider()); err != nil {
		zap.S().Fatalw("failed to register aws iam identity provider plugin", "error", err)
	}
}

func newProvider() *iamIdentityProvider {
	return &iamIdentityProvider{}
}

type iamIdentityProvider struct {
	logger *zap.SugaredLogger
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

// Authenticate will authenticate a user and return details of
// their identity.
func (p *iamIdentityProvider) Authenticate(ctx *provider.Context, clusterProvider string) (provider.Identity, error) {
	p.ensureLogger()
	p.logger.Info("using aws iam for authentication")

	cfg := &providerConfig{}
	if err := config.Unmarshall(ctx.ConfigurationItems(), cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into providerConfig: %w", err)
	}

	if err := p.validateConfig(cfg); err != nil {
		return nil, err
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

	return &kaws.Identity{
		ProfileName:     cfg.Profile,
		AWSAccessKey:    creds.AccessKeyID,
		AWSSecretKey:    creds.SecretAccessKey,
		AWSSessionToken: creds.SessionToken,
		Region:          cfg.Region,
	}, nil
}

// ConfigurationItems will return the configuration items for the intentity plugin based
// of the cluster provider that its being used in conjunction with
func (p *iamIdentityProvider) ConfigurationItems(clusterProviderName string) (config.ConfigurationSet, error) {
	p.ensureLogger()
	cs := config.NewConfigurationSet()

	kaws.AddRegionConfig(cs)
	kaws.AddPartitionConfig(cs)
	kaws.AddIAMConfigs(cs)

	return cs, nil
}

// Usage returns a string to display for help
func (p *iamIdentityProvider) Usage(clusterProvider string) (string, error) {
	return "", nil
}

func (p *iamIdentityProvider) ensureLogger() {
	if p.logger == nil {
		p.logger = zap.S().With("provider", ProviderName)
	}
}

func (p *iamIdentityProvider) validateConfig(cfg *providerConfig) error {
	if cfg.Profile != "" && cfg.AccessKey != "" {
		return ErrProfileWithAccessKey
	}
	if cfg.Profile != "" && cfg.SecretKey != "" {
		return ErrProfileWithSecretKey
	}
	if cfg.AccessKey != "" && cfg.SecretKey == "" {
		return ErrAccessAndSecretRequired
	}
	if cfg.AccessKey == "" && cfg.SecretKey != "" {
		return ErrAccessAndSecretRequired
	}

	return nil
}
