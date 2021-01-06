## kconnect

The Kubernetes Connection Manager CLI

### Synopsis


The kconnect tool uses a pre-configured Identity Provider to log in to one or
more managed Kubernetes cluster providers, discovers the list of clusters
visible to your authenticated user and options, and generates a kubectl
configutation context for the selected cluster.

Most kubectl contexts include an authentication token which kubectl sends to
Kubernetes with each request rather than a username and password to establish
your identity.  Authentication tokens typically expire after some time.  The
user must then to log in again to the managed Kubernetes service provider and
regenerate the kubectl context for that cluster connection in order to refresh
the access token.

The kconnect tool makes this much easier by automating the login and kubectl
context regeneration process, and by allowing the user to repeat previously
successful connections.

Each time kconnect creates a new connection context, the kconnect tool saves the
information for that connection in the user's connection history list.  The user
can then display their connection history entries and reconnect to any entry by
its unique ID (or by a user-friendly alias) to refresh an expired access token
for that cluster.


```bash
kconnect [flags]
```

### Examples

```bash

  # Display a help screen with kconnect commands.
  kconnect help

  # Configure the default identity provider and connection profile for your user.
  #
  # Use this command to set up kconnect the first time you use it on a new system.
  #
  kconnect configure -f FILE_OR_URL

  # Create a kubectl confirguration context for an AWS EKS cluster.
  #
  # Use this command the first time you connect to a new cluster or a new context.
  #
  kconnect use eks

  # Display connection history entries.
  #
  kconnect ls

  # Add an alias to a connection history entry.
  #
  kconnect alias add --id 012EX456834AFXR0F2NZT68RPKD --alias MYALIAS

  # Reconnect and refresh the token for an aliased connection history entry.
  #
  # Use this to reconnect to a provider and refresh an expired access token.
  #
  kconnect to MYALIAS

  # Display connection history entry aliases.
  #
  kconnect alias ls

```

### Options

```bash
      --config string      Configuration file for application wide defaults. (default "$HOME/.kconnect/config.yaml")
  -h, --help               help for kconnect
      --no-version-check   If set to true kconnect will not check for a newer version
      --non-interactive    Run without interactive flag resolution
  -v, --verbosity int      Sets the logging verbosity. Greater than 0 is debug and greater than 9 is trace.
```

### SEE ALSO

* [kconnect alias](alias.md)	 - Query and manipulate connection history entry aliases.
* [kconnect configure](configure.md)	 - Set and view your kconnect configuration.
* [kconnect logout](logout.md)	 - Logs out of a cluster
* [kconnect ls](ls.md)	 - Query the user's connection history
* [kconnect to](to.md)	 - Reconnect to a connection history entry.
* [kconnect use](use.md)	 - Connect to a Kubernetes cluster provider and cluster.
* [kconnect version](version.md)	 - Display version & build information


> NOTE: this page is auto-generated from the cobra commands
