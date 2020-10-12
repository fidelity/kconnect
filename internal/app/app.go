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

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/pflag"

	"github.com/fidelity/kconnect/internal/defaults"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/provider"
)

// App represents the kconnect application and contains the
// implementation of the apps logic
type App struct {
	historyStore    history.Store
	configDirectory string
	selectCluster   SelectClusterFunc
	sensitiveFlags  map[string]*pflag.Flag
}

// Option represents an option to use with the kcinnect application
type Option func(*App)

// SelectClusterFunc is a function type that is used to allow selecttion of a cluster
type SelectClusterFunc func(discoverOutput *provider.DiscoverOutput) (*provider.Cluster, error)

// New creates a new instance of the kconnect app with options
func New(opts ...Option) *App {
	app := &App{
		configDirectory: defaults.AppDirectory(),
		selectCluster:   DefaultSelectCluster,
		sensitiveFlags:  make(map[string]*pflag.Flag),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

// WithSelectClusterFn is an option to allow using a custom
// select cluster function
func WithSelectClusterFn(fn SelectClusterFunc) Option {
	return func(a *App) {
		a.selectCluster = fn
	}
}

func WithHistoryStore(store history.Store) Option {
	return func(a *App) {
		a.historyStore = store
	}
}

// DefaultSelectCluster is the default cluster selection function. If there is only
// 1 cluster then it automatically selects it. If there are more than 1 cluster then
// a selection is displayed and the user must choose one
func DefaultSelectCluster(discoverOutput *provider.DiscoverOutput) (*provider.Cluster, error) {
	options := []string{}
	for _, cluster := range discoverOutput.Clusters {
		options = append(options, cluster.ID)
	}

	if len(options) == 1 {
		return discoverOutput.Clusters[options[0]], nil
	}

	clusterName := ""
	prompt := &survey.Select{
		Message:  "Select the cluster",
		Options:  options,
		PageSize: defaults.DefaultUIPageSize,
		Help:     "Select a cluster to connect to from the discovered clusters",
	}
	if err := survey.AskOne(prompt, &clusterName, survey.WithValidator(survey.Required)); err != nil {
		return nil, fmt.Errorf("selecting a cluster: %w", err)
	}

	return discoverOutput.Clusters[clusterName], nil
}
