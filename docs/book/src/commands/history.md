## kconnect history

Import and export history

### Synopsis


Command to allow users to import or export history files.

A common use case would be for one member of a team to generate the history +
alias config for their teams cluster(s). They could then send this file out to
the rest of the team, who can then import it. On import, they can set their
username for all of the history entries.


```bash
kconnect history [flags]
```

### Examples

```bash

	# Export all history entries that have alias = *dev*
	kconnect history export -f history.yaml --filter alias=*dev*

	# Import history entries and set username
	kconnect history import -f history.yaml --set username=myuser

```

### Options

```bash
  -h, --help                      help for history
      --history-location string   Location of where the history is stored. (default "$HOME/.kconnect/history.yaml")
```

### Options inherited from parent commands

```bash
      --config string      Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --no-input           Explicitly disable interactivity when running in a terminal
      --no-version-check   If set to true kconnect will not check for a newer version
  -v, --verbosity int      Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md)	 - The Kubernetes Connection Manager CLI
* [kconnect history export](history_export.md)	 - Export history to an external file
* [kconnect history import](history_import.md)	 - Import history from an external file
* [kconnect history rm](history_rm.md)	 - Remove history entries


> NOTE: this page is auto-generated from the cobra commands
