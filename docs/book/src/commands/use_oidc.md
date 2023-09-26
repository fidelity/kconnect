## kconnect use oidc

Connect to the oidc cluster provider and choose a cluster.

### Synopsis


Connect to oidc via the configured identify provider, prompting the user to enter
or choose connection settings and a target cluster once connected.

The kconnect tool generates a kubectl configuration context with a fresh access
token to connect to the chosen cluster and adds a connection history entry to
store the chosen connection settings.  If given an alias name, kconnect will add
a user-friendly alias to the new connection history entry.

The user can then reconnect to the provider with the settings stored in the
connection history entry using the kconnect to command and the connection history
entry ID or alias.  When the user reconnects using a connection history entry,
kconnect regenerates the kubectl configuration context and refreshes their access
token.

* Note: kconnect use oidc requires kube-oidc-login and rename to kubectl-oidc_login.
  [kube-oidc-login](https://github.com/int128/kubelogin)


```bash
kconnect use oidc [flags]
```

### Examples

```bash

  # Setup cluster for oidc protocol using default config file
  kconnect use oidc

  # Setup cluster for oidc protocol using config url
  kconnect use oidc --config-url https://localhost:8080

  # Setup cluster and add an alias to its connection history entry
  kconnect use oidc --alias mycluster
  
  # Reconnect to a cluster by its connection history entry alias.
  kconnect to mycluster

  # Display the user's connection history as a table.
  kconnect ls

```

### Options

```bash
  -a, --alias string                Friendly name to give to give the connection
      --ca-cert string              ca cert for configuration url
      --cluster-auth string         cluster auth data
  -c, --cluster-id string           Id of the cluster to use.
      --cluster-url string          cluster api server endpoint
      --config-url string           configuration endpoint
  -h, --help                        help for oidc
      --history-location string     Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --idp-protocol string         The idp protocol to use (e.g. saml, aad). See flags additional flags for the protocol.
  -k, --kubeconfig string           Location of the kubeconfig to use. (default "$HOME/.kube/config")
      --max-history int             Sets the maximum number of history items to keep (default 100)
  -n, --namespace string            Sets namespace for context in kubeconfig
      --no-history                  If set to true then no history entry will be written
      --oidc-client-id string       oidc client id
      --oidc-client-secret string   oidc client secret
      --oidc-server string          oidc server url
      --oidc-use-pkce string        if use pkce
      --password string             The password to use for authentication
      --set-current                 Sets the current context in the kubeconfig to the selected cluster (default true)
      --skip-oidc-ssl string        flag to skip ssl for calling oidc server
      --skip-ssl string             flag to skip ssl for calling config url
      --username string             The username used for authentication
```

### Options inherited from parent commands

```bash
      --config string      Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --no-input           Explicitly disable interactivity when running in a terminal
      --no-version-check   If set to true kconnect will not check for a newer version
  -v, --verbosity int      Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### IDP Protocol Options

#### OIDC Options

Use `--idp-protocol=oidc`

```bash
      --ca-cert string              ca cert for configuration url
      --cluster-auth string         cluster auth data
      --cluster-url string          cluster api server endpoint
      --config-url string           configuration endpoint
      --oidc-client-id string       oidc client id
      --oidc-client-secret string   oidc client secret
      --oidc-server string          oidc server url
      --oidc-use-pkce string        if use pkce
      --skip-oidc-ssl string        flag to skip ssl for calling oidc server
      --skip-ssl string             flag to skip ssl for calling config url
```

### SEE ALSO

* [kconnect use](use.md)	 - Connect to a Kubernetes cluster provider and cluster.


> NOTE: this page is auto-generated from the cobra commands
