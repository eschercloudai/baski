# Baski - Build And Scan Kubernetes Images

[//]: # ([![Known Vulnerabilities]&#40;https://snyk.io/test/github/eschercloudai/baski/badge.svg&#41;]&#40;https://snyk.io/test/github/eschercloudai/baski&#41;)
[![Build on Tag](https://github.com/eschercloudai/baski/actions/workflows/tag.yml/badge.svg?branch=main&event=release)](https://github.com/eschercloudai/baski/actions/workflows/tag.yml)

A binary for building and scanning (with [Trivy](https://github.com/aquasecurity/trivy)) a Kubernetes image using
the [eschercloud-image-builder](https://github.com/eschercloudai/image-builder) repo.
Once the image has been built, the CVE results will be pushed to GitHub Pages. Simply provide the required GitHub
flags/config file, and it will do the rest for you.

# Scope

⚠️Currently in beta at the moment.

# Supported clouds

| Cloud Provider                 |
|--------------------------------|
| [Openstack](docs/openstack.md) |

*More clouds could be supported but may not be maintained by EscherCloudAI.*

*EscherCloudAI will put the framework in place to the best of their availability/ability to facilitate more clouds being added.*

# Usage

Run the binary with a config file or see the help for a list of flags.
In the [example config](baski-example.yaml), not all fields are required and any fields that are not required are left
blank - unless the fields are enabled by a bool, for example in the Nvidia options where none are required
if `enable-nvidia-support` is set to false,

The following are valid locations for the `baski.yaml` config file are:
```shell
/tmp/
/etc/baski/
$HOME/.baski/
```

### More info

For more flags and more info, run `baski --help`

### GitHub Pages - Deprecated

You will need to set up your target repo for the GitHub Pages in advanced.
It only requires a `gh-pages` branch for this to work.
GitHub Pages should be configured to point to a `docs` directory as this is where the resulting static site will be
placed.

# TODO

* Make this work for more than just Openstack so that it's more useful to the community around the Kubernetes Image
  Builder?
* Remove dependency on GitHub Pages in the publish section - have this generate an artifact instead
* Add metrics/telemetry to the process
* Create all option to allow whole process?

# License

The scripts and documentation in this project are released under the [Apache v2 License](LICENSE).
