## kconnect use eks

Connect to eks and discover clusters for use

### Synopsis

Connect to eks and discover clusters for use

```
kconnect use eks [flags]
```

### Examples

```
  # Discover EKS clusters using SAML
  kconnect use eks --idp-protocol saml

  # Discover EKS clusters using SAML with a specific role
  kconnect use eks --idp-protocol saml --role-arn arn:aws:iam::000000000000:role/KubernetesAdmin

```

### Options

```
  -a, --alias string              Friendly name to give to give the connection
  -c, --cluster-id string         Id of the cluster to use.
  -h, --help                      help for eks
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --idp-endpoint string       identity provider endpoint provided by your IT team
      --idp-protocol string       The idp protocol to use (e.g. saml)
      --idp-provider string       the name of the idp provider
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

* [kconnect use](use.md)	 - Connect to a target environment and discover clusters for use

