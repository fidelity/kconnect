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
	"errors"
	"fmt"
	"strings"
)

var (
	ErrConfigExistsAlready = errors.New("config item with same name already exists in set")
	ErrConfigNotFound      = errors.New("configuration item not found")
	ErrUnknownItemType     = errors.New("unknown item type for config item")
)

// Item represents a configuration item
type Item struct {
	Name              string
	Shorthand         string
	Type              ItemType
	Description       string
	Sensitive         bool
	ResolutionPrompt  string
	Value             any
	DefaultValue      any
	Required          bool
	Hidden            bool
	Deprecated        bool
	DeprecatedMessage string
	HistoryIgnore     bool
}

func (i *Item) HasValue() bool {
	if i == nil {
		return false
	}

	if i.Value == nil {
		return false
	}

	if i.Type == ItemTypeString {
		return i.Value.(string) != ""
	}

	return true
}

type ItemType string

var (
	ItemTypeString = ItemType("string")
	ItemTypeInt    = ItemType("int")
	ItemTypeBool   = ItemType("bool")
)

type ConfigurationSet interface {
	Get(name string) *Item
	GetAll() []*Item
	Exists(name string) bool
	ExistsWithValue(name string) bool
	ValueIsList(name string) bool
	ValueString(name string) string
	Add(item *Item) error
	AddSet(set ConfigurationSet) error
	SetSensitive(name string) error
	SetHistoryIgnore(name string) error
	SetRequired(name string) error
	SetHidden(name string) error
	SetDeprecated(name string, message string) error
	SetValue(name string, value any) error
	SetShort(name string, shorthand string) error

	String(name string, defaultValue string, description string) (*Item, error)
	Int(name string, defaultValue int, description string) (*Item, error)
	Bool(name string, defaultValue bool, description string) (*Item, error)
}

func NewConfigurationSet() ConfigurationSet {
	return &configSet{
		config: make(map[string]*Item),
	}
}

type configSet struct {
	config map[string]*Item
}

func (s *configSet) Exists(name string) bool {
	return s.Get(name) != nil
}

func (s *configSet) ExistsWithValue(name string) bool {
	item := s.Get(name)
	if item == nil {
		return false
	}

	if !item.HasValue() {
		return false
	}

	val := item.Value.(string)

	return !strings.HasPrefix(val, ListPrefix)
}

func (s *configSet) ValueIsList(name string) bool {
	item := s.Get(name)
	if item == nil {
		return false
	}

	if !item.HasValue() {
		return false
	}

	val := s.ValueString(name)

	return strings.HasPrefix(val, ListPrefix)
}

func (s *configSet) ValueString(name string) string {
	item := s.Get(name)
	if item == nil {
		return ""
	}

	switch item.Type {
	case ItemTypeString:
		return item.Value.(string)
	case ItemTypeInt:
		intVal := item.Value.(int)
		return fmt.Sprintf("%d", intVal)
	case ItemTypeBool:
		boolVal := item.Value.(bool)
		return fmt.Sprintf("%t", boolVal)
	default:
		return ""
	}
}

func (s *configSet) Get(name string) *Item {
	return s.config[name]
}

func (s *configSet) GetAll() []*Item {
	items := make([]*Item, 0, len(s.config))
	for _, item := range s.config {
		items = append(items, item)
	}

	return items
}

func (s *configSet) Add(item *Item) error {
	_, exists := s.config[item.Name]
	if exists {
		return ErrConfigExistsAlready
	}

	s.config[item.Name] = item

	return nil
}

func (s *configSet) AddSet(setToAdd ConfigurationSet) error {
	if setToAdd == nil {
		return nil
	}

	for _, item := range setToAdd.GetAll() {
		if !s.Exists(item.Name) {
			if err := s.Add(item); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *configSet) SetSensitive(name string) error {
	item := s.Get(name)
	if item == nil {
		return ErrConfigNotFound
	}

	item.Sensitive = true

	return nil
}

func (s *configSet) SetHistoryIgnore(name string) error {
	item := s.Get(name)
	if item == nil {
		return ErrConfigNotFound
	}

	item.HistoryIgnore = true

	return nil
}

func (s *configSet) SetRequired(name string) error {
	item := s.Get(name)
	if item == nil {
		return ErrConfigNotFound
	}

	item.Required = true

	return nil
}

func (s *configSet) SetHidden(name string) error {
	item := s.Get(name)
	if item == nil {
		return ErrConfigNotFound
	}

	item.Hidden = true

	return nil
}

func (s *configSet) SetDeprecated(name string, message string) error {
	item := s.Get(name)
	if item == nil {
		return ErrConfigNotFound
	}

	item.Deprecated = true
	item.DeprecatedMessage = message

	return nil
}

func (s *configSet) SetValue(name string, value any) error {
	item := s.Get(name)
	if item == nil {
		return ErrConfigNotFound
	}

	item.Value = value

	return nil
}

func (s *configSet) SetShort(name string, shorthand string) error {
	item := s.Get(name)
	if item == nil {
		return ErrConfigNotFound
	}

	item.Shorthand = shorthand

	return nil
}

func (s *configSet) String(name string, defaultValue string, description string) (*Item, error) {
	item := &Item{
		Name:         name,
		Type:         ItemTypeString,
		DefaultValue: defaultValue,
		Description:  description,
	}

	if err := s.Add(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *configSet) Int(name string, defaultValue int, description string) (*Item, error) {
	item := &Item{
		Name:         name,
		Type:         ItemTypeInt,
		DefaultValue: defaultValue,
		Description:  description,
	}

	if err := s.Add(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *configSet) Bool(name string, defaultValue bool, description string) (*Item, error) {
	item := &Item{
		Name:         name,
		Type:         ItemTypeBool,
		DefaultValue: defaultValue,
		Description:  description,
	}

	if err := s.Add(item); err != nil {
		return nil, err
	}

	return item, nil
}
