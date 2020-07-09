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

package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/fidelity/kconnect/pkg/provider/cluster"
)

var useCmd = &cli.Command{
	Name:  "use",
	Usage: "connect to a target environment and use clusters",
	OnUsageError: func(c *cli.Context, err error, isSubCommand bool) error {
		fmt.Fprintf(c.App.Writer, "use command Error: %v", err)
		return nil
	},
	Action: func(c *cli.Context) error {
		fmt.Println("in the action of use NOT subcommand")
		return nil
	},
	Before: func(c *cli.Context) error {
		return nil
	},
	After: func(c *cli.Context) error {
		return nil
	},
	SkipFlagParsing: true,
}

func init() {
	// Add flags that are common across all
	useCmd.Flags = commonUseFlags()

	//TODO: get from provider factory
	providers := []cluster.ClusterProvider{&cluster.AKSClusterProvider{}}

	for _, provider := range providers {
		providerCmd := &cli.Command{
			Name:  provider.Name(),
			Flags: provider.Flags(),
			Usage: fmt.Sprintf("use the %s cluster provider", provider.Name()),
			Action: func(c *cli.Context) error {
				return doUse(c, provider)
			},
			OnUsageError: func(c *cli.Context, err error, isSubCommand bool) error {
				fmt.Fprintf(c.App.Writer, "Provider Error: %v", err)
				return nil
			},
			Before: func(c *cli.Context) error {
				return nil
			},
			After: func(c *cli.Context) error {
				return nil
			},
			SkipFlagParsing: true,
		}

		useCmd.Subcommands = append(useCmd.Subcommands, providerCmd)
	}

	// TODO: add any additional flags
	App.Commands = append(App.Commands, useCmd)
}

func commonUseFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name: "username",
			//Required: true,
		},
	}
}

func doUse(c *cli.Context, provider cluster.ClusterProvider) error {
	fmt.Println("In do Use")
	fmt.Printf("With provider: %s\n", provider.Name())

	return nil
}
