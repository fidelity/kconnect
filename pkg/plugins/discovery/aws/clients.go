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

func newSession(region, profile string) (client.ConfigProvider, error) {
	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", profile),
	})
	if err != nil {
		return nil, fmt.Errorf("creating new aws session in region %s with profile %s: %w", region, profile, err)
	}

	return awsSession, nil
}

func newIAMClient(session client.ConfigProvider) iamiface.IAMAPI {
	iamClient := iam.New(session)
	iamClient.Handlers.Build.PushFrontNamed(getUserAgentHandler())

	return iamClient
}

func newEKSClient(session client.ConfigProvider) eksiface.EKSAPI {
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
