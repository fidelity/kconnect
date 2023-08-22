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

package oidc

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/fidelity/kconnect/pkg/oidc"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"k8s.io/client-go/tools/clientcmd/api"
)

const True = "true"

func (p *oidcClusterProvider) GetConfig(ctx context.Context, input *discovery.GetConfigInput) (*discovery.GetConfigOutput, error) {

	clusterName := input.Cluster.Name
	userName := p.identity.OidcId
	contextName := fmt.Sprintf("%s@%s", userName, clusterName)

	certData, err := base64.StdEncoding.DecodeString(*input.Cluster.CertificateAuthorityData)
	if err != nil {
		return nil, fmt.Errorf("decoding certificate: %w", err)
	}

	oidcID, _ := input.Identity.(*oidc.Identity)

	cfg := &api.Config{
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   *input.Cluster.ControlPlaneEndpoint,
				CertificateAuthorityData: certData,
			},
		},
		Contexts: map[string]*api.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		},
	}

	args := []string{
		"oidc-login",
		"get-token",
		"--oidc-issuer-url=oidc-server",
		"--oidc-client-id=" + oidcID.OidcId,
	}

	if oidcID.UsePkce == True {
		args = append(args, "--oidc-use-pkce")
	} else {
		args = append(args, "--oidc-client-secret="+oidcID.OidcSecret)
	}
	args = append(args, "--insecure-skip-tls-verify")

	execConfig := &api.ExecConfig{
		APIVersion:      "client.authentication.k8s.io/v1beta1",
		Command:         "kubectl",
		Args:            args,
		InteractiveMode: api.IfAvailableExecInteractiveMode,
	}

	cfg.AuthInfos = map[string]*api.AuthInfo{
		userName: {
			Exec: execConfig,
		},
	}

	cfg.CurrentContext = contextName

	return &discovery.GetConfigOutput{
		KubeConfig:  cfg,
		ContextName: &contextName,
	}, nil
}
