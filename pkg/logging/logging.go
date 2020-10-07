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

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	thirdPartyVerboseLevel = 9
)

var (
	ErrInvalidFomat = errors.New("invalid log format")
)

// Configure will configure the logging for kconnect and the dependent saml2aws package
func Configure(verbosity int) error {

	configureLogrus(verbosity)

	if err := configureZap(verbosity); err != nil {
		return fmt.Errorf("configuring zap logging: %w", err)
	}

	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	return nil
}

// configureLogrus will configure logrus which is used by saml2aws
func configureLogrus(verbosity int) {
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	logrus.SetOutput(os.Stderr)

	if verbosity >= thirdPartyVerboseLevel {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func configureZap(verbosity int) error {
	logConfig := zap.NewProductionConfig()

	logConfig.Encoding = "console"
	logConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	logConfig.EncoderConfig.TimeKey = ""
	logConfig.EncoderConfig.CallerKey = ""
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if verbosity > 0 {
		logConfig.Level.SetLevel(zap.DebugLevel)
	} else {
		logConfig.Level.SetLevel(zap.InfoLevel)
	}

	loggerMgr, err := logConfig.Build()
	if err != nil {
		return fmt.Errorf("building zap logger: %w", err)
	}
	zap.ReplaceGlobals(loggerMgr)

	return loggerMgr.Sync()
}
