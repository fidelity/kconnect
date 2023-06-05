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
	"os/exec"

	"go.uber.org/zap"
)

func (a *App) Suse(ctx context.Context, input *UseInput) error {
	a.logger.Debug("use command")

	clusterID := *input.ClusterID

	clusterUrl := input.ConfigSet.Get("cluster-url")
	if clusterUrl != nil {
		setCluster(clusterID, input)
	}

	setCredentials(clusterID, input)

	setContext(clusterID, input)

	return nil

}

func setContext(clusterID string, input *UseInput) {

	oidcUser := input.ConfigSet.Get("oidc-user").Value.(string)
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

func setCredentials(clusterID string, input *UseInput) {

	zap.S().Infof("Setting up users for cluster %s", clusterID)

	oidcServer := input.ConfigSet.Get("oidc-server").Value.(string)
	oidcUser := input.ConfigSet.Get("oidc-user").Value.(string)
	oidcClientID := input.ConfigSet.Get("oidc-client-id").Value.(string)
	oidcClientSecret := input.ConfigSet.Get("oidc-client-secret").Value.(string)
	tenantID := input.ConfigSet.Get("oidc-tenant-id").Value.(string)

	output, err := exec.Command("kubectl", "config",
		"set-credentials", oidcUser,
		"--exec-api-version", "client.authentication.k8s.io/v1beta1",
		"--exec-command", "kubelogin",
		"--exec-arg=get-token",
		"--exec-arg=--login=interactive",
		"--exec-arg=--server-id="+oidcServer,
		"--exec-arg=--client-id="+oidcClientID,
		"--exec-arg=--client-secret="+oidcClientSecret,
		"--exec-arg=--tenant-id="+tenantID).Output()

	if err != nil {
		zap.S().Errorf("Failed to setup user %s. Error: %s", oidcUser, err)
	} else {
		zap.S().Infof("Setup user successfully. Output: %s", output)
	}

}

func setCluster(clusterID string, input *UseInput) {

	zap.S().Infof("Config cluster %s", clusterID)

	clusterUrl := input.ConfigSet.Get("cluster-url").Value.(string)
	authority := input.ConfigSet.Get("cluster-auth").Value.(string)

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
