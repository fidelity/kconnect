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
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Context represents a context for the providers
type Context struct {
	context.Context

	Command *cobra.Command
	Logger  *logrus.Entry
}

// NewContext creates a new context
func NewContext(ctx context.Context, cmd *cobra.Command, logger *logrus.Entry) *Context {
	c := &Context{
		Context: ctx,
		Command: cmd,
		Logger:  logger,
	}

	if c.Context == nil {
		c.Context = context.Background()
	}

	return c
}
