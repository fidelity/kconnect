package kubeconfig

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
)

const (
	RecommendedConfigPathEnvVar = clientcmd.RecommendedConfigPathEnvVar
)
type Config struct {
	OverrideConfig bool
}


func DefaultKubeCfgLoc() string {
	if env := os.Getenv(RecommendedConfigPathEnvVar); len(env) > 0 {
		return env
	}
	return clientcmd.RecommendedHomeFile
}

// KubeCfgClient is a client that can be used to build, update kubeconfig file.
// path refer to the location of the kubeconfig file.
type KubeCfgClient struct {
	path string // do we need this or use cfgAccess.GetDefaultFileName()?
	cfgAccess clientcmd.ConfigAccess
	RelativizePath bool // Defaults to true, change it only if you need to. Refer https://github.com/kubernetes/client-go/blob/master/tools/clientcmd/config.go#L165
}

func NewClient() *KubeCfgClient {
	return &KubeCfgClient{
		path: DefaultKubeCfgLoc(),
		cfgAccess: getConfigAccess(DefaultKubeCfgLoc()),
		RelativizePath: true,
	}
}

func (kcli *KubeCfgClient) SetPath( f string)  {
	kcli.path = f
	kcli.cfgAccess = getConfigAccess(f)
}

func getConfigAccess(explicitPath string) clientcmd.ConfigAccess {
	pathOptions := clientcmd.NewDefaultPathOptions()
	if explicitPath != "" && explicitPath != DefaultKubeCfgLoc() {
		pathOptions.LoadingRules.ExplicitPath = explicitPath
	}

	return interface{}(pathOptions).(clientcmd.ConfigAccess)
}

func (kcli *KubeCfgClient) ConfigAccess() clientcmd.ConfigAccess {
	return kcli.cfgAccess
}

func (kc *KubeCfgClient) GetConfig() (*clientcmdapi.Config, error) {
	return kc.cfgAccess.GetStartingConfig()
}

func (kc *KubeCfgClient) GetConfigAsBytes() ([]byte, error) {
	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}
	return clientcmd.Write(*currentConfig)
}

type ModifyConfigOptions func(*Config)

func WithOverride(c *Config)  {
	c.OverrideConfig = true
}

func (kc *KubeCfgClient) ModifyConfig(newConfig *clientcmdapi.Config, opts ...ModifyConfigOptions) error {

	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}
	if err := clientcmd.Validate(*newConfig) ; err != nil {
		return fmt.Errorf("error while trying to validate given config %w", err)
	}

	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}

	if !config.OverrideConfig {
		for k, v := range newConfig.Clusters {
			currentConfig.Clusters[k] = v
		}
		for k, v := range newConfig.AuthInfos {
			currentConfig.AuthInfos[k] = v
		}
		for k, v := range newConfig.Contexts {
			currentConfig.Contexts[k] = v
		}
		for k, v := range newConfig.Extensions {
			currentConfig.Extensions[k] = v
		}
	}

	err = clientcmd.ModifyConfig(kc.cfgAccess, *currentConfig, kc.RelativizePath)
	if err !=nil {
		fmt.Errorf("unable to modify kubeconfig at %s %w",kc.path, err)
	}

	return nil
}

func (kc *KubeCfgClient) SetCurrentContext(ctx string) error {
	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}
	currentConfig.CurrentContext = ctx
	return kc.ModifyConfig(currentConfig)
}

func (kc *KubeCfgClient) AddCluster(name string, c *clientcmdapi.Cluster) error {
	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}
	currentConfig.Clusters[name] = c
	return kc.ModifyConfig(currentConfig)
}

func (kc *KubeCfgClient) RemoveCluster(names ...string) error {
	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}
	for _, name := range names {
		delete(currentConfig.Clusters,name)
	}
	return kc.ModifyConfig(currentConfig, WithOverride)
}

func (kc *KubeCfgClient) AddAuthInfo(name string, a *clientcmdapi.AuthInfo) error {
	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}
	currentConfig.AuthInfos[name] = a
	return kc.ModifyConfig(currentConfig)
}

func (kc *KubeCfgClient) RemoveAuthInfo(names ...string) error {
	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}
	for _, name := range names {
		delete(currentConfig.AuthInfos,name)
	}
	return kc.ModifyConfig(currentConfig, WithOverride)
}

func (kc *KubeCfgClient) AddContext(name string, c *clientcmdapi.Context) error {
	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}
	currentConfig.Contexts[name] = c
	return kc.ModifyConfig(currentConfig)
}

func (kc *KubeCfgClient) RemoveContext(ctx ...string) error {
	currentConfig, err := kc.cfgAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("unable to read kubeconfig file at %s %w", kc.path, err)
	}

	for _, c := range ctx {
		ctx, ok := currentConfig.Contexts[c]
		if !ok {
			continue
		}
		delete(currentConfig.Clusters,ctx.Cluster)
		delete(currentConfig.AuthInfos,ctx.AuthInfo)
		delete(currentConfig.Contexts,c)
	}

	return kc.ModifyConfig(currentConfig, WithOverride)
}
