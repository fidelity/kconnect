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
	"encoding/base64"
	"fmt"

	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/fidelity/kconnect/pkg/provider/discovery"
)

func (p *eksClusterProvider) GetConfig(ctx context.Context, input *discovery.GetConfigInput) (*discovery.GetConfigOutput, error) {
	clusterName := fmt.Sprintf("eks-%s", input.Cluster.Name)
	userName := p.identity.ProfileName
	contextName := fmt.Sprintf("%s@%s", userName, clusterName)

	certData, err := base64.StdEncoding.DecodeString(*input.Cluster.CertificateAuthorityData)
	if err != nil {
		return nil, fmt.Errorf("decoding certificate: %w", err)
	}

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

	execConfig := &api.ExecConfig{
		APIVersion: "client.authentication.k8s.io/v1beta1",
		Command:    "aws-iam-authenticator",
		Args: []string{
			"token",
			"-i",
			input.Cluster.Name,
		},
		Env: []api.ExecEnvVar{
			{
				Name:  "AWS_PROFILE",
				Value: p.identity.ProfileName,
			},
		},
	}

	cfg.AuthInfos = map[string]*api.AuthInfo{
		userName: {
			Exec: execConfig,
		},
	}

	cfg.CurrentContext = contextName

	if input.Namespace != nil && *input.Namespace != "" {
		p.logger.Debugw("setting kubernetes namespace", "namespace", *input.Namespace)
		cfg.Contexts[contextName].Namespace = *input.Namespace
	}

	return &discovery.GetConfigOutput{
		KubeConfig:  cfg,
		ContextName: &contextName,
	}, nil
}
