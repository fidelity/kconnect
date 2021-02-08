## kconnect alias remove

Remove connection history entry aliases.

### Synopsis


Remove an alias from a single connection history entry by the entry ID or the
alias.

Set the --all flag on this command to remove all connection history aliases from
the user's connection history.


```bash
kconnect alias remove [flags]
```

### Examples

```bash

  # Remove an alias using the alias name
  kconnect alias remove --alias dev-bu-1

  # Remove an alias using a histiry entry id
  kconnect alias remove --id 01EMEM5DB60TMX7D8SS2JCX3MT

  # Remove all aliases
  kconnect alias remove --all

  # List available aliases
  kconnect alias ls

  # Query your connection history - includes aliases
  kconnect ls

```

### Options

```bash
      --alias string   Alias name for a history entry
      --all            Remove all aliases from the histiry entries
  -h, --help           help for remove
      --id string      Id for a history entry
```

### Options inherited from parent commands

```bash
      --config string             Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
      --no-input                  Explicitly disable interactivity when running in a terminal
      --no-version-check          If set to true kconnect will not check for a newer version
  -v, --verbosity int             Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect alias](alias.md)	 - Query and manipulate connection history entry aliases.


> NOTE: this page is auto-generated from the cobra commands
