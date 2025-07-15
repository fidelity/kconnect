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

package time

import (
	"time"

	"github.com/fidelity/kconnect/pkg/aws/awsconfig"
	"gopkg.in/ini.v1"
)

func GetExpireTimeFromAWSCredentials(profileName string) (time.Time, error) {
	path, err := awsconfig.LocateConfigFile()
	if err != nil {
		return time.Time{}, err
	}

	cfg, err := ini.Load(path)
	if err != nil {
		return time.Time{}, err
	}

	tokenTime := cfg.Section(profileName).Key("x_security_token_expires").String()

	return time.Parse(time.RFC3339, tokenTime)
}

func GetRemainingTime(expiresAt time.Time) string {
	timeRemaining := time.Until(expiresAt)

	timeRemaining = timeRemaining.Round(time.Second)
	if timeRemaining.Seconds() < 0 {
		timeRemaining = 0
	}

	return timeRemaining.String()
}
