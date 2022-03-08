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

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/defaults"
	"github.com/fidelity/kconnect/pkg/printer"
)

const (
	NoInputConfigItem        = "no-input"
	NonInteractiveConfigItem = "non-interactive"
	NoVersionCheckConfigItem = "no-version-check"
	ConfigPathConfigItem     = "config"
)

type HistoryLocationConfig struct {
	Location string `json:"history-location"`
}

func AddHistoryLocationItems(cs config.ConfigurationSet) error {
	if _, err := cs.String("history-location", "", "Location of where the history is stored. (default \"$HOME/.kconnect/history.yaml\")"); err != nil {
		return fmt.Errorf("adding history-location config: %w", err)
	}
	cs.SetHistoryIgnore("history-location") //nolint: errcheck
	return nil
}

type HistoryConfig struct {
	HistoryLocationConfig
	MaxItems  int    `json:"max-history"`
	NoHistory bool   `json:"no-history"`
	EntryID   string `json:"entry-id"`
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
	cs.SetHistoryIgnore("max-history") //nolint: errcheck
	cs.SetHistoryIgnore("no-history")  //nolint: errcheck
	cs.SetHistoryIgnore("entry-id")    //nolint: errcheck
	return nil
}

type KubernetesConfig struct {
	Kubeconfig string `json:"kubeconfig"`
}

// AddKubeconfigConfigItems will add the kubeconfig related config items
func AddKubeconfigConfigItems(cs config.ConfigurationSet) error {
	if _, err := cs.String("kubeconfig", "", "Location of the kubeconfig to use. (default \"$HOME/.kube/config\")"); err != nil {
		return fmt.Errorf("adding kubeconfig config: %w", err)
	}
	if err := cs.SetShort("kubeconfig", "k"); err != nil {
		return fmt.Errorf("setting kubeconfig shorthand: %w", err)
	}
	return nil
}

type CommonConfig struct {
	ConfigFile          string `json:"config"`
	Verbosity           int    `json:"verbosity"`
	NoInput             bool   `json:"no-input"`
	DisableVersionCheck bool   `json:"no-version-check"`
}

func AddCommonConfigItems(cs config.ConfigurationSet) error {
	if _, err := cs.String(ConfigPathConfigItem, "", "Configuration file for application wide defaults. (default \"$HOME/.kconnect/config.yaml\")"); err != nil {
		return fmt.Errorf("adding config item: %w", err)
	}
	if _, err := cs.Int("verbosity", 0, "Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace."); err != nil {
		return fmt.Errorf("adding verbosity config: %w", err)
	}
	if _, err := cs.Bool(NonInteractiveConfigItem, false, "Run without interactive flag resolution"); err != nil {
		return fmt.Errorf("adding non-interactive config: %w", err)
	}
	if _, err := cs.Bool(NoInputConfigItem, false, "Explicitly disable interactivity when running in a terminal"); err != nil {
		return fmt.Errorf("adding no-input config: %w", err)
	}
	if _, err := cs.Bool(NoVersionCheckConfigItem, false, "If set to true kconnect will not check for a newer version"); err != nil {
		return fmt.Errorf("adding non-version-check config: %w", err)
	}
	cs.SetShort("verbosity", "v")                                       //nolint: errcheck
	cs.SetHistoryIgnore(ConfigPathConfigItem)                           //nolint: errcheck
	cs.SetHistoryIgnore("verbosity")                                    //nolint: errcheck
	cs.SetHistoryIgnore(NonInteractiveConfigItem)                       //nolint: errcheck
	cs.SetHistoryIgnore(NoInputConfigItem)                              //nolint: errcheck
	cs.SetHistoryIgnore(NoVersionCheckConfigItem)                       //nolint: errcheck
	cs.SetDeprecated(NonInteractiveConfigItem, "please use --no-input") //nolint: errcheck

	return nil
}

type CommonUseConfig struct {
	Namespace string `json:"namespace,omitempty"`
}

func AddCommonUseConfigItems(cs config.ConfigurationSet) error {
	if _, err := cs.String("namespace", "", "Sets namespace for context in kubeconfig"); err != nil {
		return fmt.Errorf("adding config item: %w", err)
	}
	cs.SetShort("namespace", "n") //nolint: errcheck
	return nil
}

type HistoryIdentifierConfig struct {
	Alias string `json:"alias,omitempty"`
	ID    string `json:"id,omitempty"`
}

func AddHistoryIdentifierConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("alias", "", "Alias name for a history entry"); err != nil {
		return fmt.Errorf("adding alias config: %w", err)
	}
	if _, err := cs.String("id", "", "Id for a history entry"); err != nil {
		return fmt.Errorf("adding id config: %w", err)
	}
	cs.SetHistoryIgnore("alias") //nolint: errcheck
	cs.SetHistoryIgnore("id")    //nolint: errcheck
	return nil
}

type HistoryQueryConfig struct {
	Filter string                 `json:"filter,omitempty"`
	Output *printer.OutputPrinter `json:"output,omitempty"`
}

func AddHistoryQueryConfig(cs config.ConfigurationSet) error {

	if _, err := cs.String("filter", "", "filter to apply to import. Can specify multiple filters by using commas, and supports wilcards (*)"); err != nil {
		return fmt.Errorf("adding filter config: %w", err)
	}
	if _, err := cs.String("output", "table", "Output format for the results"); err != nil {
		return fmt.Errorf("adding output config item: %w", err)
	}
	if err := cs.SetShort("output", "o"); err != nil {
		return fmt.Errorf("adding output short flag: %w", err)
	}
	cs.SetHistoryIgnore("output") //nolint: errcheck
	return nil
}

type HistoryImportConfig struct {
	Clean     bool   `json:"clean,omitempty"`
	File      string `json:"file,omitempty"`
	Filter    string `json:"filter,omitempty"`
	Overwrite bool   `json:"overwrite,omitempty"`
	Set       string `json:"set,omitempty"`
}

func AddHistoryImportConfig(cs config.ConfigurationSet) error {
	if _, err := cs.Bool("clean", false, "delete all existing history"); err != nil {
		return fmt.Errorf("adding clean config: %w", err)
	}
	if _, err := cs.String("file", "", "File to import"); err != nil {
		return fmt.Errorf("adding file config: %w", err)
	}
	if err := cs.SetShort("file", "f"); err != nil {
		return fmt.Errorf("adding file shorthand: %w", err)
	}
	if _, err := cs.String("filter", "", "filter to apply to import"); err != nil {
		return fmt.Errorf("adding filter config: %w", err)
	}
	if _, err := cs.Bool("overwrite", false, "overwrite conflicting entries"); err != nil {
		return fmt.Errorf("adding overwrite config: %w", err)
	}
	if _, err := cs.String("set", "", "fields to set"); err != nil {
		return fmt.Errorf("adding set config: %w", err)
	}
	return nil
}

type HistoryExportConfig struct {
	File   string `json:"file,omitempty"`
	Filter string `json:"filter,omitempty"`
	Set    string `json:"set,omitempty"`
}

func AddHistoryExportConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("file", "", "file to import"); err != nil {
		return fmt.Errorf("adding file config: %w", err)
	}
	if err := cs.SetShort("file", "f"); err != nil {
		return fmt.Errorf("adding file short: %w", err)
	}
	if _, err := cs.String("filter", "", "filter to apply to import. Can specify multiple filters by using commas, and supports wilcards (*)"); err != nil {
		return fmt.Errorf("adding filter config: %w", err)
	}
	if _, err := cs.String("set", "", "fields to set"); err != nil {
		return fmt.Errorf("adding set config: %w", err)
	}
	return nil
}

type HistoryRemoveConfig struct {
	All    bool   `json:"all,omitempty"`
	Filter string `json:"filter,omitempty"`
}

func AddHistoryRemoveConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("filter", "", "filter to apply to import. Can specify multiple filters by using commas, and supports wilcards (*)"); err != nil {
		return fmt.Errorf("adding filter config: %w", err)
	}
	if _, err := cs.Bool("all", false, "remove all entries"); err != nil {
		return fmt.Errorf("adding all config: %w", err)
	}
	return nil
}
