package helpers

import (
	"fmt"

	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/defaults"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/spf13/cobra"
)

func GetCommonConfig(cmd *cobra.Command, cfg config.ConfigurationSet) (*app.CommonConfig, error) {
	flags.PopulateConfigFromCommand(cmd, cfg)

	configPath := cfg.ValueString(app.ConfigPathConfigItem)
	if configPath == "" {
		configPath = defaults.ConfigPath()
	}

	if err := config.ApplyToConfigSetWithProvider(configPath, cfg, ""); err != nil {
		return nil, fmt.Errorf("applying app config: %w", err)
	}
	params := &app.CommonConfig{}
	if err := config.Unmarshall(cfg, params); err != nil {
		return nil, fmt.Errorf("unmarshalling config into to params: %w", err)
	}

	return params, nil
}
