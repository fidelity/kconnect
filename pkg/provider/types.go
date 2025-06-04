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

package provider

import (
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/config"
	khttp "github.com/fidelity/kconnect/pkg/http"
)

// Plugin is an interface that can be implemented for returned usage information
type Plugin interface {
	// Name returns the name of the plugin
	Name() string
}

// PluginPreReqs is an interface that providers have implement to
// indicate that they have pre-requisites that they can check for
type PluginPreReqs interface {
	ListPreReqs() []*PreReq
	CheckPreReqs() error
}

// PreReq represents a pre-requisite
type PreReq interface {
	Name() string
	Help() string
	Check() error
}

// PluginCreationInput is the input to plugin Init
type PluginCreationInput struct {
	Logger        *zap.SugaredLogger
	IsInteractive bool
	ItemSelector  SelectItemFunc
	ScopedTo      *string
	HTTPClient    khttp.Client
}

// ConfigurationItemsFunc is a function type that gets configuration items.
// The scopeTo will indicate if there is additional scope that is needed
type ConfigurationItemsFunc func(scopeTo string) (config.ConfigurationSet, error)
