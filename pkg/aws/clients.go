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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/fidelity/kconnect/internal/version"
)

func NewSession(region, profile, accessKey, secretKey, sessionToken, awsSharedCredentialsFile string) (*aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	opts = append(opts, config.WithRegion(region))

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
		if awsSharedCredentialsFile != "" {
			opts = append(opts, config.WithSharedCredentialsFiles([]string{awsSharedCredentialsFile}))
		}
	} else if accessKey != "" && secretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return nil, fmt.Errorf("creating new aws session in region %s using creds: %w", region, err)
	}

	return &awsCfg, nil
}

func NewIAMClient(cfg aws.Config) *iam.Client {
	iamClient := iam.NewFromConfig(cfg, func(o *iam.Options) {
		o.APIOptions = append(o.APIOptions, middleware.AddUserAgentKeyValue("kconnect.fidelity.github.com", version.Get().String()))
	})

	return iamClient
}

func NewEKSClient(cfg aws.Config) *eks.Client {
	eksClient := eks.NewFromConfig(cfg, func(o *eks.Options) {
		o.APIOptions = append(o.APIOptions, middleware.AddUserAgentKeyValue("kconnect.fidelity.github.com", version.Get().String()))
	})

	return eksClient
}
