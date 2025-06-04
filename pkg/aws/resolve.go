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
	"slices"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/utils"
)

var awsPartitions = []string{"aws", "aws-cn", "aws-us-gov"}
var awsGlobalRegions = []string{"us-east-2", "us-east-1", "us-west-1", "us-west-2", "af-south-1", "ap-east-1", "ap-south-2", "ap-southeast-3", "ap-southeast-5", "ap-southeast-4", "ap-south-1", "ap-northeast-3", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2", "ap-southeast-7", "ap-northeast-1", "ca-central-1", "ca-west-1", "eu-central-1", "eu-west-1", "eu-west-2", "eu-south-1", "eu-west-3", "eu-south-2", "eu-north-1", "eu-central-2", "il-central-1", "mx-central-1", "me-south-1", "me-central-1", "sa-east-1"}
var awsGovCloudRegions = []string{"us-gov-west-1", "us-gov-east-1"}
var awsChinaRegions = []string{"cn-north-1", "cn-northwest-1"}

func ResolvePartition(cfg config.ConfigurationSet) error {
	return prompt.ChooseAndSet(cfg, PartitionConfigItem, "Select the AWS partition", true, awsPartitionOptions)
}

func ResolvePartitionRegions(partitionID string) []string {
	var regions = []string{}
	switch partitionID {
	case "aws-cn":
		regions = awsChinaRegions
	case "aws-us-gov":
		regions = awsGovCloudRegions
	default:
		regions = awsGlobalRegions
	}
	return regions
}

func ResolveRegion(cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue(RegionConfigItem) {
		return nil
	}

	partitionCfg := cfg.Get(PartitionConfigItem)
	if partitionCfg == nil {
		return ErrNoPartitionSupplied
	}
	partitionID := partitionCfg.Value.(string)

	regionFilter := ""
	regionFilterCfg := cfg.Get("region-filter")
	if regionFilterCfg != nil {
		regionFilter = regionFilterCfg.Value.(string)
	}

	options := []string{}
	options = append(options, ResolvePartitionRegions(partitionID)...)
	options, err := utils.RegexFilter(options, regionFilter)
	if err != nil {
		return fmt.Errorf("applying region regex %s : %w", regionFilter, err)
	}
	slices.Sort(options)

	err = prompt.ChooseAndSet(cfg, RegionConfigItem, "Select an AWS region", true, prompt.OptionsFromStringSlice(options))
	if err != nil {
		return fmt.Errorf("choosing and setting %s: %w", RegionConfigItem, err)
	}

	return nil
}

func awsPartitionOptions() (map[string]string, error) {
	options := map[string]string{}
	for _, partition := range awsPartitions {
		options[partition] = partition
	}
	return options, nil
}
