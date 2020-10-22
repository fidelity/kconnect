## kconnect alias

Query and manipulate connection history aliases.

### Synopsis

An alias is a user-friendly name for a connection history entry, otherwise referred to by its entry ID.

The alias command and sub-commands allow you to query and manipulate aliases for connection history entries.

```
kconnect alias [flags]
```

### Options

```
  -h, --help                      help for alias
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
```

### Options inherited from parent commands

```
      --config string     Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --non-interactive   Run without interactive flag resolution
  -v, --verbosity int     Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md) - The Kubernetes Connection Manager CLI
* [kconnect alias add](alias_add.md) - Add an alias to a history entry
* [kconnect alias ls](alias_ls.md) - List all the aliases currently defined
* [kconnect alias remove](alias_remove.md) - Remove an alias from a history entry. Or remove all aliases.

