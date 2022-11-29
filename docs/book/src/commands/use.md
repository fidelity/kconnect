## kconnect use

Connect to a Kubernetes cluster provider and cluster.

### Synopsis


Connect to a managed Kubernetes cluster provider via the configured identity
provider, prompting the user to enter or choose connection settings appropriate
to the provider and a target cluster once connected.

The kconnect tool generates a kubectl configuration context with a fresh access
token to connect to the chosen cluster and adds a connection history entry to
store the chosen connection settings.  If given an alias name, kconnect will add
a user-friendly alias to the new connection history entry.

The user can then reconnect to the provider with the settings stored in the
connection history entry using the kconnect to command and the connection history
entry ID or alias.  When the user reconnects using a connection history entry,
kconnect regenerates the kubectl configuration context and refreshes their access
token.

The use command requires a target provider name as its first parameter. If no
value is supplied for --idp-protocol the first supported protocol for the
specified cluster provider.

* Note: interactive mode is not supported in windows git-bash application currently.

* Note: kconnect use eks requires aws-iam-authenticator.
  [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)

* Note: kconnect use aks requires kubelogin and azure cli.
  [kubelogin](https://github.com/Azure/kubelogin)
  [azure-cli](https://github.com/Azure/azure-cli)


```bash
kconnect use [flags]
```

### Examples

```bash

  # Connect to EKS and choose an available EKS cluster.
  kconnect use eks

  # Connect to an EKS cluster and create an alias for its connection history entry.
  kconnect use eks --alias mycluster

  # Reconnect to a cluster by its connection history entry alias.
  kconnect to mycluster

  # Display the user's connection history as a table.
  kconnect ls

```

### Options

```bash
  -h, --help   help for use
```

### Options inherited from parent commands

```bash
      --config string      Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --no-input           Explicitly disable interactivity when running in a terminal
      --no-version-check   If set to true kconnect will not check for a newer version
  -v, --verbosity int      Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md)	 - The Kubernetes Connection Manager CLI
* [kconnect use aks](use_aks.md)	 - Connect to the aks cluster provider and choose a cluster.
* [kconnect use eks](use_eks.md)	 - Connect to the eks cluster provider and choose a cluster.
* [kconnect use rancher](use_rancher.md)	 - Connect to the rancher cluster provider and choose a cluster.


> NOTE: this page is auto-generated from the cobra commands
