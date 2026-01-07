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

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"golang.org/x/mod/semver"
)

const minKubectlVersion string = "v1.17.0"

var ErrorTooLowKubectlVersion = errors.New("kubectl version is too low")

type KubectlVersion struct {
	ClientVersion struct {
		Major      string `json:"major"`
		Minor      string `json:"minor"`
		GitVersion string `json:"gitVersion"`
	} `json:"clientVersion"`
}

func CheckKubectlPrereq() error {
	cmd := exec.Command("kubectl", "version", "--output=json", "--client=true")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running kubectl version: %w", err)
	}

	var kubectlVersion KubectlVersion

	err = json.Unmarshal(output, &kubectlVersion)
	if err != nil {
		return fmt.Errorf("error checking for kubectl version: %w", err)
	}

	if semver.Compare(kubectlVersion.ClientVersion.GitVersion, minKubectlVersion) < 0 {
		return ErrorTooLowKubectlVersion
	}

	return nil
}

func CheckAWSIAMAuthPrereq() error {
	cmd := exec.Command("aws-iam-authenticator")

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error finding aws-iam-authenticator: %w", err)
	}

	return nil
}

func CheckKubeloginPrereq() error {
	cmd := exec.Command("kubelogin")

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error finding kubelogin: %w", err)
	}

	return nil
}

func CheckKubectlOidcLoginPrereq() error {
	cmd := exec.Command("kubectl", "oidc-login", "version")

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error finding kubectl oidc-login: %w", err)
	}

	return nil
}
