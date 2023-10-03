# Baski - Build And Scan Kubernetes Images

[//]: # ([![Known Vulnerabilities]&#40;https://snyk.io/test/github/eschercloudai/baski/badge.svg&#41;]&#40;https://snyk.io/test/github/eschercloudai/baski&#41;)
[![Build on Tag](https://github.com/eschercloudai/baski/actions/workflows/tag.yml/badge.svg?branch=main&event=release)](https://github.com/eschercloudai/baski/actions/workflows/tag.yml)

A binary for building and scanning (with [Trivy](https://github.com/aquasecurity/trivy)) a Kubernetes image using
the [kubernetes-sigs/image-builder](https://github.com/kubernetes-sigs/image-builder) repo or
the [eschercloud-image-builder](https://github.com/eschercloudai/image-builder) repo where new functionality is required
but not yet merged upstream.

Baski also supports signing images and will tag the image with a digest so that a verification can be done against
images.

The scanning and signing functionality are separate from the build meaning these can be used on none Kubernetes images.

# Scope

⚠️Currently in beta at the moment.

# Supported clouds

| Cloud Provider                 |
|--------------------------------|
| [Openstack](docs/openstack.md) |

*More clouds could be supported but may not be maintained by EscherCloudAI.*

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

If you wish to run it locally then you can either build the binary and run it, or you can run it in docker by doing the
following:

```shell
docker build -t baski:v0.0.0 -f docker/baski/Dockerfile .

docker run --name baski -it --rm --env OS_CLOUD=some-cloud -v /path/to/openstack/clouds.yaml:/home/baski/.config/openstack/clouds.yaml -v /path/to/baski.yaml:/tmp/baski.yaml baski:v0.0.0

#Then from in here
baski build / scan / sign
```

### Baski Server

Baski server has been built so that all scan results are easily obtainable from an S3 endpoint. Run the server as
follows, and then you can query the server using the API.
The CVE results are searched for based on the locations generated during the `sign single` and `sign multiple`

```shell
docker build -t baski-server:v0.0.0 -f docker/server/Dockerfile .

docker run --name baski-server -p 8080 -it --rm baski-server:v0.0.0

baski-server run -a 0.0.0.0 -p 8080 --access-key SOME-ACCESS-KEY --secret-key SOME-SECRET-KEY --endpoint https://SOME-ENDPOINT --bucket baski

curl http://127.0.0.1:DOCKER-PORT/api/v1/scan/SOME-IMAGE-ID
```

# TODO

* Automatically clear up resources when ctrl-c is pressed.
* Make this work for more than just Openstack so that it's more useful to the community around the Kubernetes Image
  Builder?
* Add metrics/telemetry to the process.
* Create all option to allow whole process?

# License

The scripts and documentation in this project are released under the [Apache v2 License](LICENSE).
