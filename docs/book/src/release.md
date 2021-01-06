# Releasing kconnect

The release of *kconnect* is done using [goreleaser](https://goreleaser.com/) which is orchestrated using Githib Actions.

## Process

The following steps are required to do a release:

1. Merge any PRs into the **main** branch that contain features, bug fixes and other changes that you want to include in the release.
2. We use [Semver 2.0](https://semver.org/) for the release numbering. You will need to decide what the next release number will be based on the changes and whether it will be a *"pre-release"*:
    * *Normal release*: it will follow the MAJOR.MINOR.PATCH format, for example 0.3.0
    * *Pre-release*: it will follow the MAJOR.MINOR.PATCH-rc.RCNUMBER format, for example 0.3.0-rc.1
3. Locally on your machine get the latest **main** branch code:
```bash
git checkout main
git pull
```
4. Tag the main branch with the release number previously determined:
```bash
git tag -a 0.3.0 -m "0.3.0"
```
5. Push the new tag to GitHub:
```bash
git push origin 0.3.0
```
6. Go to GitHub and check on the **goreleaser** [action](https://github.com/fidelity/kconnect/actions?query=workflow%3Agoreleaser). This action is what does the actual release.
7. Once the **goreleaser** action completes go to the [releases on GitHub](https://github.com/fidelity/kconnect/releases) and check the release is available.
8. Click **Edit** next to the release and tidy up the **Changelog** entries. If there are any breaking changes then a new markdown section should be added to the top that documents this.
9. Locally on your machine, run this command:
```bash
sudo docker run -v /path/to/repo/.krew.yaml:/tmp/template-file.yaml rajatjindal/krew-release-bot:v0.0.38 krew-release-bot template --tag 0.3.0 --template-file /tmp/template-file.yaml
```
10. Take the output of this command (excluding logging output). Update this file https://github.com/fidelity/krew-index/blob/main/plugins/connect.yaml with the output you just obtainted (note: these manual steps will be automated in future)

## Implementation

We use [goreleaser](https://goreleaser.com/) to do the majority of the build, packaging and release. The [.goreleaser.yml](https://github.com/fidelity/kconnect/blob/main/.goreleaser.yml) configuration file drives this.

The **goreleaser** GitHub Action that kicks off goreleaser on tagging the main branch is located [here](https://github.com/fidelity/kconnect/blob/main/.github/workflows/release.yml).

There is an additional GitHub workflow thats used to publish the docs to GitHub pages and that's located [here](https://github.com/fidelity/kconnect/blob/main/.github/workflows/release-docs.yml).

