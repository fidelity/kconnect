package commands

import (
	"fmt"
	"os"

	"github.com/kris-nova/logger"
	home "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile string
)

var RootCmd = &cobra.Command{
	Use:   "kconnect",
	Short: "The Kubernetes Connection Manager CLI",
	Run: func(c *cobra.Command, _ []string) {
		if err := c.Help(); err != nil {
			logger.Debug("ignoring cobra error %q", err.Error())
		}
	},
	SilenceUsage: true,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file (defaults to $HOME/.kconnect/config")
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := home.Dir()
		if err != nil {
			panic(err)
		}

		//TODO: construct path properyl
		viper.AddConfigPath(home)
		viper.SetConfigName(".kconnect")
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
