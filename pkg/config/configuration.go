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

package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	kconnectv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/defaults"
)

const (
	// ListPrefix is the prefix for a list name
	ListPrefix = "$"
)

type AppConfiguration interface {
	Get() (*kconnectv1alpha.Configuration, error)
	Save(configuration *kconnectv1alpha.Configuration) error
	Parse(reader io.Reader) (*kconnectv1alpha.Configuration, error)
}

func NewAppConfiguration() (AppConfiguration, error) {
	configPath := defaults.ConfigPath()

	return NewAppConfigurationWithPath(configPath)
}

func NewAppConfigurationWithPath(path string) (AppConfiguration, error) {
	configPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("getting config absolute file path for %s: %w", path, err)
	}

	configDirPath := filepath.Dir(configPath)

	info, err := os.Stat(configDirPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("getting details of config dir %s: %w", configDirPath, err)
		}

		err = os.MkdirAll(configDirPath, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("creating config dir %s: %w", configDirPath, err)
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("supplied path is not directory %s: %w", configDirPath, err)
	}

	info, err = os.Stat(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("getting details of config file %s: %w", configPath, err)
		}

		emptyConfigFile, err := os.Create(configPath)
		if err != nil {
			return nil, fmt.Errorf("creating empty config file %s: %w", configPath, err)
		}

		emptyConfigFile.Close()
	} else if info.IsDir() {
		return nil, fmt.Errorf("supplied path is a directory %s: %w", configPath, err)
	}

	return &appConfiguration{path: configPath}, nil
}

func GetValue(name string, provider string) (string, error) {
	appCfg, err := NewAppConfiguration()
	if err != nil {
		return "", fmt.Errorf("creating application configuration: %w", err)
	}

	cfg, err := appCfg.Get()
	if err != nil {
		return "", fmt.Errorf("getting application configuration: %w", err)
	}

	if provider != "" {
		providerValues, hasProvider := cfg.Spec.Providers[provider]
		if hasProvider {
			providerVal, hasProviderVal := providerValues[name]
			if hasProviderVal {
				return providerVal, nil
			}
		}
	}

	globalVal, hasGlobalVal := cfg.Spec.Global[name]
	if hasGlobalVal {
		return globalVal, nil
	}

	return "", nil
}

type appConfiguration struct {
	path string
}

func (a *appConfiguration) Get() (*kconnectv1alpha.Configuration, error) {
	data, err := ioutil.ReadFile(a.path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", a.path, err)
	}

	if len(data) == 0 {
		return kconnectv1alpha.NewConfiguration(), nil
	}

	_, apiCodecs, err := kconnectv1alpha.NewSchemeAndCodecs()
	if err != nil {
		return nil, fmt.Errorf("getting kconnect codec: %w", err)
	}

	appConfiguration := &kconnectv1alpha.Configuration{}
	if err := runtime.DecodeInto(apiCodecs.UniversalDecoder(), data, appConfiguration); err != nil {
		return nil, fmt.Errorf("decoding config file: %w", err)
	}

	if appConfiguration.Spec.VersionCheck == nil {
		appConfiguration.Spec.VersionCheck = &kconnectv1alpha.VersionCheck{}
	}

	return appConfiguration, nil
}

func (a *appConfiguration) Save(configuration *kconnectv1alpha.Configuration) error {
	data, err := yaml.Marshal(configuration)
	if err != nil {
		return fmt.Errorf("marshalling configuration: %w", err)
	}

	if err := ioutil.WriteFile(a.path, data, os.ModePerm); err != nil {
		return fmt.Errorf("saving configuration file to %s: %w", a.path, err)
	}

	return nil
}

func (a *appConfiguration) Parse(reader io.Reader) (*kconnectv1alpha.Configuration, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading all from reader: %w", err)
	}

	_, apiCodecs, err := kconnectv1alpha.NewSchemeAndCodecs()
	if err != nil {
		return nil, fmt.Errorf("getting kconnect codec: %w", err)
	}

	appConfiguration := &kconnectv1alpha.Configuration{}
	if err := runtime.DecodeInto(apiCodecs.UniversalDecoder(), data, appConfiguration); err != nil {
		return nil, fmt.Errorf("decoding config file: %w", err)
	}

	return appConfiguration, nil
}
