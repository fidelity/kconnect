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

	"k8s.io/client-go/tools/clientcmd"

	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/config"
)

type HistoryLocationConfig struct {
	Location string `json:"history-location"`
}

type HistoryConfig struct {
	HistoryLocationConfig
	MaxItems  int    `json:"max-history"`
	NoHistory bool   `json:"no-history"`
	EntryID   string `json:"entry-id"`
}

func AddHistoryLocationItems(cs config.ConfigurationSet) error {
	if _, err := cs.String("history-location", defaults.HistoryPath(), "Location of where the history is stored"); err != nil {
		return fmt.Errorf("adding history-location config: %w", err)
	}

	cs.SetHistoryIgnore("history-location") //nolint

	return nil
}

func AddHistoryConfigItems(cs config.ConfigurationSet) error {
	if err := AddHistoryLocationItems(cs); err != nil {
		return err
	}

	if _, err := cs.Int("max-history", defaults.MaxHistoryItems, "Sets the maximum number of history items to keep"); err != nil {
		return fmt.Errorf("adding max-history config: %w", err)
	}
	if _, err := cs.Bool("no-history", false, "If set to true then no history entry will be written"); err != nil {
		return fmt.Errorf("adding no-history config: %w", err)
	}
	if _, err := cs.String("entry-id", "", "existing entry id."); err != nil {
		return fmt.Errorf("adding entry-id config: %w", err)
	}

	if err := cs.SetHidden("entry-id"); err != nil {
		return fmt.Errorf("setting entry-id hidden: %w", err)
	}

	cs.SetHistoryIgnore("max-history") //nolint
	cs.SetHistoryIgnore("no-history")  //nolint
	cs.SetHistoryIgnore("entry-id")    //nolint

	return nil
}

type KubernetesConfig struct {
	Kubeconfig string `json:"kubeconfig"`
}

// AddKubeconfigConfigItems will add the kubeconfig related config items
func AddKubeconfigConfigItems(cs config.ConfigurationSet) error {
	pathOptions := clientcmd.NewDefaultPathOptions()
	if _, err := cs.String("kubeconfig", pathOptions.GetDefaultFilename(), "Location of the kubeconfig to use"); err != nil {
		return fmt.Errorf("adding kubeconfig config: %w", err)
	}
	if err := cs.SetShort("kubeconfig", "k"); err != nil {
		return fmt.Errorf("setting kubeconfig shorthand: %w", err)
	}

	return nil
}

type CommonConfig struct {
	ConfigFile  string `json:"config"`
	Verbosity   int    `json:"verbosity"`
	Interactive bool   `json:"non-interactive"`
}

func AddCommonConfigItems(cs config.ConfigurationSet) error {
	configLocation := defaults.ConfigPath()

	if _, err := cs.String("config", configLocation, "Configuration file for application defaults"); err != nil {
		return fmt.Errorf("adding config item: %w", err)
	}
	if _, err := cs.Int("v", 0, "sets the logging verbosity"); err != nil {
		return fmt.Errorf("adding verbosity config: %w", err)
	}
	if _, err := cs.Bool("non-interactive", false, "Run without interactive flag resolution"); err != nil {
		return fmt.Errorf("adding non-interactive config: %w", err)
	}

	cs.SetShort("verbosity", "v")          //nolint
	cs.SetHistoryIgnore("config")          //nolint
	cs.SetHistoryIgnore("verbosity")       //nolint
	cs.SetHistoryIgnore("non-interactive") //nolint

	return nil
}

func AddHistoryIdentifierConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("alias", "", "Alias name for a history entry"); err != nil {
		return fmt.Errorf("adding alias config: %w", err)
	}
	if _, err := cs.String("id", "", "Id for a history entry"); err != nil {
		return fmt.Errorf("adding id config: %w", err)
	}

	cs.SetHistoryIgnore("alias") //nolint
	cs.SetHistoryIgnore("id")    //nolint

	return nil
}

func AddHistoryQueryConfig(cs config.ConfigurationSet) error {
	if err := AddHistoryIdentifierConfig(cs); err != nil {
		return fmt.Errorf("adding history identifier config items: %w", err)
	}

	if _, err := cs.String("cluster-provider", "", "Name of a cluster provider (i.e. eks)"); err != nil {
		return fmt.Errorf("adding cluster-provider-id config: %w", err)
	}
	if _, err := cs.String("identity-provider", "", "Name of a identity provider (i.e. saml)"); err != nil {
		return fmt.Errorf("adding identity-provider-id config: %w", err)
	}
	if _, err := cs.String("provider-id", "", "Provider specific for a cluster"); err != nil {
		return fmt.Errorf("adding provider-id config: %w", err)
	}

	cs.SetHistoryIgnore("cluster-provider")  //nolint
	cs.SetHistoryIgnore("identity-provider") //nolint
	cs.SetHistoryIgnore("provider-id")       //nolint

	return nil

}
