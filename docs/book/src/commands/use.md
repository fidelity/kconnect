## kconnect use

Connect to a target environment and discover clusters for use

### Synopsis

The `use` command connects to a Kubernetes cluster provider through the configured identity provider in order to discover available 
Kubernetes clusters and obtain or update `kubectl` configuration contexts with a fresh access token.

The `use` command requires a target provider name as its first parameter.

* [kconnect use eks](use_eks.md) connects to AWS Elastic Kubernetes Service
  * NOTE: requires [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)

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

