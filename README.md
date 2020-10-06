# `kconnect` - The Kubernetes Connection Manager CLI

![GitHub issues](https://img.shields.io/github/issues/fidelity/kconnect)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/fidelity/kconnect)](https://goreportcard.com/report/github.com/fidelity/kconnect)

## What is kconnect?

kconnect is a CLI utility that can be used to discover and securely access Kubernetes clusters across multiple operating environments.

Based on the authentication mechanism chosen the CLI will discover Kubernetes clusters you are allowed to access in a target hosting environment (i.e. EKS, AKS, Rancher) and generate a kubeconfig for a chosen cluster.

## Features

- Authenticate using SAML against a number of IdP
- Discover EKS clusters
- Generate a kubeconfig for a cluster
- Query history of connected servers
- Regenerate the kubeconfig from your history by using an id or an alias
- Import defaults values for your company

## Installation

[Releases](https://github.com/fidelity/kconnect/releases) are available for download for OSX, Linux and Windows.

To install on OSX you can use homebrew:

```bash
brew install fidelity/tap/kconnect
```

## Getting Started

Once you have installed kconnect you can see a list of the commands available by running:

```bash
kconnect
```

If you wanted to discover clusters in EKS and generate a kubeconfig for a selected cluster you can run the following command which will guide you through connecting:

```bash
kconnect use eks --idp-protocol saml
```

NOTE: only saml is supported at present for IdP.

Documentation is contained in the `/docs` directory. The [index is here](docs/README.md).

## Contributions

Contributions are very welcome. Please read the [contributing guide](CONTRIBUTING.md) or see the docs.
