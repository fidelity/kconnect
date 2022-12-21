# Contributing to kconnect

Thank you for considering to contribute to kconnect 🎉👍

This document will provide help on how to go about this. You can contribute in the following ways:

* Reporting bugs
* Disclosing security concerns responsibly
* Suggesting features
* Contributing code

All contributions and project spaces are subject to our [Code of Conduct](https://github.com/fidelity/.github/blob/main/CODE_OF_CONDUCT.md).

## Reporting Bugs

Reporting bugs is an essential part in making kconnect better for its end users.

Please open an issue **unless** you are making a significant security disclosure. A new **Bug Report** can be raised [here](https://github.com/fidelity/kconnect/issues/new?assignees=&labels=kind%2Fbug&template=bug_report.md&title=).

When raising bugs please include as much information as possible including steps about how to reproduce the problem and what you expect the behavior to be.

## How to disclose security concerns responsibly

Please follow the instructions in our [security policy](https://github.com/fidelity/.github/blob/main/SECURITY.md) (also visible in the Security tab on the project's repo).

## Suggesting features

If there is a feature that you would like in kconnect then please let us know about it.

Features are also suggested using GitHub Issues. A new **Feature enhancement request** can be raised [here](https://github.com/fidelity/kconnect/issues/new?labels=kind%2Ffeature&template=feature_request.md&title=).

Include as much information as possible, understanding the problem that the feature is trying to solve will really help us in understanding the benefit.

## Contributing Code

Code contributions to kconnect are very welcome as long as you follow a few rules:

* Your contribution must be received under the project's open source license.
* You must have permission to make the contribution. We strongly recommend including a Signed-off-by line to indicate your adherence to the [Developer Certificate of Origin](https://developercertificate.org/).
* All code contributions must be made via PR, and all checks must pass before merging.

If you need a pointer on where to start you can look at the **good first issue** and **help wanted** issues:

* [good first issue](https://github.com/fidelity/kconnect/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) - small changes that are suitable for a beginner
* [help wanted](https://github.com/fidelity/kconnect/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) - more involved changes

You can also choose your own issue to work on from this list of available issues.

When choosing an issue to work on its preferable that you choose a issue that is planned for the next milestone and that has a higher priority....but this is just a nice to have and any contribution would be considered and welcomed.

### Getting started

1. Install Go >= 1.17
2. Fork the kconnect repo
3. Create a branch for your change:

```bash
git checkout -b <feature-name>
```

4. Use `git submodule` for saml2aws third party dependency

```bash
git submodule update --init --recursive
```

5. Make the change, including any additional tests
6. Run the tests:

```bash
make test
```

7. Check for linting errors:

```bash
make lint
```

8. Build and manually test kconnect locally:

```bash
make build
```

9. Commit and push your branch. We follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for the commits and PRs. .
10. Create a pull request. If the PR is a work in progress ensure that that **PR is created as a draft**.
11. Check that the PR checks pass

Once a PR has been created it will be reviewed by one of the maintainers of kconnect.

## Getting help

If you have other questions about this project, please open an issue. To reach the Fidelity OSPO directly, please email [opensource@fmr.com](mailto:opensource@fmr.com).
