## kconnect history import

Import history from an external file

### Synopsis


Allows users to import history from an external file. This can then be viewed 
using the ls command, or connected to using the to command.

These imported entries will be merged with existing ones. If there is a 
conflict, then the conflicting entry will not be imported (unless 
--overwrite flag is supplied).

Users can optionally set any fields in the imported entry


```bash
kconnect history import [flags]
```

### Examples

```bash

  # Imports the file into your history 
  kconnect history import -f importfile.yaml
  
  # Overwrite conflicting entries
  kconnect history import -f importfile.yaml --overwrite

  # Wipe existing history
  kconnect history import -f importfile.yaml --clean

  # Set username and namespace for imported entries
  kconnect history import -f importfile.yaml --set username=MYUSER,namespace=kube-system

  # Only import entries that match filter
  kconnect history import -f importfile.yaml --filter region=us-east-1,alias=*dev*

```

### Options

```bash
      --clean           delete all existing history
  -f, --file string     File to import
      --filter string   filter to apply to import
  -h, --help            help for import
      --overwrite       overwrite conflicting entries
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
