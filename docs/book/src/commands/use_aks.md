## kconnect use aks

Connect to the aks cluster provider and choose a cluster.

### Synopsis


Connect to aks via the configured identify provider, prompting the user to enter
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


```bash
kconnect use aks [flags]
```

### Examples

```bash
  # Discover AKS clusters using Azure AD
  kconnect use aks --idp-protocol aad

  # Discover AKS clusters using file based credentials
  export AZURE_TENANT_ID="123455"
  export AZURE_CLIENT_ID="76849"
  export AZURE_CLIENT_SECRET="supersecret"
  kconnect use aks --idp-protocol az-env

  # Reconnect to a cluster by its connection history entry alias.
  kconnect to mycluster

  # Display the user's connection history as a table.
  kconnect ls

```

### Options

```bash
      --admin                      Generate admin user kubeconfig
  -a, --alias string               Friendly name to give to give the connection
  -c, --cluster-id string          Id of the cluster to use.
      --cluster-name string        The name of the AKS cluster
  -h, --help                       help for aks
      --history-location string    Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --idp-protocol string        The idp protocol to use (e.g. saml, aad). See flags additional flags for the protocol.
  -k, --kubeconfig string          Location of the kubeconfig to use. (default "$HOME/.kube/config")
      --max-history int            Sets the maximum number of history items to keep (default 100)
  -n, --namespace string           Sets namespace for context in kubeconfig
      --no-history                 If set to true then no history entry will be written
      --password string            The password to use for authentication
  -r, --resource-group string      The Azure resource group to use
      --set-current                Sets the current context in the kubeconfig to the selected cluster (default true)
      --subscription-id string     The Azure subscription to use (specified by ID)
      --subscription-name string   The Azure subscription to use (specified by name)
      --username string            The username used for authentication
```

### Options inherited from parent commands

```bash
      --config string      Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --no-input           Explicitly disable interactivity when running in a terminal
      --no-version-check   If set to true kconnect will not check for a newer version
  -v, --verbosity int      Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### IDP Protocol Options

#### AAD Options

Use `--idp-protocol=aad`

```bash
      --aad-host string       The AAD host to use (default "login.microsoftonline.com")
      --client-id string      The azure ad client id (default "04b07795-8ddb-461a-bbee-02f9e1bf7b46")
      --idp-protocol string   The idp protocol to use (e.g. saml). Each protocol has its own flags.
      --password string       The password to use for authentication
  -t, --tenant-id string      The azure tenant id
      --username string       The username used for authentication
```

#### AZ-ENV Options

Use `--idp-protocol=az-env`

```bash
      --use-file   Use file based authorization
```

### SEE ALSO

* [kconnect use](use.md)	 - Connect to a Kubernetes cluster provider and cluster.


> NOTE: this page is auto-generated from the cobra commands
