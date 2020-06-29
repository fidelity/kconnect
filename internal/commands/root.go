package commands

import (
	"os"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

var RootCmd = cobra.Command{
	Use:   "kconnect [command]",
	Short: "The Kubernetes Connection Manager CLI",
	Run: func(c *cobra.Command, _ []string) {
		if err := c.Help(); err != nil {
			log.Debugf("ignoring error %s", err.Error())
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolP("help", "h", false, "help for a command")
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "", "", "Configuration file (defaults to $HOME/.kconnect/config)")

}
