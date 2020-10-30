## kconnect alias ls

List all the aliases currently defined

### Synopsis

List all the aliases currently defined for connection history entries in the
user's connection history.

An alias is a user-friendly name for a connection history entry.


```
kconnect alias ls [flags]
```

### Examples

```
  # Display all the aliases as a table
  kconnect alias ls

  # Display all connection history entry aliases as a table
  kconnect alias ls

  # Display all connection history entry aliases as json
  kconnect alias ls --output json

  # Connect to a cluster using a connection history entry alias
  kconnect to ${alias}

  # List all connection history entries as a table - includes aliases
  kconnect ls

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

* [kconnect alias](alias.md)	 - Query and manipulate connection history entry aliases.

