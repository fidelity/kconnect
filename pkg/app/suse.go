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
	"fmt"
	"os/exec"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/prompt"
	"go.uber.org/zap"
)

func (a *App) Suse(ctx context.Context, input *UseInput) error {
	a.logger.Debug("use command")

	clusterID := *input.ClusterID
	if clusterID == "" {
		clusterID = getValue(input, "cluster-id", "input cluster ID.", true)
	}
	clusterUrl := getValue(input, "cluster-url", "input cluster url.", false)
	authority := getValue(input, "cluster-auth", "input certificate authority data.", false)
	oidcServer := getValue(input, "oidc-server", "input oidc server.", true)
	oidcClientID := getValue(input, "oidc-client-id", "input oidc client ID.", true)
	var oidcClientSecret string

	usePkce := false
	oidcUsePkce := input.ConfigSet.Get("oidc-use-pkce")
	if oidcUsePkce != nil {
		usePkce = true
	} else {
		oidcClientSecret = getValue(input, "oidc-client-secret", "input oidc client secret.", true)
	}
	// tenantID := getValue(input, true, "oidc-tenant-id", "input tenant ID.", false)
	// login := getValue(input, "login", "input login type.", false)
	// azLogin := getValue(input, "azure-kubelogin", "Use azure kubelogin.", false)

	if clusterUrl != "" {
		setCluster(clusterID, clusterUrl, authority)
	}

	// if azLogin == "true" {
	// 	setCredentials(clusterID, oidcServer, oidcUser, oidcClientID, oidcClientSecret, tenantID, login)
	// } else {
	if usePkce {
		setCredentials3(clusterID, oidcServer, oidcClientID, oidcClientID)
	} else {
		setCredentials2(clusterID, oidcServer, oidcClientID, oidcClientID, oidcClientSecret)
	}
	// }

	setContext(clusterID, oidcClientID)

	if !input.NoHistory {
		entry := historyv1alpha.NewHistoryEntry()
		entry.Spec.Alias = input.Alias
		entry.Spec.ConfigFile = input.Kubeconfig
		entry.Spec.Flags = a.filterConfig(input)
		entry.Spec.Identity = "oidc"
		entry.Spec.Flags["username"] = oidcClientID
		entry.Spec.Provider = "azure-ad"
		entry.Spec.ProviderID = clusterID

		if err := a.historyStore.Add(entry); err != nil {
			return fmt.Errorf("adding connection to history: %w", err)
		}

	}

	return nil

}

func getValue(input *UseInput, key string, msg string, required bool) (value string) {
	// if noInput {
	item := input.ConfigSet.Get(key)
	if item == nil {
		value = readUserInput(input, key, msg, required)
	} else {
		value = item.Value.(string)
		fmt.Println("Use default value for " + key + ", " + value)
	}
	// }
	if required && value == "" {
		panic("You need to input, no default value found for " + key)
	}
	return
}

func readUserInput(input *UseInput, key string, msg string, required bool) string {
	userInput, _ := prompt.Input(key, msg, false)
	if userInput == "" {
		item := input.ConfigSet.Get(key)
		if item == nil {
			if required {
				panic(key + " is neither passed nor found in config file!")
			}
		} else {
			return item.Value.(string)
		}
	} else {
		fmt.Println("Use input for " + key + ", " + userInput)
		return userInput
	}
	return ""
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

func setCredentials3(clusterID string, oidcServer string, oidcUser string, oidcClientID string) {

	zap.S().Infof("Setting up users for cluster %s", clusterID)

	output, err := exec.Command("kubectl", "config",
		"set-credentials", oidcUser,
		"--exec-api-version", "client.authentication.k8s.io/v1beta1",
		"--exec-command", "kubectl",
		"--exec-arg=oidc-login",
		"--exec-arg=get-token",
		"--exec-arg=--oidc-issuer-url="+oidcServer,
		"--exec-arg=--oidc-client-id="+oidcClientID,
		"--exec-arg=--oidc-use-pkce",
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
