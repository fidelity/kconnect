# Getting started

Once you have [installed](installation.md) kconnect you can see a list of the commands available by running:

```bash
kconnect help
```

<em>NOTE:</em> `kconnect` requires:
* [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) to authenticate to AWS EKS clusters.
* [kubelogin](https://github.com/Azure/kubelogin) to authenticate to Azure AKS clusters.

The general workflow for using kconnect is the following:

- `kconnect configure` - import configuration that contains defaults for your origanisation - **1 time**
- `kconnect use` - connect to a cluster for the first time - **only the first time**
- `kconnect to` - use to reconnect to a cluster that you have already connected to - **most used command day-to-day**

## Creating and importing configuration

Before using `kconnect` to connect to a Kubernetes cluster you may want to import an idetitiy provider configuration with your (or your organisations) defaults so that you don't have to supply all connection settings each time you connect to a new cluster.

You will need to create a configuration file (see example [here](https://github.com/fidelity/kconnect/blob/main/examples/config.yaml)). The configuration file can be imported from a local file or remote location via HTTP/HTTPS (and from stdin).

Each new user in your organization can then import the default configurations in this file using the `kconnect configure` command with the `-f` flag:

```bash
kconnect configure -f https://raw.githubusercontent.com/fidelity/kconnect/main/examples/config.yaml
```

Once the user has created their local configuration file, they should be able to display their configuration settings.

```bash
kconnect configure
```

## First time connection to a cluster

When discovering and connecting to a cluster for the first time you can do the following:

```bash
kconnect use eks --idp-protocol saml
```

This will guide you interactively setting the flags and selecting a cluster. It also gives you the option to set an easy-to-remember alias.

NOTE: only saml is supported at present for IdP.

## Reconnecting to a cluster

If you've previously connected to a cluster you can reconnect to it using the alias (if you set one):

```bash
kconnect to dev-bu-1
```

Or using the history id (which can be found by using `kconnect ls`):

```bash
kconnect to 01EM615GB2YX3C6WZ9MCWBDWBF
```

## Setting Flags

Flags can be replaced with environment variables by following the format `UPPERCASED_SNAKE_CASE` and appending to the `KCONNECT_` prefix.

For example`--username`can be set as`KCONNECT_USERNAME`; or `--idp-protocol` as`KCONNECT_IDP_PROTOCOL`.
