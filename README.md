# Baskio - Build And Scan Kubernetes Images Openstack

A binary for building and scanning (with [Trivy](https://github.com/aquasecurity/trivy)) a Kubernetes image using the [eschercloud-image-builder](https://github.com/eschercloudai/image-builder) repo.
Once the image has been built, the CVE results will be pushed to GitHub Pages. Simply provide the required GitHub flags/env vars, and it will do the rest for you.


# Scope

⚠️Currently in ~~beta~~ alpha at the moment.

# Prerequisites
### GitHub Pages
You will need to set up your target repo for the GiHub Pages in advanced.
It only requires a `gh-pages` branch for this to work.
GitHub Pages should be configured to point to a `docs` directory as this is where the resulting static site will be placed.

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
  "build_name": "ubuntu-2204",
  "distro_name": "ubuntu",
  "guest_os_type": "ubuntu-64",
  "os_display_name": "Ubuntu 22.04",
  "shutdown_command": "shutdown -P now",
  "ssh_username": "ubuntu",
  "source_image": "SOURCE_IMAGE_ID",
  "networks": "NETWORK_ID",
  "flavor": "INSTANCE_FLAVOR",
  "attach_config_drive": "false",
  "use_floating_ip": "true",
  "floating_ip_network": "Internet",
  "crictl_version": "1.25.0",
  "kubernetes_semver": "v1.25.3",
  "kubernetes_rpm_version": "1.25.3-0",
  "kubernetes_series": "v1.25",
  "kubernetes_deb_version": "1.25.3-00"
}
```

# Usage

Simply run the binary with the following flags (minimum). See the example below.
This will also work with environment variables too - see the help for more info.
```shell
baski --build-os ubuntu-2204 \
--github-user GH_USER \
--github-project GH_PROJECT \
--github-token GH_TOKEN \
--github-pages-branch GH_PAGES_BRANCH \
--build-config openstack.json \
--network-id NETWORK_ID \
--server-flavor-id FLAVOR_ID \
--os-auth-url OS_AUTH_URL \
--os-project-name PROJECT_NAME \
--os-project-id PROJECT_ID \
--os-username OS_USERNAME \
--os-password OS_PASSWORD
```

### More info
For more flags, run `baskio --help`

# TODO
* Change the generated image and results name to be based on a date stamp and UUID for better image management.
* Have option to set the image as public in Openstack
* Provide option to choose name of gh-pages branch name
* Docker-ise this. (Maybe)


# License
The scripts and documentation in this project are released under the [Apache v2 License](LICENSE).