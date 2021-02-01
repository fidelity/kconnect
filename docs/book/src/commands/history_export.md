## kconnect history export

Export history to an external file

### Synopsis


Allows users to export history to an external file. This file can then be 
imported by another user using the import command"


```bash
kconnect history export [flags]
```

### Examples

```bash

  # Export your history into a file
  kconnect history export -f exportfile.yaml
  
  # Set username and namespace for exported entries
  kconnect history export -f exportfile.yaml --set username=MYUSER,namespace=kube-system
  
  # Only export entries that match filter
  kconnect history export -f exportfile.yaml --filter region=us-east-1,alias=*dev*

```

### Options

```bash
  -f, --file string     file to import
      --filter string   filter to apply to import. Can specify multiple filters by using commas, and supports wilcards (*)
  -h, --help            help for export
      --set string      fields to set
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
