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

package rancher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/fidelity/kconnect/pkg/defaults"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
	"github.com/fidelity/kconnect/pkg/rancher"
)

const (
	generateKubeconfigActionName = "generateKubeconfig"
)

func (p *rancherClusterProvider) GetConfig(ctx context.Context, input *discovery.GetConfigInput) (*discovery.GetConfigOutput, error) {
	p.logger.Debug("getting cluster config")

	p.logger.Debugw("getting cluster details from Rancher api", "cluster", input.Cluster.ID)
	clusterDetail, err := p.getClusterDetails(input.Cluster.ID)
	if err != nil {
		return nil, fmt.Errorf("getting cluster detail: %w", err)
	}

	cfg, err := p.getKubeconfig(clusterDetail)
	if err != nil {
		return nil, fmt.Errorf("getting kubeconfig: %w", err)
	}

	if input.Namespace != nil && *input.Namespace != "" {
		p.logger.Debugw("setting kubernetes namespace", "namespace", *input.Namespace)
		cfg.Contexts[cfg.CurrentContext].Namespace = *input.Namespace
	}

	return &discovery.GetConfigOutput{
		KubeConfig:  cfg,
		ContextName: &cfg.CurrentContext,
	}, nil
}

func (p *rancherClusterProvider) getClusterDetails(clusterID string) (*clusterDetails, error) {
	resolver, err := rancher.NewStaticEndpointsResolver(p.config.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("creating endpoint resolver: %w", err)
	}

	headers := defaults.Headers(defaults.WithJSON(), defaults.WithBearerAuth(p.token))
	httpClient := khttp.NewHTTPClient()

	resp, err := httpClient.Get(resolver.Cluster(clusterID), headers)
	if err != nil {
		return nil, fmt.Errorf("getting cluster %s using api: %w", clusterID, err)
	}

	if resp.ResponseCode() != http.StatusOK {
		return nil, ErrGetClusterDetail
	}

	clusterResponse := &clusterDetails{}
	if err := json.Unmarshal([]byte(resp.Body()), clusterResponse); err != nil {
		return nil, fmt.Errorf("unmarshalling api response: %w", err)
	}

	return clusterResponse, nil
}

func (p *rancherClusterProvider) getKubeconfig(clusterDetail *clusterDetails) (*api.Config, error) {
	actionURL, ok := clusterDetail.Actions[generateKubeconfigActionName]
	if !ok {
		return nil, ErrNoKubeconfigAction
	}

	headers := defaults.Headers(defaults.WithJSON(), defaults.WithBearerAuth(p.token))
	httpClient := khttp.NewHTTPClient()

	resp, err := httpClient.Post(actionURL, "{}", headers)
	if err != nil {
		return nil, fmt.Errorf("getting cluster %s kubeconfig using api: %w", clusterDetail.ID, err)
	}

	if resp.ResponseCode() != http.StatusOK {
		return nil, ErrGettingKubeconfig
	}

	kubeconfigResponse := &generateKubeConfigResponse{}
	if err := json.Unmarshal([]byte(resp.Body()), kubeconfigResponse); err != nil {
		return nil, fmt.Errorf("unmarshalling api response: %w", err)
	}

	kubeCfg, err := clientcmd.Load([]byte(kubeconfigResponse.Config))
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig: %w", err)
	}

	return kubeCfg, nil
}
