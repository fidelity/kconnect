## kconnect use eks

Connect to eks and discover clusters for use

### Synopsis

The `use eks` command attempts to authenticate through the configured identity provider using the supplied credentials, connect to AWS EKS using the supplied 
connection settings, discover available EKS clusters, and obtain `kubectl` configurations for the chosen cluster.

If successful, the command creates a new connection history entry to store those credentials and AWS EKS connection settings.

The user can then use the `to` command to reconnect to AWS EKS using the values stored in the connection history entry and generate a new `kubectl` 
configuration context for the chosen cluster with a fresh Kubernetes access token when the stored access token expires.

If supplied with an alias name, the `use eks` command will define an alias for the new connection history entry.  When run in interactive mode, the command will 
prompt the user to create an alias.

The `ls` command lists previously successful connection history entries - including their aliases.

The `alias ls` command lists all available connection history entry aliases.

The `to` command accepts either a connection history entry ID, alias or reference when reconnecting to AWS EKS to request a fresh access token.

```
kconnect use eks [flags]
```

### Examples

```
  # Discover EKS clusters using SAML
  kconnect use eks --idp-protocol saml

  # Discover EKS clusters using SAML with a specific role
  kconnect use eks --idp-protocol saml --role-arn arn:aws:iam::000000000000:role/KubernetesAdmin

  # List available connection history entries
  kconnect ls

  # Reconnect to a cluster using a connection history entry ID
  kconnect to ${entryId}

  # Reconnect to a cluster using a connection history entry alias
  kconnect to ${alias}

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
      --namespace string          Sets namespace for context in kubeconfig
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

* [kconnect use](use.md) - Connect to a target environment and discover clusters for use
* [kconnect to](to.md) - Connect to a cluster using an alias or history entry
