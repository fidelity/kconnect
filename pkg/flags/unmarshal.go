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

var (
	ErrMustBePtr           = errors.New("must marshall into a pointer")
	ErrNilStruct           = errors.New("cannot unmarshall flags into nil structure")
	ErrRequireStruct       = errors.New(("must unmarshall into struct"))
	ErrFieldNotAddressable = errors.New("field cannot be set as its not addressable")
	ErrFlagNotFound        = errors.New("flag not found in flagset")
	ErrUnsupportedType     = errors.New("type not supported for unmarshalling")
)

// Unmarshal will decode the flagset into the out interface
func Unmarshal(flagset *pflag.FlagSet, out interface{}, opts ...BinderOption) error {
	b := newFlagBinder(opts...)
	return b.Unmarshal(flagset, out)
}

// BinderOption defines a functional option for creations of the flags binder
type BinderOption func(*flagBinder)

// IgnoreFlagNotFound is an option that specifies that if a flag doesn't
// exist then no error should be returned
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

// Unmarshal will decode the flagset into the out structure
func (b *flagBinder) Unmarshal(flagset *pflag.FlagSet, out interface{}) error {
	rv := reflect.ValueOf(out)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("out interface is kind %s: %w", rv.Kind(), ErrMustBePtr)
	}
	if rv.IsNil() {
		return ErrNilStruct
	}

	val := reflect.Indirect(rv)
	t := val.Type()
	if val.Kind() != reflect.Struct {
		return ErrRequireStruct
	}

	// Loop through the struct fields and see if there
	// is a matching flag
	for i := range val.NumField() {
		field := t.Field(i)

		fieldV := val.Field(i)
		fieldT := fieldV.Type()

		if !fieldV.CanAddr() {
			return ErrFieldNotAddressable
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
			return fmt.Errorf("failed looking up flag %s: %w", flagName, ErrFlagNotFound)
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

	flagValueStr := flag.Value.String()

	switch fieldT.Kind() { //nolint: exhaustive
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
		return fmt.Errorf("failed unmarshalling to field of type %s: %w", fieldT.Name(), ErrUnsupportedType)
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
