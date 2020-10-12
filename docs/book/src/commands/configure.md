## kconnect configure

Set and view your default kconnect configuration. If no flags are supplied your config is displayed.

### Synopsis

Set and view your default kconnect configuration. If no flags are supplied your config is displayed.

```
kconnect configure [flags]
```

### Examples

```
  # Display the current configuration
  kconnect configure

  # Display the configuration as json
  kconnect configure --output json

  # Set the configuration from a local file
  kconnect configure -f ./defaults.yaml

  # Set the configuration from a remote location via HTTP
  kconnect configure -f https://mycompany.com/config.yaml

  # Set the congigiration from stdin
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

* [kconnect](index.md)	 - The Kubernetes Connection Manager CLI

