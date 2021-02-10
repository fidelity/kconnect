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
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	"github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/printer"
)

// ConfigureInput is the input type for the configure command
type ConfigureInput struct {
	SourceLocation *string                `json:"file,omitempty"`
	Output         *printer.OutputPrinter `json:"output,omitempty"`
	Username       string                 `json:"username,omitempty"`
	Password       string                 `json:"password,omitempty"`
}

var ErrNotOKHTTPStatusCode = errors.New("non 200 status code")

// Configuration implements the configure command
func (a *App) Configuration(ctx context.Context, input *ConfigureInput) error {
	if input.SourceLocation == nil || *input.SourceLocation == "" {
		return a.printConfiguration(input.Output)
	}
	return a.importConfiguration(input)
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

func (a *App) importConfiguration(input *ConfigureInput) error {

	sourceLocation := *input.SourceLocation
	zap.S().Infow("importing configuration", "file", sourceLocation)

	if sourceLocation == "" {
		return ErrSourceLocationRequired
	}

	appConfig, err := config.NewAppConfiguration()
	if err != nil {
		return fmt.Errorf("creating app config: %w", err)
	}

	reader, err := a.getReader(sourceLocation, input.Username, input.Password)
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

func (a *App) getReader(location, username, password string) (io.Reader, error) {
	switch {
	case location == "-":
		return os.Stdin, nil
	case strings.Index(location, "http://") == 0 || strings.Index(location, "https://") == 0:
		url, err := url.Parse(location)
		if err != nil {
			return nil, fmt.Errorf("parsing location as URL %s: %w", location, err)
		}
		headers := make(map[string]string)
		if username != "" && password != "" {
			http.SetBasicAuthHeaders(headers, username, password)
		}
		resp, err := a.httpClient.Get(url.String(), headers)
		if err != nil {
			return nil, fmt.Errorf("error executing request: %w", err)
		}
		if resp.ResponseCode() != http.StatusCodeOK {

			return nil, fmt.Errorf("received status code %d, %s: %w", resp.ResponseCode(), resp.Body(), ErrNotOKHTTPStatusCode)
		}
		reader := strings.NewReader(resp.Body())
		return reader, nil
	default:
		return os.Open(location)
	}
}
