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
	"strings"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

func CreateCommandFlags(cmd *cobra.Command, cs config.ConfigurationSet) error {
	flags, err := CreateFlagsFromConfig(cs)
	if err != nil {
		return fmt.Errorf("creating flags from config set: %w", err)
	}

	cmd.Flags().AddFlagSet(flags)

	return nil
}

// CreateFlagsFromConfig will create a FlagSet from a configuration set
func CreateFlagsFromConfig(cs config.ConfigurationSet) (*pflag.FlagSet, error) {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)

	for _, configItem := range cs.GetAll() {
		switch configItem.Type {
		case config.ItemTypeString:
			defVal := configItem.DefaultValue.(string)
			if configItem.Shorthand != "" {
				fs.StringP(configItem.Name, configItem.Shorthand, defVal, configItem.Description)
			} else {
				fs.String(configItem.Name, defVal, configItem.Description)
			}
		case config.ItemTypeInt:
			defVal := configItem.DefaultValue.(int)
			if configItem.Shorthand != "" {
				fs.IntP(configItem.Name, configItem.Shorthand, defVal, configItem.Description)
			} else {
				fs.Int(configItem.Name, defVal, configItem.Description)
			}
		case config.ItemTypeBool:
			defVal := configItem.DefaultValue.(bool)
			if configItem.Shorthand != "" {
				fs.BoolP(configItem.Name, configItem.Shorthand, defVal, configItem.Description)
			} else {
				fs.Bool(configItem.Name, defVal, configItem.Description)
			}
		default:
			return nil, config.ErrUnknownItemType
		}

		if configItem.Deprecated {
			if err := fs.MarkDeprecated(configItem.Name, configItem.DeprecatedMessage); err != nil {
				return nil, fmt.Errorf("marking flag deprecated %s: %w", configItem.Name, err)
			}
		}

		if configItem.Hidden {
			if err := fs.MarkHidden(configItem.Name); err != nil {
				return nil, fmt.Errorf("marking flag hidden %s: %w", configItem.Name, err)
			}
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

// ConvertToMap will convert a flagset to a map
func ConvertToMap(fs *pflag.FlagSet) map[string]string {
	flags := make(map[string]string)

	fs.VisitAll(func(flag *pflag.Flag) {
		flags[flag.Name] = flag.Value.String()
	})

	return flags
}

func BindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindEnv(f.Name) //nolint: errcheck

		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)) //nolint: errcheck
		}
	})
}

// GetFlagValueDirect will get a flag value directly from args or config.
// Note: this should only be used in exceptional circumstances
func GetFlagValueDirect(args []string, longName, shortName string) (string, error) {
	if value, found := GetFlagValueFromArgs(args, longName, shortName); found {
		return value, nil
	}

	// look in app config
	return config.GetValue(longName, "")
}

// GetFlagValueFromArgs looks for a flag in the raw command line arguments,
// before they have been parsed, and returns its value. The forms "--name value",
// "--name=value", "-n value" and "-n=value" are supported. The second return
// value is false if the flag isn't present or if it has no value.
func GetFlagValueFromArgs(args []string, longName, shortName string) (string, bool) {
	names := []string{"--" + longName}
	if shortName != "" {
		names = append(names, "-"+shortName)
	}

	for i, arg := range args {
		// Everything after a "--" is a positional argument and not a flag
		if arg == "--" {
			break
		}

		for _, name := range names {
			if arg == name {
				if i+1 < len(args) {
					return args[i+1], true
				}

				return "", false
			}

			if value, ok := strings.CutPrefix(arg, name+"="); ok {
				return value, true
			}
		}
	}

	return "", false
}

func CopyFlagValue(sourceFlagName, destinationFlagName string, fs *pflag.FlagSet, ignoreNotFound bool) error {
	sourceFlag := fs.Lookup(sourceFlagName)
	if sourceFlag == nil && !ignoreNotFound {
		return fmt.Errorf("getting source flag %s: %w", sourceFlagName, ErrFlagNotFound)
	}

	destinationFlag := fs.Lookup(destinationFlagName)
	if destinationFlag == nil && !ignoreNotFound {
		return fmt.Errorf("getting destination flag %s: %w", destinationFlagName, ErrFlagNotFound)
	}

	if sourceFlag != nil && sourceFlag.Changed && destinationFlag != nil {
		destinationFlag.Value = sourceFlag.Value
	}

	return nil
}

func ParseFlagMultiValueToMap(flag string) map[string]string {
	flagsMap := make(map[string]string)

	if flag != "" {
		setFlagsArray := strings.SplitSeq(flag, ",")
		for e := range setFlagsArray {
			// Only split on the first "=" as a value can contain one, for
			// example a URL with a query string.
			key, value, found := strings.Cut(e, "=")
			if !found || key == "" {
				zap.S().Warnf("ignoring %q as it is not in the form key=value", e)
				continue
			}

			flagsMap[key] = value
		}
	}

	return flagsMap
}
