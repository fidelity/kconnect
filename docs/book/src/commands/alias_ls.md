## kconnect alias ls

List all the aliases currently defined

### Synopsis

List all the aliases currently defined

```
kconnect alias ls [flags]
```

### Examples

```
  # Display all the aliases as a table
  kconnect alias ls

  # Display all the aliases as json
  kconnect alias ls --output json

```

### Options

```
  -h, --help            help for ls
      --output string   Output format for the results (default "table")
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

