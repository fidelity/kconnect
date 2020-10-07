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

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/internal/commands"
	intver "github.com/fidelity/kconnect/internal/version"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/logging"
	_ "github.com/fidelity/kconnect/pkg/plugins" // Import all the plugins
)

func main() {
	if err := setupLogging(); err != nil {
		log.Fatalf("failed to configure logging %v", err)
	}

	v := intver.Get()
	zap.S().Infow("kconnect - the kubernetes cli", "version", v.Version)
	zap.S().Debugw("build information", "date", v.BuildDate, "commit", v.CommitHash, "gover", v.GoVersion)

	rootCmd, err := commands.RootCmd()
	if err != nil {
		zap.S().Fatalw("failed getting root command", "error", err.Error())
	}
	if err := rootCmd.Execute(); err != nil {
		zap.S().Fatalw("failed executing root command", "error", err.Error())
	}
}

func setupLogging() error {
	verbosity, err := flags.GetFlagValueDirect(os.Args, "verbosity", "v")
	if err != nil {
		return fmt.Errorf("getting verbosity flag: %w", err)
	}

	logVerbosity := 0
	if verbosity != "" {
		logVerbosity, err = strconv.Atoi(verbosity)
		if err != nil {
			return fmt.Errorf("parsing verbosity level: %w", err)
		}
	}

	if err := logging.Configure(logVerbosity); err != nil {
		log.Fatalf("failed to configure logging %v", err)
	}

	return nil
}
