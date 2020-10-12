## kconnect renew

Reconnect to last cluster

### Synopsis

Reconnect to last cluster

```
kconnect renew [flags]
```

### Options

```
  -h, --help                      help for renew
      --history-location string   Location of where the history is stored (default "/home/richard/.kconnect/history.yaml")
  -k, --kubeconfig string         Location of the kubeconfig to use (default "/home/richard/.kube/config")
      --password string           Password to use
      --set-current               Sets the current context in the kubeconfig to the selected cluster (default true)
```

### Options inherited from parent commands

```
      --config string     Configuration file for application defaults (default "/home/richard/.kconnect/config.yaml")
      --non-interactive   Run without interactive flag resolution
  -v, --verbosity int     Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md)	 - The Kubernetes Connection Manager CLI

