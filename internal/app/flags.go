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

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
)

type HistoryConfig struct {
	HistoryLocation string `json:"history-location"`
	HistoryMaxItems int    `json:"max-history"`
}

func AddHistoryConfigItems(cs config.ConfigurationSet) error {
	if _, err := cs.String("history-location", defaults.HistoryPath(), "the location of where the history is stored"); err != nil {
		return fmt.Errorf("adding history-location config: %w", err)
	}
	if _, err := cs.Int("max-history", defaults.MaxHistoryItems, "sets the maximum number of history items to keep"); err != nil {
		return fmt.Errorf("adding max-history config: %w", err)
	}

	return nil
}

type KubernetesConfig struct {
	Kubeconfig string `json:"kubeconfig"`
}

// AddKubeconfigConfigItems will add the kubeconfig related config items
func AddKubeconfigConfigItems(cs config.ConfigurationSet) error {
	pathOptions := clientcmd.NewDefaultPathOptions()
	if _, err := cs.String("kubeconfig", pathOptions.GetDefaultFilename(), "location of the kubeconfig to use"); err != nil {
		return fmt.Errorf("adding kubeconfig config: %w", err)
	}
	if err := cs.SetShort("kubeconfig", "k"); err != nil {
		return fmt.Errorf("setting kubeconfig shorthand: %w", err)
	}

	return nil
}

type CommonConfig struct {
	ConfigFile  string `json:"config"`
	LogLevel    string `json:"log-level"`
	LogFormat   string `json:"log-format"`
	Interactive bool   `json:"non-interactive"`
}

func AddCommonConfigItems(cs config.ConfigurationSet) error {
	if _, err := cs.String("config", "", "Configuration file (defaults to $HOME/.kconnect/config"); err != nil {
		return fmt.Errorf("adding config item: %w", err)
	}
	if _, err := cs.String("log-level", logrus.DebugLevel.String(), "Log level for the CLI. Defaults to INFO"); err != nil {
		return fmt.Errorf("adding log-level config: %w", err)
	}
	if _, err := cs.String("log-format", "TEXT", "Format of the log output. Defaults to text."); err != nil {
		return fmt.Errorf("adding log-format config: %w", err)
	}
	if _, err := cs.Bool("non-interactive", false, "Run without interactive flag resolution. Defaults to false"); err != nil {
		return fmt.Errorf("adding non-interactive config: %w", err)
	}

	if err := cs.SetShort("log-level", "l"); err != nil {
		return fmt.Errorf("setting shorthand for log-level: %w", err)
	}

	return nil
}
