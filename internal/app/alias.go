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
	"os"

	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/printer"
	"github.com/fidelity/kconnect/pkg/provider"
)

// AliasListInput defines the inputs for AliasList
type AliasListInput struct {
	HistoryLocationConfig
	Output *printer.OutputPrinter `json:"output,omitempty"`
}

// AliasList implements the alias listing functionality
func (a *App) AliasList(ctx *provider.Context, input *AliasListInput) error {
	zap.S().Debug("listing aliases")

	list, err := a.historyStore.GetAll()
	if err != nil {
		return fmt.Errorf("getting history entries: %w", err)
	}

	aliases := []string{}
	for _, entry := range list.Items {
		if entry.Spec.Alias != nil && *entry.Spec.Alias != "" {
			aliases = append(aliases, *entry.Spec.Alias)
		}
	}

	objPrinter, err := printer.New(*input.Output)
	if err != nil {
		return fmt.Errorf("getting printer for output %s: %w", *input.Output, err)
	}

	if *input.Output == printer.OutputPrinterTable {
		table := printer.ConvertSliceToTable("Alias", aliases)
		return objPrinter.Print(table, os.Stdout)
	}

	return objPrinter.Print(aliases, os.Stdout)
}
