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
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/utils"
)

func ResolvePartition(cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(PartitionConfigItem) {
		return nil
	}

	return prompt.Choose(cfg, PartitionConfigItem, "Select the AWS partition", true, awsPartitionOptions)
}

func ResolveRegion(cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(RegionConfigItem) {
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

	regionFilter := ""
	regionFilterCfg := cfg.Get("region-filter")
	if regionFilterCfg != nil {
		regionFilter = regionFilterCfg.Value.(string)
	}

	options := []string{}
	for _, region := range partition.Regions() {
		if regionFilter == "" || strings.Contains(region.ID(), regionFilter) {
			options = append(options, region.ID())
		}
	}
	sort.Slice(options, func(i, j int) bool { return options[i] < options[j] })

	region := ""
	prompt := &survey.Select{
		Message: "Select an AWS region",
		Options: options,
		Filter:  utils.SurveyFilter,
	}
	if err := survey.AskOne(prompt, &region, survey.WithValidator(survey.Required)); err != nil {
		if errors.Is(err, terminal.InterruptErr) {
			zap.S().Info("Received interrupt, exiting..")
			os.Exit(0)
		}
		return fmt.Errorf("asking for region: %w", err)
	}

	if err := cfg.SetValue(RegionConfigItem, region); err != nil {
		return fmt.Errorf("setting region config: %w", err)
	}

	return nil
}

func awsPartitionOptions() (map[string]string, error) {
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()

	options := map[string]string{}
	for _, partition := range partitions {
		options[partition.ID()] = partition.ID()
	}

	return options, nil
}
