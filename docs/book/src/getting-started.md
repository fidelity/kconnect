# Getting started

Once you have installed kconnect you can see a list of the commands available by running:

```bash
kconnect
```

If you wanted to discover clusters in EKS and generate a kubeconfig for a selected cluster you can run the following command which will guide you through connecting:

```bash
kconnect use eks --idp-protocol saml
```

NOTE: only saml is supported at present for IdP.

## Setting Flags

Flags can be replaced with environment variables by following the format `UPPERCASED_SNAKE_CASE` and appending to the `KCONNECT_` prefix.

For example`--username`can be set as`KCONNECT_USERNAME`; or `--idp-protocol` as`KCONNECT_IDP_PROTOCOL`.
