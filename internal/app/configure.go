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
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/printer"
	"github.com/fidelity/kconnect/pkg/provider"
)

// ConfigureInput is the input type for the configure command
type ConfigureInput struct {
	SourceLocation *string                `json:"file,omitempty"`
	Output         *printer.OutputPrinter `json:"output,omitempty"`
}

// Configuration implements the configure command
func (a *App) Configuration(ctx *provider.Context, input *ConfigureInput) error {
	if input.SourceLocation == nil || *input.SourceLocation == "" {
		return a.printConfiguration(input.Output)
	}
	return a.importConfiguration(*input.SourceLocation)
}

func (a *App) printConfiguration(printerType *printer.OutputPrinter) error {
	zap.S().Debug("printing configuration")

	appConfig, err := config.NewAppConfiguration()
	if err != nil {
		return fmt.Errorf("creating app config: %w", err)
	}

	cfg, err := appConfig.Get()
	if err != nil {
		return fmt.Errorf("getting app config: %w", err)
	}

	objPrinter, err := printer.New(*printerType)
	if err != nil {
		return fmt.Errorf("getting printer for output %s: %w", *printerType, err)
	}

	if *printerType == printer.OutputPrinterTable {
		return objPrinter.Print(cfg.ToTable(), os.Stdout)
	}

	return objPrinter.Print(cfg, os.Stdout)
}

func (a *App) importConfiguration(sourceLocation string) error {
	zap.S().Infow("importing configuration", "file", sourceLocation)

	if sourceLocation == "" {
		return ErrSourceLocationRequired
	}

	appConfig, err := config.NewAppConfiguration()
	if err != nil {
		return fmt.Errorf("creating app config: %w", err)
	}

	reader, err := getReader(sourceLocation)
	if err != nil {
		return fmt.Errorf("getting reader from location: %w", err)
	}

	cfg, err := appConfig.Parse(reader)
	if err != nil {
		return fmt.Errorf("parsing config from reader: %w", err)
	}

	if err := appConfig.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	zap.S().Info("successfully imported configuration")

	return nil
}

func getReader(location string) (io.Reader, error) {
	switch {
	case location == "-":
		return os.Stdin, nil
	case strings.Index(location, "http://") == 0 || strings.Index(location, "https://") == 0:
		url, err := url.Parse(location)
		if err != nil {
			return nil, fmt.Errorf("parsing location as URL %s: %w", location, err)
		}
		resp, err := http.Get(url.String()) //nolint
		if err != nil {
			return nil, fmt.Errorf("getting configuration from %s: %w", location, err)
		}
		return resp.Body, nil
	default:
		return os.Open(location)
	}
}
