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
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/fidelity/kconnect/pkg/config"
)

const (
	RegionConfigItem       = "region"
	PartitionConfigItem    = "partition"
	ProfileConfigItem      = "profile"
	AccessKeyConfigItem    = "access-key"
	SecretKeyConfigItem    = "secret-key"
	SessionTokenConfigItem = "session-token"

	RegionPrompt  = "AWS Region:"
	ProfilePrompt = "AWS Profile:"
)

// SharedConfig will return shared configuration items for AWS based cluster and identity providers
func SharedConfig() config.ConfigurationSet {
	cs := config.NewConfigurationSet()
	AddPartitionConfig(cs)
	AddRegionConfig(cs)
	cs.String("static-profile", "", "AWS profile to use. Only for advanced use cases") //nolint: errcheck
	cs.SetHidden("static-profile")                                                     //nolint: errcheck

	return cs
}

func AddRegionConfig(cs config.ConfigurationSet) {
	cs.String(RegionConfigItem, "", "AWS region to connect to") //nolint: errcheck
	cs.SetRequiredWithPrompt(RegionConfigItem, RegionPrompt)    //nolint: errcheck
	cs.SetPriority(RegionConfigItem, 10)                        //nolint: errcheck
	cs.SetResolver(RegionConfigItem, ResolveRegion)

	// region := &config.Item{
	// 	Name: RegionConfigItem,
	// 	Shorthand: "",
	// 	Description: "AWS region to connect to",
	// 	Priority: 10,
	// 	Type: config.ItemTypeString,
	// 	Required: true,
	// 	ResolutionPrompt: RegionPrompt,
	// }
	// cs.Add(region)
	// cs.AddWithResolver(region, ResolveRegion)
}

func AddPartitionConfig(cs config.ConfigurationSet) {
	cs.String(PartitionConfigItem, endpoints.AwsPartition().ID(), "AWS partition to use") //nolint: errcheck
	cs.SetRequiredWithPrompt(ProfileConfigItem, ProfilePrompt)
	cs.SetResolver(PartitionConfigItem, ResolvePartition) //nolint: errcheck
}

func AddIAMConfigs(cs config.ConfigurationSet) {
	cs.String(ProfileConfigItem, "", "AWS profile to use")            //nolint: errcheck
	cs.String(AccessKeyConfigItem, "", "AWS access key to use")       //nolint: errcheck
	cs.String(SecretKeyConfigItem, "", "AWS secret key to use")       //nolint: errcheck
	cs.String(SessionTokenConfigItem, "", "AWS session token to use") //nolint: errcheck
}
