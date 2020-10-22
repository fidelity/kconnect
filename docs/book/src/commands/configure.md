## kconnect configure

Set and view your kconnect configuration.

### Synopsis

The `configure` command creates `kconnect` configuration files and displays previously-defined configurations
in a user-friendly display format.

If run with no flags, the command displays the configurations stored in the current user's 
`~/.kconnect/config.yaml` file.

The `configure` command can create a set of default configurations for a new system or a new user via the `-f` 
flag and a local filename or remote URL.  You would typically use this flag the first time you use `kconnect`.

```
kconnect configure [flags]
```

### Examples

```
  # Display the current configuration
  kconnect configure

  # Display the configuration as json
  kconnect configure --output json

  # Create a set of user configurations from a local file
  kconnect configure -f ./defaults.yaml

  # Create a set of user configurations from a remote location via HTTP
  kconnect configure -f https://mycompany.com/config.yaml

  # Create a set of user configirations from stdin
  cat ./config.yaml | kconnect configure -f -

```

### Options

```
  -f, --file string     File or remote location to use to set the default configuration
  -h, --help            help for configure
      --output string   Controls the output format for the result. (default "yaml")
```

### Options inherited from parent commands

```
      --config string     Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
      --non-interactive   Run without interactive flag resolution
  -v, --verbosity int     Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect](index.md) - The Kubernetes Connection Manager CLI

