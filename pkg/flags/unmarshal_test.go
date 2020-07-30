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

package flags_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"

	"github.com/fidelity/kconnect/pkg/flags"
)

func TestUnmarshallStruct(t *testing.T) {
	g := NewWithT(t)

	type testStruct struct {
		Name     string  `flag:"name"`
		Number   int     `flag:"num"`
		Loglevel *string `flag:"log-level"`
		Weight   float64 `flag:"weight"`
	}
	input := &testStruct{}

	flagset := pflag.NewFlagSet("", pflag.ContinueOnError)
	flagset.String("name", "", "the name flag")
	flagset.Int("num", 0, "the number flag")
	flagset.String("log-level", "INFO", "the logging level")
	flagset.Float32("weight", 0.0, "the weight of something")

	args := []string{"--name", "test", "--num", "99", "--weight", "10.2"}
	err := flagset.Parse(args)
	g.Expect(err).To(BeNil())

	err = flags.Unmarshal(flagset, input)
	g.Expect(err).To(BeNil())
	g.Expect(input.Name).To(Equal("test"))
	g.Expect(input.Number).To(Equal(99))
	g.Expect(input.Loglevel).NotTo(BeNil())
	g.Expect(*input.Loglevel).To(Equal("INFO"))
	g.Expect(input.Weight).To(Equal(10.2))
}

func TestUnmarshallStructWithStructField(t *testing.T) {
	g := NewWithT(t)

	type subStruct struct {
		ReadOnly bool `flag:"read-only"`
	}

	type testStruct struct {
		SubStruct subStruct
	}

	input := &testStruct{}

	flagset := pflag.NewFlagSet("", pflag.ContinueOnError)
	flagset.Bool("read-only", false, "indicates that we are read-only")

	args := []string{"--read-only", "true"}
	err := flagset.Parse(args)
	g.Expect(err).To(BeNil())

	err = flags.Unmarshal(flagset, input)
	g.Expect(err).To(BeNil())
	g.Expect(input.SubStruct.ReadOnly).To(BeTrue())
}

func TestUnmarshallStructWithEmbeddedStruct(t *testing.T) {
	g := NewWithT(t)

	type BaseStruct struct {
		LogLevel string `flag:"log-level"`
	}

	type TestStruct struct {
		BaseStruct
		Name *string `flag:"name"`
	}

	input := &TestStruct{}

	flagset := pflag.NewFlagSet("", pflag.ContinueOnError)
	flagset.String("log-level", "INFO", "whats the logging level")
	flagset.String("name", "", "the name of the thing")

	args := []string{"--name", "testname", "--log-level", "DEBUG"}
	err := flagset.Parse(args)
	g.Expect(err).To(BeNil())

	err = flags.Unmarshal(flagset, input)
	g.Expect(err).To(BeNil())
	g.Expect(input.Name).NotTo(BeNil())
	g.Expect(*input.Name).To(Equal("testname"))
	g.Expect(input.LogLevel).To(Equal("DEBUG"))
}
