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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/printer"
	"github.com/fidelity/kconnect/pkg/provider"
)

type HistoryQueryInput struct {
	HistoryConfig
	KubernetesConfig

	ClusterProvider  *string `json:"cluster-provider,omitempty"`
	IdentityProvider *string `json:"identity-provider,omitempty"`

	ProviderID *string `json:"provider-id,omitempty"`
	HistoryID  *string `json:"id,omitempty"`
	Alias      *string `json:"alias,omitempty"`

	Flags map[string]string `json:"flags,omitempty"`

	Output *printer.OutputPrinter `json:"output,omitempty"`

	Context *provider.Context
}

func (a *App) QueryHistory(ctx *provider.Context, input *HistoryQueryInput) error {
	a.logger.Debug("Querying history")

	list, err := a.historyStore.GetAll()
	if err != nil {
		return fmt.Errorf("getting history entries: %w", err)
	}

	filterSpec := &history.FilterSpec{
		Alias:            input.Alias,
		ClusterProvider:  input.ClusterProvider,
		Flags:            input.Flags,
		HistoryID:        input.HistoryID,
		IdentityProvider: input.IdentityProvider,
		Kubeconfig:       &input.Kubeconfig,
		ProviderID:       input.ProviderID,
	}

	if err := history.FilterHistory(list, filterSpec); err != nil {
		return fmt.Errorf("filtering history list: %w", err)
	}

	objPrinter, err := printer.New(*input.Output)
	if err != nil {
		return fmt.Errorf("getting printer for output %s: %w", *input.Output, err)
	}

	if *input.Output == printer.OutputPrinterTable {
		table := convertToTable(list)
		return objPrinter.Print(table, os.Stdout)

	}

	return objPrinter.Print(list, os.Stdout)
}

func convertToTable(list *historyv1alpha.HistoryEntryList) *metav1.Table {
	table := &metav1.Table{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metav1.SchemeGroupVersion.String(),
			Kind:       "Table",
		},
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{
				Name: "Id",
				Type: "string",
			},
			{
				Name: "Provider",
				Type: "string",
			},
			{
				Name: "ProviderID",
				Type: "string",
			},
			{
				Name: "Identity",
				Type: "string",
			},
		},
	}

	for _, entry := range list.Items {
		row := metav1.TableRow{
			Cells: []interface{}{entry.Spec.ID, entry.Spec.Provider, entry.Spec.ProviderID, entry.Spec.Identity},
		}
		table.Rows = append(table.Rows, row)
	}

	return table
}
