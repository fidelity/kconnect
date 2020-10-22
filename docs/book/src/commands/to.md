## kconnect to

Reconnect to a previously connected cluster

### Synopsis

Reconnect to a previously connected cluster using an alias or connection history entry.

The `to` command accepts `-` or `LAST` as a proxy for the most recent connection history entry, or `LAST~N` for previous entries.

Use the `kconnect ls` command to query your connection history.

The `to` command will connect using the password passed in via the `--password` flag or stored in the `KCONNECT_PASSWORD` 
environment variable.  It will prompt the user for a password if they have not supplied one via the flag or environment variable.

```
kconnect to [historyid/alias/-/LAST/LAST~N] [flags]
```

### Examples

```
  # List connection history entries - including aliases
  kconnect ls

  # Reconnect to a cluster by its connection history entry ID
  kconnect to 01EM615GB2YX3C6WZ9MCWBDWBF

  # List available connection history entry aliases
  kconnect alias ls

  # Reconnect to a cluster by its connection history entry alias
  kconnect to uat-bu1

  # Reconnect to the most recent connection - useful for renewing credentials
  kconnect to -
  # or 
  kconnect to LAST

  # Reconnect to cluster one entry before the most recent connection (second in the history list)
  kconnect to LAST~1

  # Reconnect using the supplied password
  kconnect to uat-bu1 --password supersecret

  # Store a password in the KCONNECT_PASSWORD environment variable and reconnect using the stored password
  KCONNECT_PASSWORD=supersecret kconnect to uat-bu1

```

### Options

```
  -h, --help                      help for to
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
  -k, --kubeconfig string         Location of the kubeconfig to use. (default "$HOME/.kube/config")
      --password string           Reconnect using the supplied password string
      --set-current               Sets the current context in the kubeconfig to the selected cluster (default true)
```

### Options inherited from parent commands

```
      --config string     Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --non-interactive   Run without interactive flag resolution
  -v, --verbosity int     Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md) - The Kubernetes Connection Manager CLI
* [kconnect ls](ls.md) - Query your connection history
* [kconnect alias ls](alias_ls.md) - List available aliases
