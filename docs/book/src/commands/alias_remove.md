## kconnect alias remove

Remove connection history entry aliases.

### Synopsis

Remove an alias from a single connection history entry by the entry ID or the alias.

This command will remove all connection history aliases if passed the --all flag.

```
kconnect alias remove [flags]
```

### Examples

```
  # Remove an alias using the alias name
  kconnect alias remove --alias dev-bu-1

  # Remove an alias using a history entry id
  kconnect alias remove --id 01EMEM5DB60TMX7D8SS2JCX3MT

  # Remove all aliases
  kconnect alias remove --all

  # List available aliases
  kconnect alias ls

  # Query your connection history - includes aliases
  kconnect ls

```

### Options

```
      --alias string   Alias name for a history entry
      --all            Remove all aliases from the histiry entries
  -h, --help           help for remove
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
* [kconnect alias ls](alias_ls.ms) - List available aliases.
* [kconnect ls](ls.md) - Query your connection history.
