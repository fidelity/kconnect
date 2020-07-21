package aws

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"

	"github.com/fidelity/kconnect/internal/version"
)

func newIAMClient(session client.ConfigProvider) iamiface.IAMAPI {
	iamClient := iam.New(session)
	iamClient.Handlers.Build.PushFrontNamed(getUserAgentHandler())

	return iamClient
}

func getUserAgentHandler() request.NamedHandler {
	return request.NamedHandler{
		Name: "kconnect/user-agent",
		Fn:   request.MakeAddToUserAgentHandler("kconnect.fidelity.github.com", version.Get().String()),
	}
}
