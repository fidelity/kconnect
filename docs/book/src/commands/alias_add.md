## kconnect alias add

Add an alias to a history entry

### Synopsis

Adds a user-friendly alias to a connection history entry.

You can then connect again to the same cluster using the `kconnect to` command and the alias.

```
kconnect alias add [flags]
```

### Examples

```
  # Add an alias to a history entry
  kconnect alias add --id 01EMEM5DB60TMX7D8SS2JCX3MT --alias dev-bu-1

  # Connect to a cluster using the alias
  kconnect to dev-bu-1

  # List available aliases
  kconnect alias ls

  # List available history entries - includes aliases
  kconnect ls

```

### Options

```
      --alias string   Alias name for a history entry
  -h, --help           help for add
      --id string      Id for a history entry
```

### Options inherited from parent commands

```
      --config string             Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --non-interactive           Run without interactive flag resolution
  -v, --verbosity int             Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect alias](alias.md) - Query and manipulate aliases for connection history entries.
* [kconnect alias ls](alias_ls.md) - List available aliases.
* [kconnect to](to.md) - Connect to a cluster using an alias or history entry.
* [kconnect ls](ls.md) - Query your connection history.
