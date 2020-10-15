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

// SharedConfig will return shared configuration items for AWS based cluster and identity providers
func SharedConfig() config.ConfigurationSet {
	cs := config.NewConfigurationSet()
	cs.String("partition", endpoints.AwsPartition().ID(), "AWS partition to use")      //nolint: errcheck
	cs.String("region", "", "AWS region to connect to")                                //nolint: errcheck
	cs.String("static-profile", "", "AWS profile to use. Only for advanced use cases") //nolint: errcheck

	cs.SetRequired("region")    //nolint: errcheck
	cs.SetRequired("partition") //nolint: errcheck

	cs.SetHidden("static-profile") //nolint: errcheck

	return cs
}
