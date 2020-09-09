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

package flags

import (
	"errors"
	"fmt"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// ErrFlagMissing is an error when there is no flag with a given name
	ErrFlagMissing = errors.New("flag missing")
)

// ExistsWithValue returns true if a flag exists in a flagset and has a value
// and that value is non-empty
func ExistsWithValue(name string, flags *pflag.FlagSet) bool {
	flag := flags.Lookup(name)
	if flag == nil {
		return false
	}

	if flag.Value == nil {
		return false
	}

	if flag.Value.String() == "" {
		return false
	}

	return true
}

// CreateFlagsFromConfig will create a FlagSet from a configuration set
func CreateFlagsFromConfig(cs config.ConfigurationSet) (*pflag.FlagSet, error) {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)

	for _, configItem := range cs.GetAll() {
		switch configItem.Type {
		case config.ItemTypeString:
			defVal := configItem.DefaultValue.(string)
			fs.String(configItem.Name, defVal, configItem.Description)
		case config.ItemTypeInt:
			defVal := configItem.DefaultValue.(int)
			fs.Int(configItem.Name, defVal, configItem.Description)
		case config.ItemTypeBool:
			defVal := configItem.DefaultValue.(bool)
			fs.Bool(configItem.Name, defVal, configItem.Description)
		default:
			return nil, config.ErrUnknownItemType
		}
	}

	return fs, nil
}

func PopulateConfigFromFlags(flags *pflag.FlagSet, cs config.ConfigurationSet) {
	flags.VisitAll(func(f *pflag.Flag) {

		switch f.Value.Type() {
		case "bool":
			val, _ := flags.GetBool(f.Name)
			cs.SetValue(f.Name, val) //nolint: errcheck
		case "string":
			cs.SetValue(f.Name, f.Value.String()) //nolint: errcheck
		case "int":
			val, _ := flags.GetInt(f.Name)
			cs.SetValue(f.Name, val) //nolint: errcheck
		}
	})
}

func PopulateConfigFromCommand(cmd *cobra.Command, cs config.ConfigurationSet) {
	PopulateConfigFromFlags(cmd.Flags(), cs)
	PopulateConfigFromFlags(cmd.PersistentFlags(), cs)
}

func AddCommonCommandConfig(cs config.ConfigurationSet) error {
	if _, err := cs.String("config", "", "Configuration file (defaults to $HOME/.kconnect/config"); err != nil {
		return fmt.Errorf("adding config item: %w", err)
	}
	if _, err := cs.String("log-level", logrus.DebugLevel.String(), "Log level for the CLI. Defaults to INFO"); err != nil {
		return fmt.Errorf("adding log-level config: %w", err)
	}
	if _, err := cs.String("log-format", "TEXT", "Format of the log output. Defaults to text."); err != nil {
		return fmt.Errorf("adding log-format config: %w", err)
	}
	if err := cs.SetShort("log-level", "l"); err != nil {
		return fmt.Errorf("setting shorthand for log-level: %w", err)
	}

	return nil
}

// ConvertToMap will convert a flagset to a map
func ConvertToMap(fs *pflag.FlagSet) map[string]string {
	flags := make(map[string]string)
	fs.VisitAll(func(flag *pflag.Flag) {
		flags[flag.Name] = flag.Value.String()
	})

	return flags
}
