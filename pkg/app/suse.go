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

package app

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/fidelity/kconnect/pkg/prompt"
	"go.uber.org/zap"
)

type eks struct {
	ClusterID        string `json:"cluster-id"`
	OidcServer       string `json:"oidc-server"`
	OidcClientId     string `json:"oidc-client-id"`
	OidcClientSecret string `json:"oidc-client-secret"`
	ClusterUrl       string `json:"cluster-url"`
	ClusterAuth      string `json:"cluster-auth"`
	LoginType        string `json:"login-type"`
}

func (a *App) Suse(ctx context.Context, input *UseInput) error {
	a.logger.Debug("use command")

	config := getValue(input, true, "config", "input config file path", true)

	res, err := a.httpClient.Get(config, nil)
	if err != nil {
		panic(err)
	}

	eks := &eks{}
	if err := json.Unmarshal([]byte(res.Body()), eks); err != nil {
		return errors.New("bad payload")
	}

	clusterID := eks.ClusterID
	if clusterID == "" {
		panic("cluster ID is not set!")
	}

	clusterUrl := eks.ClusterUrl
	authority := eks.ClusterAuth
	oidcServer := eks.OidcServer

	oidcClientID := eks.OidcClientId
	oidcClientSecret := eks.OidcClientSecret

	// login := ""
	// oidcUser := ""
	// tenantID := ""
	// azLogin := getValue(input, true, "azure-kubelogin", "Use azure kubelogin.", false)
	// if login == "" {
	// 	login = "devicecode"
	// }

	if clusterUrl != "" {
		setCluster(clusterID, clusterUrl, authority)
	}

	// if azLogin == "true" {
	// 	setCredentials(clusterID, oidcServer, oidcUser, oidcClientID, oidcClientSecret, tenantID, login)
	// } else {
	setCredentials2(clusterID, oidcServer, oidcClientID, oidcClientID, oidcClientSecret)
	// }

	setContext(clusterID, oidcClientID)

	return nil

}

func getValue(input *UseInput, noInput bool, key string, msg string, required bool) (value string) {
	if noInput {
		item := input.ConfigSet.Get(key)
		if item == nil {
			if required {
				panic(key + " is neither passed nor found in config file!")
			}
		} else {
			value = item.Value.(string)
		}
	} else {
		userInput, _ := prompt.Input(key, msg, false)
		if userInput == "" {
			item := input.ConfigSet.Get(key)
			if item == nil {
				if required {
					panic(key + " is neither passed nor found in config file!")
				}
			} else {
				value = item.Value.(string)
			}
		} else {
			value = userInput
		}
	}
	return
}

func setContext(clusterID string, oidcUser string) {

	zap.S().Infof("Set context for cluster %s, user %s", clusterID, oidcUser)

	context := oidcUser + "@" + clusterID

	output, err := exec.Command("Kubectl", "config",
		"set-context", context,
		"--cluster", clusterID,
		"--user", oidcUser).Output()

	if err != nil {
		zap.S().Errorf("Failed to setup context %s. Error: %s", oidcUser, err)
	} else {
		zap.S().Infof("Successfully. Output: %s", output)
	}

	output, err = exec.Command("Kubectl", "config",
		"use-context", context).Output()

	if err != nil {
		zap.S().Errorf("Failed to current context %s. Error: %s", oidcUser, err)
	} else {
		zap.S().Infof("Successfully. Output: %s", output)
	}

}

func setCredentials2(clusterID string, oidcServer string, oidcUser string, oidcClientID string,
	oidcClientSecret string) {

	zap.S().Infof("Setting up users for cluster %s", clusterID)

	output, err := exec.Command("kubectl", "config",
		"set-credentials", oidcUser,
		"--exec-api-version", "client.authentication.k8s.io/v1beta1",
		"--exec-command", "kubectl",
		"--exec-arg=oidc-login",
		"--exec-arg=get-token",
		"--exec-arg=--oidc-issuer-url="+oidcServer,
		"--exec-arg=--oidc-client-id="+oidcClientID,
		"--exec-arg=--oidc-client-secret="+oidcClientSecret,
		"--exec-arg=--insecure-skip-tls-verify").Output()

	if err != nil {
		zap.S().Errorf("Failed to setup user %s. Error: %s", oidcUser, err)
	} else {
		zap.S().Infof("Setup user successfully. Output: %s", output)
	}

}

// func setCredentials(clusterID string, oidcServer string, oidcUser string, oidcClientID string,
// 	oidcClientSecret string, tenantID string, login string) {

// 	zap.S().Infof("Setting up users for cluster %s", clusterID)

// 	output, err := exec.Command("kubectl", "config",
// 		"set-credentials", oidcUser,
// 		"--exec-api-version", "client.authentication.k8s.io/v1beta1",
// 		"--exec-command", "kubelogin",
// 		"--exec-arg=get-token",
// 		"--exec-arg=--login="+login,
// 		"--exec-arg=--server-id="+oidcServer,
// 		"--exec-arg=--client-id="+oidcClientID,
// 		"--exec-arg=--client-secret="+oidcClientSecret,
// 		"--exec-arg=--tenant-id="+tenantID).Output()

// 	if err != nil {
// 		zap.S().Errorf("Failed to setup user %s. Error: %s", oidcUser, err)
// 	} else {
// 		zap.S().Infof("Setup user successfully. Output: %s", output)
// 	}

// }

func setCluster(clusterID string, clusterUrl string, authority string) {

	zap.S().Infof("Config cluster %s", clusterID)

	output, err := exec.Command("Kubectl", "config",
		"set-cluster", clusterID,
		"--server", clusterUrl).Output()

	if err != nil {
		zap.S().Errorf("Failed to config cluster %s server. Error: %s", clusterID, err)
	} else {
		zap.S().Infof("Successfully. Output: %s", output)
	}

	output, err = exec.Command("Kubectl", "config",
		"set", "clusters."+clusterID+".certificate-authority-data", authority).Output()

	if err != nil {
		zap.S().Errorf("Failed to config cluster %s authority-data. Error: %s", clusterID, err)
	} else {
		zap.S().Infof("Successfully. Output: %s", output)
	}

}
