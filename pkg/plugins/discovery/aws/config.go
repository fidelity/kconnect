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
	"encoding/base64"
	"fmt"

	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/fidelity/kconnect/pkg/provider"
)

func (p *eksClusterProvider) GetClusterConfig(ctx *provider.Context, cluster *provider.Cluster) (*api.Config, string, error) {
	clusterName := fmt.Sprintf("kconnect-eks-%s", cluster.Name)
	userName := fmt.Sprintf("kconnect-%s-%s", p.identity.ProfileName, cluster.Name)
	contextName := fmt.Sprintf("%s@%s", userName, clusterName)

	certData, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthorityData)
	if err != nil {
		return nil, "", fmt.Errorf("decoding certificate: %w", err)
	}

	cfg := &api.Config{
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   *cluster.ControlPlaneEndpoint,
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
		APIVersion: "client.authentication.k8s.io/v1alpha1",
		Command:    "aws-iam-authenticator",
		Args: []string{
			"token",
			"-i",
			cluster.Name,
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

	return cfg, contextName, nil
}
