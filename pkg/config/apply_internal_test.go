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
	"testing"
)

func TestSetItemValueTypes(t *testing.T) {
	t.Run("int is stored as an int", func(t *testing.T) {
		item := &Item{Name: "session-duration", Type: ItemTypeInt}

		if err := setItemValue(item, "3600"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Readers of the value assert that it is an int, which panics for any other type
		value, ok := item.Value.(int)
		if !ok {
			t.Fatalf("expected an int but got %T", item.Value)
		}

		if value != 3600 {
			t.Fatalf("expected 3600 but got %d", value)
		}
	})

	t.Run("invalid int is an error", func(t *testing.T) {
		item := &Item{Name: "session-duration", Type: ItemTypeInt}

		if err := setItemValue(item, "an hour"); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("string", func(t *testing.T) {
		item := &Item{Name: "region", Type: ItemTypeString}

		if err := setItemValue(item, "us-east-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if item.Value != "us-east-1" {
			t.Fatalf("unexpected value %v", item.Value)
		}
	})
}

func TestIntFromConfigCanBeReadBack(t *testing.T) {
	cs := NewConfigurationSet()

	if _, err := cs.Int("session-duration", 900, "duration"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := setItemValue(cs.Get("session-duration"), "3600"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ValueString asserts the type of the value
	if value := cs.ValueString("session-duration"); value != "3600" {
		t.Fatalf("expected 3600 but got %q", value)
	}
}
