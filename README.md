# Baskio - Build And Scan Kubernetes Images Openstack

A binary for building and scanning (with [Trivy](https://github.com/aquasecurity/trivy)) a Kubernetes image using
the [eschercloud-image-builder](https://github.com/eschercloudai/image-builder) repo.
Once the image has been built, the CVE results will be pushed to GitHub Pages. Simply provide the required GitHub
flags/env vars, and it will do the rest for you.

# Scope

⚠️Currently in beta at the moment.

# Prerequisites

### GitHub Pages

You will need to set up your target repo for the GiHub Pages in advanced.
It only requires a `gh-pages` branch for this to work.
GitHub Pages should be configured to point to a `docs` directory as this is where the resulting static site will be
placed.

### Openstack

It is expected that you have a network and sufficient security groups in place to run this.<br>
It will not create the network or security groups for you.

For example:

```
openstack network create image-builder
openstack subnet create image-builder --network image-builder --dhcp --dns-nameserver 1.1.1.1 --subnet-range 10.10.10.0/24 --allocation-pool start=10.10.10.10,end=10.10.10.200
openstack router create image-builder --external-gateway public1
openstack router add subnet image-builder image-builder

OS_SG=$(openstack security group list -c ID -c Name -f json | jq '.[]|select(.Name == "default") | .ID')
openstack security group rule create "${OS_SG}" --ingress --ethertype IPv4 --protocol TCP --dst-port 22 --remote-ip 0.0.0.0/0 --description "Allows SSH access"
openstack security group rule create "${OS_SG}" --egress --ethertype IPv4 --protocol TCP --dst-port -1 --remote-ip 0.0.0.0/0 --description "Allows TCP Egress"
openstack security group rule create "${OS_SG}" --egress --ethertype IPv4 --protocol UDP --dst-port -1 --remote-ip 0.0.0.0/0 --description "Allows UDP Egress"
```

### Openstack-build variables file

You will also require a source image to reference for the build to succeed.
When using, this you need to provide the following build config - changing any variables as required.

```
{
  "source_image": "SOURCE_IMAGE_ID",
  "networks": "NETWORK_ID",
  "flavor": "INSTANCE_FLAVOR",
  "floating_ip_network": "Internet",
}
```

# Usage
Simply run the binary with the following flags (minimum required). See the example below.
You must supply a clouds.yaml file for Openstack connectivity, 
alternatively the standard Openstack environment variables will work too. 

```shell
# Build an image
baskio build \
--build-os ubuntu-2204 \
--build-config openstack.json \
--network-id NETWORK_ID \
#--os-auth-url OS_AUTH_URL \
#--os-username OS_USERNAME \
#--os-password OS_PASSWORD \
#--os-project-id PROJECT_ID \
#--os-project-name PROJECT_NAME

# Scan an image
baskio scan \
--instance-flavor c.small \
--network-id NETWORK_ID \
--network-id IMAGE_ID \
#--os-auth-url OS_AUTH_URL \
#--os-username OS_USERNAME \
#--os-password OS_PASSWORD \
#--os-project-id PROJECT_ID \
#--os-project-name PROJECT_NAME

# Publish the CVE results of the scan to GitHub Pages
baskio publish \
--github-project GH_PROJECT \
--github-user GH_USER \
--github-token GH_TOKEN \
--github-user IMAGE_ID \
#--os-auth-url OS_AUTH_URL \
#--os-username OS_USERNAME \
#--os-password OS_PASSWORD \
#--os-project-id PROJECT_ID \
#--os-project-name PROJECT_NAME
```

### More info

For more flags and more info, run `baskio --help`

# TODO
* Have option to set the image as public in Openstack
* create all option to allow whole process 
* Make scanning a separate binary instead of packaging it in here - started process here by separating out building, scanning & publishing
* Make this work for more than just Openstack so that it's more useful to the community around the Kubernetes Image Builder?

# License

The scripts and documentation in this project are released under the [Apache v2 License](LICENSE).