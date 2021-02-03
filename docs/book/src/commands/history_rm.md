## kconnect history rm

Remove history entries

### Synopsis


Allows users to delete history entries from their history


```bash
kconnect history rm [flags]
```

### Examples

```bash

  # Display history entries
  kconnect ls
  
  # Delete a history entry with specific ID
  kconnect history rm 01exm3ty400w9sr28jawc8fkae
  
  # Delete multiple history entries with specific IDs
  kconnect history rm 01exm3ty400w9sr28jawc8fkae 01exm3tvw2f5snkj18rk1ngmyb

  # Delete all history entries
  kconnect history rm --all
  
  # Delete all history entries that match the filter (e.g. alias that have "prod" in them)
  kconnect history rm --filter alias=*prod*

```

### Options

```bash
      --all             remove all entries
      --filter string   filter to apply to import. Can specify multiple filters by using commas, and supports wilcards (*)
  -h, --help            help for rm
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

* [kconnect history](history.md)	 - Import and export history


> NOTE: this page is auto-generated from the cobra commands
