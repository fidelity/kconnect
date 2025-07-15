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

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/fidelity/kconnect/pkg/defaults"
	"github.com/fidelity/kconnect/pkg/history"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/discovery"
)

const (
	// Name is the name of the application
	Name = "kconnect"
)

// App represents the kconnect application and contains the
// implementation of the apps logic
type App struct {
	historyStore    history.Store
	configDirectory string
	sensitiveFlags  map[string]*pflag.Flag

	selectCluster SelectClusterFunc
	itemSelector  provider.SelectItemFunc

	interactive bool
	httpClient  khttp.Client
	logger      *zap.SugaredLogger
}

// Option represents an option to use with the kcinnect application
type Option func(*App)

// SelectClusterFunc is a function type that is used to allow selection of a cluster
type SelectClusterFunc func(discoverOutput *discovery.DiscoverOutput) (*discovery.Cluster, error)

// New creates a new instance of the kconnect app with options
func New(opts ...Option) *App {
	app := &App{
		configDirectory: defaults.AppDirectory(),
		selectCluster:   DefaultSelectCluster,
		sensitiveFlags:  make(map[string]*pflag.Flag),
		logger:          zap.S().With("app", Name),
		httpClient:      khttp.NewHTTPClient(),
		interactive:     true,
		itemSelector:    provider.DefaultItemSelection,
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

func WithLogger(logger *zap.SugaredLogger) Option {
	return func(a *App) {
		a.logger = logger.With("app", Name)
	}
}

func WithHTTPClient(client khttp.Client) Option {
	return func(a *App) {
		a.httpClient = client
	}
}

func WithItemSelectorFunc(itemSelector provider.SelectItemFunc) Option {
	return func(a *App) {
		a.itemSelector = itemSelector
	}
}

func WithSelectClusterFunc(selectFunc SelectClusterFunc) Option {
	return func(a *App) {
		a.selectCluster = selectFunc
	}
}

func WithInteractive(interactive bool) Option {
	return func(a *App) {
		a.interactive = interactive
	}
}

// DefaultSelectCluster is the default cluster selection function. If there is only
// 1 cluster then it automatically selects it. If there are more than 1 cluster then
// a selection is displayed and the user must choose one
func DefaultSelectCluster(discoverOutput *discovery.DiscoverOutput) (*discovery.Cluster, error) {
	clusterNameToID := make(map[string]string)
	options := []string{}

	for _, cluster := range discoverOutput.Clusters {
		clusterNameToID[cluster.Name] = cluster.ID
		options = append(options, cluster.Name)
	}

	clusterName, err := prompt.Choose("cluster", "Select a cluster", true, prompt.OptionsFromStringSlice(options))
	if err != nil {
		return nil, fmt.Errorf("choosing cluster: %w", err)
	}

	clusterID := clusterNameToID[clusterName]
	zap.S().Debugw("selected cluster", "id", clusterID)

	return discoverOutput.Clusters[clusterID], nil
}

func (a *App) SetNonInteractive() {
	a.interactive = false
}
