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
