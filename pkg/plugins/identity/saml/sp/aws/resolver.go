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
	"sort"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws/endpoints"

	"github.com/fidelity/kconnect/pkg/config"
)

// ResolveConfiguration will resolve the values for the AWS specific config items that have no value.
// It will query AWS and interactively ask the user for selections.
func (p *ServiceProvider) ResolveConfiguration(cfg config.ConfigurationSet) error {
	p.logger.Debug("resolving AWS identity configuration items")

	// NOTE: resolution is only needed for required fields
	if err := p.resolveIdpProvider("idp-provider", cfg); err != nil {
		return fmt.Errorf("resolving idp-provider: %w", err)
	}
	if err := p.resolveIdpEndpoint("idp-endpoint", cfg); err != nil {
		return fmt.Errorf("resolving idp-endpoint: %w", err)
	}
	if err := p.resolveProfile("profile", cfg); err != nil {
		return fmt.Errorf("resolving profile: %w", err)
	}
	if err := p.resolvePartition("partition", cfg); err != nil {
		return fmt.Errorf("resolving partition: %w", err)
	}
	if err := p.resolveRegion("region", cfg); err != nil {
		return fmt.Errorf("resolving region: %w", err)
	}
	if err := p.resolveUsername("username", cfg); err != nil {
		return fmt.Errorf("resolving username: %w", err)
	}
	if err := p.resolvePassword("password", cfg); err != nil {
		return fmt.Errorf("resolving password: %w", err)
	}

	return nil
}

func (p *ServiceProvider) resolveProfile(name string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	profile := ""
	prompt := &survey.Input{
		Message: "Enter the name of AWS profile",
	}
	if err := survey.AskOne(prompt, &profile, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for profile name: %w", err)
	}

	if err := cfg.SetValue(name, profile); err != nil {
		p.logger.Errorf("failed setting profile config to %s: %s", profile, err.Error())
		return fmt.Errorf("setting profile config: %w", err)
	}

	return nil
}

func (p *ServiceProvider) resolvePartition(name string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()

	options := []string{}
	for _, partition := range partitions {
		options = append(options, partition.ID())
	}
	sort.Slice(options, func(i, j int) bool { return options[i] < options[j] })

	partitionID := ""
	prompt := &survey.Select{
		Message: "Select the AWS partition",
		Options: options,
	}
	if err := survey.AskOne(prompt, &partitionID, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for partition: %w", err)
	}

	if err := cfg.SetValue(name, partitionID); err != nil {
		p.logger.Errorf("failed setting partition config to %s: %s", partitionID, err.Error())
		return fmt.Errorf("setting partition config: %w", err)
	}

	return nil
}

func (p *ServiceProvider) resolveRegion(name string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	partitionCfg := cfg.Get("partition")
	if partitionCfg == nil {
		return ErrNoPartitionSupplied
	}
	partitionID := partitionCfg.Value.(string)

	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()

	var partition endpoints.Partition
	for _, p := range partitions {
		if p.ID() == partitionID {
			partition = p
			break
		}
	}
	if partition.ID() == "" {
		return fmt.Errorf("finding partition with id %s: %w", partitionID, ErrPartitionNotFound)
	}

	options := []string{}
	for _, region := range partition.Regions() {
		options = append(options, region.ID())
	}
	sort.Slice(options, func(i, j int) bool { return options[i] < options[j] })

	region := ""
	prompt := &survey.Select{
		Message: "Select a AWS region",
		Options: options,
	}
	if err := survey.AskOne(prompt, &region, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for region: %w", err)
	}

	if err := cfg.SetValue(name, region); err != nil {
		p.logger.Errorf("failed setting region config to %s: %s", region, err.Error())
		return fmt.Errorf("setting region config: %w", err)
	}

	return nil
}

func (p *ServiceProvider) resolveUsername(name string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	username := ""
	prompt := &survey.Input{
		Message: "Enter your username",
	}
	if err := survey.AskOne(prompt, &username, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for username name: %w", err)
	}

	if err := cfg.SetValue(name, username); err != nil {
		p.logger.Errorf("failed setting username config to %s: %s", username, err.Error())
		return fmt.Errorf("setting username config: %w", err)
	}

	return nil
}

func (p *ServiceProvider) resolvePassword(name string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	password := ""
	prompt := &survey.Password{
		Message: "Enter your password",
	}
	if err := survey.AskOne(prompt, &password, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for password name: %w", err)
	}

	if err := cfg.SetValue(name, password); err != nil {
		p.logger.Errorf("failed setting password config to %s: %s", password, err.Error())
		return fmt.Errorf("setting password config: %w", err)
	}

	return nil
}

func (p *ServiceProvider) resolveIdpEndpoint(name string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}

	endpoint := ""
	prompt := &survey.Input{
		Message: "Enter the endpoint for the IdP",
	}
	if err := survey.AskOne(prompt, &endpoint, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for idp-endpoint name: %w", err)
	}

	if err := cfg.SetValue(name, endpoint); err != nil {
		p.logger.Errorf("failed setting idp-endpoint config to %s: %s", endpoint, err.Error())
		return fmt.Errorf("setting idp-endpoint config: %w", err)
	}

	return nil
}

func (p *ServiceProvider) resolveIdpProvider(name string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(name) {
		return nil
	}
	//TODO: get this from saml2aws????
	options := []string{"Akamai", "AzureAD", "ADFS", "ADFS2", "GoogleApps", "Ping", "PingNTLM", "JumpCloud", "Okta", "OneLogin", "PSU", "KeyCloak", "F5APM", "Shibboleth", "ShibbolethECP", "NetIQ"}

	idpProvider := ""
	prompt := &survey.Select{
		Message: "Select your identity provider",
		Options: options,
	}
	if err := survey.AskOne(prompt, &idpProvider, survey.WithValidator(survey.Required)); err != nil {
		return fmt.Errorf("asking for idp-provider: %w", err)
	}

	if err := cfg.SetValue(name, idpProvider); err != nil {
		p.logger.Errorf("failed setting idp-provider config to %s: %s", idpProvider, err.Error())
		return fmt.Errorf("setting idp-provider config: %w", err)
	}

	return nil
}
