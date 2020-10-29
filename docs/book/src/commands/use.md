## kconnect use

Connect to a Kubernetes cluster provider and cluster.

### Synopsis

Connect to a managed Kubernetes cluster provider via the configured identity provider, prompting the user to enter or
choose connection settings appropriate to the provider and a target cluster once connected.

The kconnect tool generates a kubectl configuration context with a fresh access token to connect to the chosen cluster
and adds a connection history entry to store the chosen connection settings.  The user can then reconnect to the provider 
using the stored setting and refresh their access token by the connection history entry ID or alias.

The use command requires a target provider name as its first parameter.

* Note: kconnect requires [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) to use the AWS EKS provider.

```
kconnect use [flags]
```

### Options

```
  -h, --help   help for use
```

### Options inherited from parent commands

```
      --config string     Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --non-interactive   Run without interactive flag resolution
  -v, --verbosity int     Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md) - The Kubernetes Connection Manager CLI
* [kconnect use eks](use_eks.md) - Connect to eks and discover clusters for use

