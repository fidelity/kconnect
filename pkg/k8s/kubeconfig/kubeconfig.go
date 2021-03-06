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

package kubeconfig

import (
	"fmt"

	"go.uber.org/zap"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Write will write the kubeconfig to the specified file. If there
// is an existing kubeconfig it will be merged if flag is set to true
func Write(path string, clusterConfig *api.Config, merge, setCurrent bool) error {
	zap.S().Debugw("writing kubeconfig", "path", path)

	pathOptions := clientcmd.NewDefaultPathOptions()
	if path != "" {
		pathOptions.LoadingRules.ExplicitPath = path
	}
	var newConfig *api.Config
	if merge {
		existingConfig, err := pathOptions.GetStartingConfig()
		if err != nil {
			return fmt.Errorf("getting existing kubeconfig: %w", err)
		}

		zap.S().Debug("merging kubeconfig files")
		for k, v := range clusterConfig.Clusters {
			existingConfig.Clusters[k] = v
		}
		for k, v := range clusterConfig.AuthInfos {
			existingConfig.AuthInfos[k] = v
		}
		for k, v := range clusterConfig.Contexts {
			existingConfig.Contexts[k] = v
		}
		newConfig = existingConfig
	} else {
		newConfig = clusterConfig
	}

	if setCurrent {
		zap.S().Infow("setting current context", "context", clusterConfig.CurrentContext)
		newConfig.CurrentContext = clusterConfig.CurrentContext
	}

	if err := clientcmd.ModifyConfig(pathOptions, *newConfig, true); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	zap.S().Infow("kubeconfig updated", "path", path)

	return nil
}

func Read(path string) (*api.Config, error) {

	pathOptions := clientcmd.NewDefaultPathOptions()
	if path != "" {
		pathOptions.LoadingRules.ExplicitPath = path
	}

	existingConfig, err := pathOptions.GetStartingConfig()
	if err != nil {
		return nil, fmt.Errorf("getting existing kubeconfig: %w", err)
	}
	return existingConfig, nil
}

func GetCurrentContext(path string) (*api.Context, error) {

	pathOptions := clientcmd.NewDefaultPathOptions()
	if path != "" {
		pathOptions.LoadingRules.ExplicitPath = path
	}
	existingConfig, err := pathOptions.GetStartingConfig()
	if err != nil {
		return nil, fmt.Errorf("getting existing kubeconfig: %w", err)
	}
	currentContext := existingConfig.Contexts[existingConfig.CurrentContext]
	return currentContext, nil
}
