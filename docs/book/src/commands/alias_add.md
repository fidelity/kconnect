## kconnect alias add

Add an alias to a history entry

### Synopsis

Add an alias to a history entry

```
kconnect alias add [flags]
```

### Examples

```
  # Add an alias for a entry
  kconnect alias add --id 01EMEM5DB60TMX7D8SS2JCX3MT --alias dev-bu-1

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

* [kconnect alias](alias.md)	 - Query and manipulate aliases for your connection history.

