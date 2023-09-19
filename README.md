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

### Running locally
If you wish to run it locally then you can either build the binary and run it, or you can run it in docker by doing the following:
```shell
docker build -t baski:v0.0.0 -f docker/Dockerfile .

docker run --name baski -it --rm --env OS_CLOUD=some-cloud -v /path/to/openstack/clouds.yaml:/home/baski/.config/openstack/clouds.yaml -v /path/to/baski.yaml:/tmp/baski.yaml baski:v0.0.0

#Then from in here
baski build / scan / sign
```

# TODO
* Automatically clear up resources when ctrl-c is pressed.
* Make this work for more than just Openstack so that it's more useful to the community around the Kubernetes Image
  Builder?
* Add metrics/telemetry to the process.
* Create all option to allow whole process?

# License

The scripts and documentation in this project are released under the [Apache v2 License](LICENSE).
