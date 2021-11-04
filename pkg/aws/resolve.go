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

	"github.com/aws/aws-sdk-go/aws/endpoints"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/utils"
)

func ResolvePartition(cfg config.ConfigurationSet) error {
	return prompt.ChooseAndSet(cfg, PartitionConfigItem, "Select the AWS partition", true, awsPartitionOptions)
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
	// for _, region := range partition.Regions() {
	// 	if regionFilter == "" || strings.Contains(region.ID(), regionFilter) {
	// 		options = append(options, region.ID())
	// 	}
	// }
	for _, region := range partition.Regions() {
		options = append(options, region.ID())
	}
	options, err := utils.RegexFilter(options, regionFilter)
	if err != nil {
		return fmt.Errorf("applying region regex %s : %w", regionFilter, err)
	}

	sort.Slice(options, func(i, j int) bool { return options[i] < options[j] })

	err = prompt.ChooseAndSet(cfg, RegionConfigItem, "Select an AWS region", true, prompt.OptionsFromStringSlice(options))
	if err != nil {
		return fmt.Errorf("choosing and setting %s: %w", RegionConfigItem, err)
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
