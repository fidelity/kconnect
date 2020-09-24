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

package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	errInvalidFomat = errors.New("invalid log format")
)

// Configure will configure the logging
func Configure(logLevel, logFormat string) error {
	if logLevel != "" {
		level, err := logrus.ParseLevel(strings.ToUpper(logLevel))
		if err != nil {
			return fmt.Errorf("setting log level to %s: %w", logLevel, err)
		}
		logrus.SetLevel(level)
	}

	switch strings.ToUpper(logFormat) {
	case "TEXT":
		logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	case "JSON":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		return fmt.Errorf("setting log output to %s: %w", logFormat, errInvalidFomat)
	}

	logrus.SetOutput(os.Stderr)
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	return nil
}
