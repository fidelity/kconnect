## kconnect logout

Logs out of a cluster

### Synopsis


Logs out of a cluster. Can logout of specific cluster by their alias or entry ID. 
Log out of all clusters by using the --all flag
If neither above options are selected, will log out of current cluster


```bash
kconnect logout [flags]
```

### Options

```bash
      --alias string              comma delimited list of aliass
  -a, --all                       Logs out of all clusters
  -h, --help                      help for logout
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --ids string                comma delimited list of ids
  -k, --kubeconfig string         Location of the kubeconfig to use. (default "$HOME/.kube/config")
```

### Options inherited from parent commands

```bash
      --config string      Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --no-version-check   If set to true kconnect will not check for a newer version
      --non-interactive    Run without interactive flag resolution
  -v, --verbosity int      Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md)	 - The Kubernetes Connection Manager CLI


> NOTE: this page is auto-generated from the cobra commands
