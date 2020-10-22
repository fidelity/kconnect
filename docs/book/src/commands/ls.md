## kconnect ls

Query your connection history

### Synopsis

Displays your connection history entries along with their entry IDs and aliases.

Use the `kconnect to` command and an alias or connection history entry ID to reconnect to one of the listed clusters.

```
kconnect ls [flags]
```

### Examples

```
  # Display all the history as a table
  kconnect ls

  # Display the history as yaml
  kconnect ls --output yaml

  # Get the history for a specific entry id
  kconnect ls --id 01EM615GB2YX3C6WZ9MCWBDWBF

  # Get the history entries for eks
  kconnect ls --cluster-provider eks

  # Connect to a cluster using a history entry
  kconnect to ${entryId}

  # Connect to a cluster using an alias
  kconnect to ${alias}

```

### Options

```
      --alias string               Alias name for a history entry
      --cluster-provider string    Name of a cluster provider (i.e. eks)
  -h, --help                       help for ls
      --history-location string    Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --id string                  Id for a history entry
      --identity-provider string   Name of a identity provider (i.e. saml)
  -k, --kubeconfig string          Location of the kubeconfig to use. (default "$HOME/.kube/config")
      --max-history int            Sets the maximum number of history items to keep (default 100)
      --no-history                 If set to true then no history entry will be written
  -o, --output string              Output format for the results (default "table")
      --provider-id string         Provider specific for a cluster
```

### Options inherited from parent commands

```
      --config string     Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --non-interactive   Run without interactive flag resolution
  -v, --verbosity int     Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md) - The Kubernetes Connection Manager CLI
* [kconnect to](to.md) - Connect to a cluster using an alias or history entry
