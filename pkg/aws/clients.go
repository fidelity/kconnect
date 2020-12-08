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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/fidelity/kconnect/internal/version"
)

func NewSession(region, profile, accessKey, secretKey, sessionToken string) (*session.Session, error) {
	cfg := aws.Config{
		Region: aws.String(region),
	}

	if profile != "" {
		cfg.Credentials = credentials.NewSharedCredentials("", profile)
	} else if accessKey != "" && secretKey != "" {
		cfg.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, sessionToken)
	}

	options := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            cfg,
	}

	awsSession, err := session.NewSessionWithOptions(options)
	if err != nil {
		return nil, fmt.Errorf("creating new aws session in region %s using creds: %w", region, err)
	}

	return awsSession, nil
}

func NewIAMClient(session client.ConfigProvider) iamiface.IAMAPI {
	iamClient := iam.New(session)
	iamClient.Handlers.Build.PushFrontNamed(getUserAgentHandler())

	return iamClient
}

func NewEKSClient(session client.ConfigProvider) eksiface.EKSAPI {
	eksClient := eks.New(session)
	eksClient.Handlers.Build.PushFrontNamed(getUserAgentHandler())

	return eksClient
}

func getUserAgentHandler() request.NamedHandler {
	return request.NamedHandler{
		Name: "kconnect/user-agent",
		Fn:   request.MakeAddToUserAgentHandler("kconnect.fidelity.github.com", version.Get().String()),
	}
}
