# Contributing to kconnect

Thank you for considering to contribute to kconnect 🎉👍

This document will provide help on how to go about this. You can contribute in the following ways:

* Reporting bugs
* Suggesting features
* Contributing code

## Reporting Bugs

Reporting bugs is an essential part in making kconnect better for its end users.

Bugs are reported using GitHub issues. A new **Bug Report** can be raised [here](https://github.com/fidelity/kconnect/issues/new?assignees=&labels=kind%2Fbug&template=bug_report.md&title=).

When raising bugs please include as much information as possible including steps about how to reproduce the problem and what you expect the behavior to be.

## Suggesting features

If there is a feature that you would like in kconnect then please let us know about it.

Features are also suggested using GitHub Issues. A new **Feature enhancement request** can be raised [here](https://github.com/fidelity/kconnect/issues/new?labels=kind%2Ffeature&template=feature_request.md&title=).

Include as much information as possible, understanding the problem that the feature is trying to solve will really help us in understanding the benefit.

## Contributing Code

Code contributions to kconnect are very welcome.

If you need a pointer on where to start you can look at the **good first issue** and **help wanted** issues:

* [good first issue](https://github.com/fidelity/kconnect/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) - small changes that are suitable for a beginner
* [help wanted](https://github.com/fidelity/kconnect/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) - more involved changes

You can also choose your own issue to work on from this list of available issues.

When choosing an issue to work on its preferable that you choose a issue that is planned for the next milestone and that has a higher priority....but this is just a nice to have and any contribution would be considered and welcomed.

### Getting started

1. Install Go >= 1.13
2. Fork the kconnect repo
3. Create a branch for your change:

```bash
git checkout -b <feature-name>
```

4. Make the change, including any additional tests
5. Run the tests:

```bash
make test
```

6. Check for linting errors:

```bash
make lint
```

7. Build and manually test kconnect locally:

```bash
make build
```

8. Commit and push your branch. We follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for the commits and PRs. If the PR is a work in progress ensure that you prefix the title with WIP:.
9. Create a pull request
10. Check that the PR checks pass

Once a PR has been created it will be reviwed by one of the maintainers of kconnect.
