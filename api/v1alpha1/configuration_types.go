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

package v1alpha1

import (
	"bytes"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigurationSpec represents the configuration of kconnect
type ConfigurationSpec struct {
	// Global holds global configuration in the form of key:value pairs
	Global map[string]string `json:"global,omitempty"`
	// Providers holds provider specific configuration valies in the form of
	// key:value pairs per provider name. Values specified here will
	// overwrite an global config of the same name. If a value of a config
	// starts with $ then this is taken to be a name of a list to use
	// duration resolution to allow the user to select from the list.
	Providers map[string]map[string]string `json:"providers,omitempty"`
	// Lists contains predefined name lists of name/value pairs that can be
	// used to offer a selection to a user for a configuration item
	Lists map[string][]ListItem `json:"lists,omitempty"`
	// ImportedFrom holds where this configuration was originally imported from
	ImportedFrom *string `json:"importedFrom,omitempty"`
	// VersionCheck holds details of the last version cehck
	VersionCheck *VersionCheck `json:"versionCheck,omitempty"`
}

// AppDefaults represents the default values for the kconnect app
type AppDefaults struct {
	Name string
	Args map[string]string `json:"args,omitempty"`
}

// VersionCheck represents version information from a check
type VersionCheck struct {
	// LastChecked holds the date/time the last check was made to see
	// if there was a newer version available
	LastChecked metav1.Time `json:"lastVersionCheck"`
	// LatestReleaseVersion holds the latest release version of kconnect thats been retrieved from GitHub
	LatestReleaseVersion *string `json:"latestReleaseVersion,omitempty"`
	// LatestReleaseURL holds the URL for latest release version of kconnect thats been retrieved from GitHub
	LatestReleaseURL *string `json:"latestReleaseURL,omitempty"`
}

// ListItem represents an item in a list
type ListItem struct {
	// Name is the the name to display to the user for the list item
	Name string `json:"name"`
	// Value is the value to use if the list item is selected
	Value string `json:"value"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration represents the kconnect configuration
type Configuration struct {
	metav1.TypeMeta `json:",inline"`

	Spec ConfigurationSpec `json:"spec,omitempty"`
}

// NewConfiguration will create a new configuration
func NewConfiguration() *Configuration {
	return &Configuration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "Configuration",
		},
		Spec: ConfigurationSpec{
			VersionCheck: &VersionCheck{},
		},
	}
}

func (c *Configuration) ToTable() *metav1.Table {
	table := &metav1.Table{
		TypeMeta: metav1.TypeMeta{
			APIVersion: metav1.SchemeGroupVersion.String(),
			Kind:       "Table",
		},
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{
				Name: "Provider",
				Type: "string",
			},
			{
				Name: "App Defaults",
				Type: "string",
			},
		},
	}

	if c.Spec.Global != nil {
		convertedArgs := argsToString(c.Spec.Global)
		row := metav1.TableRow{
			Cells: []interface{}{"GLOBAL", convertedArgs},
		}
		table.Rows = append(table.Rows, row)

	}

	for providerKey, providerDefaults := range c.Spec.Providers {
		convertedArgs := argsToString(providerDefaults)
		row := metav1.TableRow{
			Cells: []interface{}{providerKey, convertedArgs},
		}
		table.Rows = append(table.Rows, row)
	}

	return table
}

func argsToString(args map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range args {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}
