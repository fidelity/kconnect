package commands

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

var App = cli.App{
	Name:        "kconnect",
	Description: "The Kubernetes connection manager CLI",
}

func Execute() {
	if err := App.Run(os.Args); err != nil {
		fmt.Printf("Error: %#v", err)
		os.Exit(1)
	}
}

func init() {
	App.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "Configuration file (defaults to $HOME/.kconnect/config)",
		},
	}
}
