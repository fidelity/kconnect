## kconnect alias remove

Remove an alias from a history entry. Or remove all aliases.

### Synopsis

Remove an alias from a history entry. Or remove all aliases.

```
kconnect alias remove [flags]
```

### Examples

```
  # Remove an alias using the alis name
  kconnect alias remove --alias dev-bu-1

  # Remove an alias using a histiry entry id
  kconnect alias remove --id 01EMEM5DB60TMX7D8SS2JCX3MT

  # Remove all aliases
  kconnect alias remove --all

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

* [kconnect alias](alias.md)	 - Query and manipulate aliases for your connection history.

