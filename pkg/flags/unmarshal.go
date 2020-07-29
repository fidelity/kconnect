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

package flags

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

const (
	// TagName defines the name of struct tag to look for
	TagName = "flag"
)

// Unmarshal will decode the flagset into the out interface
func Unmarshal(flagset *pflag.FlagSet, out interface{}, opts ...BinderOption) error {
	b := newFlagBinder(opts...)
	return b.Unmarshal(flagset, out)
}

type BinderOption func(*flagBinder)

func IgnoreFlagNotFound() BinderOption {
	return func(b *flagBinder) {
		b.IgnoreNotFound = true
	}
}

func newFlagBinder(opts ...BinderOption) *flagBinder {
	b := &flagBinder{
		IgnoreNotFound: false,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

type flagBinder struct {
	IgnoreNotFound bool
}

func (b *flagBinder) Unmarshal(flagset *pflag.FlagSet, out interface{}) error {
	rv := reflect.ValueOf(out)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("must be a pointer, got kind %s", rv.Kind())
	}
	if rv.IsNil() {
		return errors.New("cannot unmarshall into nil structure")
	}

	val := reflect.Indirect(rv)
	t := val.Type()
	if val.Kind() != reflect.Struct {
		return errors.New("must unmarshall into struct")
	}

	// Loop through the struct fields and see if there
	// is a matching flag
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)

		fieldV := val.Field(i)
		fieldT := fieldV.Type()

		if !fieldV.CanAddr() {
			return fmt.Errorf("field cannot be set as its not addressable")
		}

		isStruct := fieldT.Kind() == reflect.Struct

		if isStruct {
			fieldV = fieldV.Addr()
			if err := b.Unmarshal(flagset, fieldV.Interface()); err != nil {
				return err
			}
			continue
		}

		tagStr, tagExists := field.Tag.Lookup(TagName)
		if !tagExists {
			continue
		}

		if fieldT.Kind() == reflect.Ptr && fieldV.IsNil() {
			fieldV.Set(reflect.New(fieldT.Elem()))
		}

		flagName := getFlagNameFromTag(tagStr)

		flag := flagset.Lookup(flagName)
		if flag == nil {
			if b.IgnoreNotFound {
				continue
			}
			return fmt.Errorf("no flag named %s found", flagName)
		}
		if flag.Value == nil {
			//TODO: if field is pointer set to nil
			continue
		}

		if err := unmarshallFlag(flag, fieldV); err != nil {
			return err
		}
	}

	return nil
}

func unmarshallFlag(flag *pflag.Flag, out reflect.Value) error {
	fieldT := out.Type()
	fmt.Println(fieldT.Kind())

	flagValueStr := flag.Value.String()

	switch fieldT.Kind() {
	case reflect.String:
		out.SetString(flagValueStr)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(flagValueStr, 10, 64)
		if err != nil {
			return err
		}
		out.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ui, err := strconv.ParseUint(flagValueStr, 10, 64)
		if err != nil {
			return err
		}
		out.SetUint(ui)
	case reflect.Bool:
		b, err := strconv.ParseBool(flagValueStr)
		if err != nil {
			return err
		}
		out.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(flagValueStr, 64)
		if err != nil {
			return err
		}
		out.SetFloat(f)
	case reflect.Ptr:
		out = out.Elem()
		return unmarshallFlag(flag, out)
	default:
		return fmt.Errorf("can'y unmarshall to field of type: %s", fieldT.Name())
	}

	return nil
}

func getFlagNameFromTag(tag string) string {
	if tag == "" {
		return ""
	}

	// NOTE: the split is pointless at the moment
	// as there is only 1 part of the tag but
	// it will be expanded in the future
	parts := strings.Split(tag, ",")
	return parts[0]
}
