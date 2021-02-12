## kconnect config

Set and view your kconnect configuration.

### Synopsis


The configure command creates kconnect configuration files and displays
previously-defined configurations in a user-friendly display format.

If run with no flags, the command displays the configurations stored in the
current user's $HOME/.kconnect/config.yaml file.

The configure command can create a set of default configurations for a new
system or a new user via the -f flag and a local filename or remote URL.

The user typically only needs to use this command the first time they use
kconnect.


```bash
kconnect config [flags]
```

### Examples

```bash

  # Display user's current configurations
  kconnect config

  # Display the user's configurations as json
  kconnect config --output json

  # Set the user's configurations from a local file
  kconnect config -f ./defaults.yaml

  # Set the user's configurations from a remote location via HTTP
  kconnect config -f https://mycompany.com/config.yaml

  # Set the user's configurations from stdin
  cat ./config.yaml | kconnect config -f -

```

### Options

```bash
  -f, --file string       File or remote location to use to set the default configuration
  -h, --help              help for config
      --output string     Controls the output format for the result. (default "yaml")
      --password string   The password used for authentication
      --username string   The username used for authentication
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


> NOTE: this page is auto-generated from the cobra commands
