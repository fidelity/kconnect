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
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	errInvalidFomat = errors.New("invalid log format")
)

// Configure will configure the logging
func Configure(logLevel, logFormat string) error {
	if logLevel != "" {
		level, err := log.ParseLevel(strings.ToUpper(logLevel))
		if err != nil {
			return fmt.Errorf("setting log level to %s: %w", logLevel, err)
		}
		log.SetLevel(level)
	}

	switch strings.ToUpper(logFormat) {
	case "TEXT":
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	case "JSON":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		return fmt.Errorf("setting log output to %s: %w", logFormat, errInvalidFomat)
	}

	return nil
}
