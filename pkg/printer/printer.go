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

package printer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v2"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cliprint "k8s.io/cli-runtime/pkg/printers"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
)

type OutputPrinter string

var (
	OutputPrinterYAML  = OutputPrinter("yaml")
	OutputPrinterJSON  = OutputPrinter("json")
	OutputPrinterTable = OutputPrinter("table")

	SupportedPrinters = []OutputPrinter{OutputPrinterYAML, OutputPrinterJSON, OutputPrinterTable}

	ErrUnknownPrinterOutput = errors.New("unknown printer output type. Supported types are yaml, json, table")
	ErrTableRequired        = errors.New("table printer can only be used with a metav1.Table")
)

type ObjectPrinter interface {
	Print(in any, writer io.Writer) error
}

func New(outputPrinter OutputPrinter) (ObjectPrinter, error) {
	switch outputPrinter {
	case OutputPrinterYAML:
		return &yamlObjectPrinter{}, nil
	case OutputPrinterTable:
		return &tableObjectPrinter{}, nil
	case OutputPrinterJSON:
		return &jsonObjectPrinter{}, nil
	default:
		return nil, ErrUnknownPrinterOutput
	}
}

type jsonObjectPrinter struct {
}

func (p *jsonObjectPrinter) Print(in any, writer io.Writer) error {
	inObj, ok := in.(runtime.Object)
	if ok {
		jsonprinter := &cliprint.JSONPrinter{}
		return printObject(jsonprinter, inObj, writer)
	}

	data, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshing object as json: %w", err)
	}

	_, err = writer.Write(data)

	return err
}

type yamlObjectPrinter struct {
}

func (p *yamlObjectPrinter) Print(in any, writer io.Writer) error {
	inObj, ok := in.(runtime.Object)
	if ok {
		yamlPrinter := &cliprint.YAMLPrinter{}
		return printObject(yamlPrinter, inObj, writer)
	}

	data, err := yaml.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshing object as yaml: %w", err)
	}

	_, err = writer.Write(data)

	return err
}

type tableObjectPrinter struct {
}

func (p *tableObjectPrinter) Print(in any, writer io.Writer) error {
	inObj, ok := in.(*metav1.Table)
	if !ok {
		return ErrTableRequired
	}

	options := cliprint.PrintOptions{}
	tablePrinter := cliprint.NewTablePrinter(options)
	scheme, _, _ := historyv1alpha.NewSchemeAndCodecs()

	printer, err := cliprint.NewTypeSetter(scheme).WrapToPrinter(tablePrinter, nil)
	if err != nil {
		return err
	}

	return printer.PrintObj(inObj, writer)
}

func printObject(resPrinter cliprint.ResourcePrinter, object runtime.Object, writer io.Writer) error {
	scheme, _, _ := historyv1alpha.NewSchemeAndCodecs()

	printer, err := cliprint.NewTypeSetter(scheme).WrapToPrinter(resPrinter, nil)
	if err != nil {
		return err
	}

	if !meta.IsListType(object) {
		return printer.PrintObj(object, writer)
	}

	items, err := meta.ExtractList(object)
	if err != nil {
		return fmt.Errorf("extracting list: %w", err)
	}

	for _, item := range items {
		if err := printer.PrintObj(item, writer); err != nil {
			return err
		}
	}

	return nil
}

// ConvertSliceToTable will convert a string slice to a table
func ConvertSliceToTable(columnName string, items []string) *metav1.Table {
	table := &metav1.Table{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metav1.SchemeGroupVersion.String(),
			Kind:       "Table",
		},
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{
				Name: columnName,
				Type: "string",
			},
		},
	}

	for _, item := range items {
		row := metav1.TableRow{
			Cells: []any{item},
		}
		table.Rows = append(table.Rows, row)
	}

	return table
}
