## kconnect alias add

Add an alias to a connection history entry

### Synopsis


Adds a user-friendly alias to a connection history entry.

The user can then reconnect and refresh the access token for that cluster using 
the alias instead of the connection history entry's unique ID.


```bash
kconnect alias add [flags]
```

### Examples

```bash

  # Add an alias to a connection history entry
  kconnect alias add --id 01EMEM5DB60TMX7D8SS2JCX3MT --alias dev-bu-1

  # Connect to a cluster using the alias
  kconnect to dev-bu-1

  # List available aliases
  kconnect alias ls

  # List available history entries - includes aliases
  kconnect ls

```

### Options

```bash
      --alias string   Alias name for a history entry
  -h, --help           help for add
      --id string      Id for a history entry
```

### Options inherited from parent commands

```bash
      --config string             Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --non-interactive           Run without interactive flag resolution
  -v, --verbosity int             Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect alias](alias.md)	 - Query and manipulate connection history entry aliases.


> NOTE: this page is auto-generated from the cobra commands
