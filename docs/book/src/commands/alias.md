## kconnect alias

Query and manipulate connection history entry aliases.

### Synopsis


An alias is a user-friendly name for a connection history entry, otherwise
referred to by its entry ID.

The alias command and sub-commands allow you to query and manipulate aliases for
connection history entries.


```bash
kconnect alias [flags]
```

### Examples

```bash

  # Add an alias to an existing connection history entry
  kconnect alias add --id 123456 --alias appdev

  # List available connection history entry aliases
  kconnect alias ls

  # Remove an alias from a connection history entry
  kconnect alias remove --alias appdev

```

### Options

```bash
  -h, --help                      help for alias
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
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
* [kconnect alias add](alias_add.md)	 - Add an alias to a connection history entry
* [kconnect alias ls](alias_ls.md)	 - List all the aliases currently defined
* [kconnect alias remove](alias_remove.md)	 - Remove connection history entry aliases.


> NOTE: this page is auto-generated from the cobra commands
