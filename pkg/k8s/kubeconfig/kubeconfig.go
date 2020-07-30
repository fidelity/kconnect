package kubeconfig

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	logger = logrus.WithField("package", "kubeconfig")
)

// Write will write the kubeconfig to the specified file. If there
// is an existing kubeconfig it will be merged
func Write(path string, clusterConfig *api.Config) error {
	logger.Infof("Writing kubeconfig to %s", path)

	pathOptions := clientcmd.NewDefaultPathOptions()
	if path != "" {
		pathOptions.LoadingRules.ExplicitPath = path
	}

	existingConfig, err := pathOptions.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("getting existing kubeconfig: %w", err)
	}

	if err := mergo.Merge(existingConfig, clusterConfig); err != nil {
		return fmt.Errorf("merging cluster config: %w", err)
	}

	if err := clientcmd.ModifyConfig(pathOptions, *existingConfig, true); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}

	return nil
}
