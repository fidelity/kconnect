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
	"errors"
	"io"

	"k8s.io/apimachinery/pkg/api/meta"
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
)

type ObjectPrinter interface {
	Print(object runtime.Object, writer io.Writer) error
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

func (p *jsonObjectPrinter) Print(object runtime.Object, writer io.Writer) error {

	jsonprinter := cliprint.JSONPrinter{}
	scheme, _, _ := historyv1alpha.NewSchemeAndCodecs()
	printer, err := cliprint.NewTypeSetter(scheme).WrapToPrinter(&jsonprinter, nil)
	if err != nil {
		return err
	}

	if meta.IsListType(object) {
		//TODO: loop around
	}

	printer.PrintObj(object, writer)

	return nil
}

type yamlObjectPrinter struct {
}

func (p *yamlObjectPrinter) Print(object runtime.Object, writer io.Writer) error {
	yamlPrinter := cliprint.YAMLPrinter{}
	scheme, _, _ := historyv1alpha.NewSchemeAndCodecs()
	printer, _ := cliprint.NewTypeSetter(scheme).WrapToPrinter(&yamlPrinter, nil)
	printer.PrintObj(object, writer)

	return nil
}

type tableObjectPrinter struct {
}

func (p *tableObjectPrinter) Print(object runtime.Object, writer io.Writer) error {
	options := cliprint.PrintOptions{}
	tablePrinter := cliprint.NewTablePrinter(options)
	scheme, _, _ := historyv1alpha.NewSchemeAndCodecs()
	printer, err := cliprint.NewTypeSetter(scheme).WrapToPrinter(tablePrinter, nil)
	if err != nil {
		return err
	}

	return printer.PrintObj(object, writer)
}
