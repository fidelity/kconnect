## kconnect use eks

Connect to the eks cluster provider and choose a cluster.

### Synopsis


Connect to eks via the configured identify provider, prompting the user to enter 
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

* Note: kconnect use eks requires aws-iam-authenticator.
  [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)



```
kconnect use eks [flags]
```

### Examples

```

  # Discover EKS clusters using SAML
  kconnect use eks --idp-protocol saml

  # Discover EKS clusters using SAML with a specific role
  kconnect use eks --idp-protocol saml --role-arn arn:aws:iam::000000000000:role/KubernetesAdmin

  # Discover an EKS cluster and add an alias to its connection history entry
  kconnect use eks --alias mycluster

  # Reconnect to a cluster by its connection history entry alias.
  kconnect to mycluster

  # Display the user's connection history as a table.
  kconnect ls

```

### Options

```
  -a, --alias string              Friendly name to give to give the connection
  -c, --cluster-id string         Id of the cluster to use.
  -h, --help                      help for eks
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --idp-protocol string       The idp protocol to use (e.g. saml)
  -k, --kubeconfig string         Location of the kubeconfig to use. (default "$HOME/.kube/config")
      --max-history int           Sets the maximum number of history items to keep (default 100)
  -n, --namespace string          Sets namespace for context in kubeconfig
      --no-history                If set to true then no history entry will be written
      --partition string          AWS partition to use (default "aws")
      --password string           The password to use for authentication
      --region string             AWS region to connect to
      --region-filter string      A filter to apply to the AWS regions list, e.g. 'us-' will only show US regions
      --role-arn string           ARN of the AWS role to be assumed
      --role-filter string        A filter to apply to the roles list, e.g. 'EKS' will only show roles that contain EKS in the name
      --set-current               Sets the current context in the kubeconfig to the selected cluster (default true)
      --username string           The username used for authentication
```

### Options inherited from parent commands

```
      --config string     Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --non-interactive   Run without interactive flag resolution
  -v, --verbosity int     Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### IDP Protocol Options

#### SAML Options

```
      --idp-endpoint string   identity provider endpoint provided by your IT team
      --idp-provider string   the name of the idp provider
      --partition string      AWS partition to use (default "aws")
      --region string         AWS region to connect to
```

### SEE ALSO

* [kconnect use](use.md)	 - Connect to a Kubernetes cluster provider and cluster.


> NOTE: this page is auto-generated from the cobra commands
