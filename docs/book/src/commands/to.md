## kconnect to

Reconnect to a connection history entry.

### Synopsis


Reconnect to a cluster in the connection history by its entry ID or alias.

The kconnect tool creates an entry in the user's connection history with all the
connection settings each time it generates a new kubectl configuration context
for a Kubernetes cluster.  The user can then reconnect to the same cluster and
refresh their access token or regenerate the kubectl configuration context using
the connection history entry's ID or alias.

The to command also accepts - or LAST as proxy references to the most recent
connection history entry, or LAST~N for the Nth previous entry.

Although kconnect does not save the user's password in the connection history,
the user can avoid having to enter their password interactively by setting the
KCONNECT_PASSWORD environment variable or the --password command-line flag.
Otherwise kconnect will promot the user to enter their password.


```bash
kconnect to [historyid/alias/-/LAST/LAST~N] [flags]
```

### Examples

```bash

  # Reconnect based on an alias - aliases can be found using kconnect ls
  kconnect to uat-bu1

  # Reconnect based on an history id - history id can be found using kconnect ls
  kconnect to 01EM615GB2YX3C6WZ9MCWBDWBF

  # Reconnect interactively from history list
  kconnect to

  # Reconnect to current cluster (this is useful for renewing credentials)
  kconnect to -
  OR
  kconnect to LAST

  # Reconnect to cluster used before current one
  kconnect to LAST~1

  # Reconnect based on an alias supplying a password
  kconnect to uat-bu1 --password supersecret

  # Reconnect based on an alias supplying a password via env var
  KCONNECT_PASSWORD=supersecret kconnect to uat-bu2
 
```

### Options

```bash
  -h, --help                      help for to
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
  -k, --kubeconfig string         Location of the kubeconfig to use. (default "$HOME/.kube/config")
      --password string           Password to use
      --set-current               Sets the current context in the kubeconfig to the selected cluster (default true)
```

### Options inherited from parent commands

```bash
      --config string     Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --non-interactive   Run without interactive flag resolution
  -v, --verbosity int     Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md)	 - The Kubernetes Connection Manager CLI


> NOTE: this page is auto-generated from the cobra commands
