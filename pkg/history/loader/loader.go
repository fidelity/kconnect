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

package loader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/defaults"
	"k8s.io/apimachinery/pkg/runtime"
)

type Loader interface {
	Load() (*historyv1alpha.HistoryEntryList, error)

	Save(historyList *historyv1alpha.HistoryEntryList) error
}

func NewFileLoader(path string) (Loader, error) {
	if path == "" {
		path = defaults.HistoryPath()
	}

	historyFile, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("getting absolute file path for %s: %w", path, err)
	}

	info, err := os.Stat(historyFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("getting details of file %s: %w", historyFile, err)
		}

		emptyHistoryFile, err := os.Create(historyFile)
		if err != nil {
			return nil, fmt.Errorf("creating empty history file %s: %w", historyFile, err)
		}

		emptyHistoryFile.Close()
	} else if info.IsDir() {
		return nil, fmt.Errorf("supplied path is a directory %s: %w", historyFile, err)
	}

	return &fileLoader{
		path: historyFile,
	}, nil
}

type fileLoader struct {
	path string
}

func (f *fileLoader) Load() (*historyv1alpha.HistoryEntryList, error) {
	data, err := ioutil.ReadFile(f.path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", f.path, err)
	}

	if len(data) == 0 {
		return historyv1alpha.NewHistoryEntryList(), nil
	}

	_, historyCodecs, err := historyv1alpha.NewSchemeAndCodecs()
	if err != nil {
		return nil, fmt.Errorf("getting history codec: %w", err)
	}

	historyList := &historyv1alpha.HistoryEntryList{}
	if err := runtime.DecodeInto(historyCodecs.UniversalDecoder(), data, historyList); err != nil {
		return nil, fmt.Errorf("decoding history file: %w", err)
	}

	return historyList, nil
}

func (f *fileLoader) Save(historyList *historyv1alpha.HistoryEntryList) error {
	data, err := yaml.Marshal(historyList)
	if err != nil {
		return fmt.Errorf("marshalling history list: %w", err)
	}

	if err := ioutil.WriteFile(f.path, data, os.ModePerm); err != nil {
		return fmt.Errorf("saving history file to %s: %w", f.path, err)
	}

	return nil
}
