/*
Copyright 2023 EscherCloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ostack

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/eschercloudai/baski/pkg/constants"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

// Client contains the Env vars of the program as well as the Provider and any EndpointOptions.
// This is used in gophercloud connections.
type Client struct {
	Cloud           OpenstackCloud
	Provider        *gophercloud.ProviderClient
	EndpointOptions *gophercloud.EndpointOpts
}

// NewOpenstackClient creates the initial client for connecting to Openstack.
func NewOpenstackClient(cloud OpenstackCloud) *Client {
	client := &Client{
		Cloud: cloud,
	}
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: client.Cloud.Auth.AuthURL + "/" + strings.Join([]string{"v", strconv.Itoa(client.Cloud.IdentityApiVersion)}, ""),
		Username:         client.Cloud.Auth.Username,
		Password:         client.Cloud.Auth.Password,
		DomainName:       client.Cloud.Auth.UserDomainName,
		TenantName:       client.Cloud.Auth.ProjectName,
	}
	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		panic(err)
	}
	epOpts := &gophercloud.EndpointOpts{
		Region: client.Cloud.RegionName,
	}
	client.Provider = provider
	client.EndpointOptions = epOpts

	return client
}

// createImageClient will generate the image client required for updating image metadata.
func createImageClient(client *Client) *gophercloud.ServiceClient {
	c, err := openstack.NewImageServiceV2(client.Provider, *client.EndpointOptions)
	if err != nil {
		panic(err)
	}

	return c
}

// UpdateImageMetadata updates an images metadata.
func (c *Client) UpdateImageMetadata(imgID string, digest string) *images.Image {
	client := createImageClient(c)
	client.Microversion = "2.2"

	updateOpts := images.UpdateOpts{
		images.UpdateImageProperty{
			Op:    images.AddOp,
			Name:  "/digest",
			Value: digest,
		},
	}

	img, err := images.Update(client, imgID, updateOpts).Extract()

	if err != nil {
		log.Fatalln(err)
	}

	return img
}

// RemoveImage will delete an image from Openstack.
func (c *Client) RemoveImage(imgID string) {
	log.Println("removing image")
	client := createImageClient(c)
	res := images.Delete(client, imgID)
	if res.Err != nil {
		log.Println(res.Err)
	}
}

// createComputeClient will generate the compute client required for creating Servers and KeyPairs in Openstack.
func createComputeClient(client *Client) *gophercloud.ServiceClient {
	c, err := openstack.NewComputeV2(client.Provider, *client.EndpointOptions)
	if err != nil {
		panic(err)
	}

	return c
}

// CreateKeypair creates a new KeyPair in Openstack.
func (c *Client) CreateKeypair(keyNamePrefix string) *keypairs.KeyPair {
	client := createComputeClient(c)
	client.Microversion = "2.2"

	kp, err := keypairs.Create(client, keypairs.CreateOpts{
		Name: keyNamePrefix + "-baski-key",
		Type: "ssh",
	}).Extract()
	if err != nil {
		log.Fatalln(err)
	}

	return kp
}

// CreateServer creates a compute instance in Openstack.
func (c *Client) CreateServer(keypair *keypairs.KeyPair, imageID, flavorName, networkID string, enableConfigDrive bool) (*servers.Server, string) {
	trivyVersion := constants.TrivyVersion
	client := createComputeClient(c)

	serverFlavorID := c.GetFlavorIDByName(flavorName)

	serverOpts := servers.CreateOpts{
		Name:           imageID + "-scanner",
		FlavorRef:      serverFlavorID,
		ImageRef:       imageID,
		SecurityGroups: []string{"default"},
		UserData: []byte(fmt.Sprintf(`#!/bin/bash
wget -q -O- "https://github.com/aquasecurity/trivy/releases/download/v%s/trivy_%s_Linux-64bit.tar.gz" | tar xzf -
chmod u+x trivy
sudo ./trivy rootfs --scanners vuln -f json -o /tmp/results.json /
`, trivyVersion, trivyVersion)),
		AvailabilityZone: "",
		Networks: []servers.Network{
			{
				UUID: networkID,
			},
		},
		ConfigDrive: &enableConfigDrive,
		Min:         1,
		Max:         1,
	}

	createOpts := keypairs.CreateOptsExt{
		CreateOptsBuilder: serverOpts,
		KeyName:           keypair.Name,
	}

	server, err := servers.Create(client, createOpts).Extract()
	if err != nil {
		c.RemoveKeypair(keypair.Name)
		panic(err)
	}

	//TODO: If no IP is available, allocate one and attach. If none available to allocate, fail.
	freeIP := attachFloatingIP(client, server.ID)

	return server, freeIP
}

// RemoveServer will delete a Server from Openstack.
func (c *Client) RemoveServer(serverID string) {
	log.Println("removing scanning server")
	client := createComputeClient(c)
	res := servers.Delete(client, serverID)

	if res.Err != nil {
		log.Println(res.Err)
	}
}

// RemoveKeypair will delete a Keypair from Openstack.
func (c *Client) RemoveKeypair(keyName string) {
	log.Println("removing keypair.")
	client := createComputeClient(c)
	res := keypairs.Delete(client, keyName, keypairs.DeleteOpts{})
	if res.Err != nil {
		log.Println(res.Err)
	}
}

// GetFlavorIDByName will take a name of a flavor and attempt to find the ID from Openstack.
func (c *Client) GetFlavorIDByName(name string) string {
	client := createComputeClient(c)

	listOpts := flavors.ListOpts{
		AccessType: flavors.PublicAccess,
	}

	allPages, err := flavors.ListDetail(client, listOpts).AllPages()
	if err != nil {
		panic(err)
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		panic(err)
	}

	for _, flavor := range allFlavors {
		if flavor.Name == name {
			return flavor.ID
		}
	}
	return ""
}

// attachFloatingIP will find the first free IP available and attach it to the instance.
// If it cannot find one, it will error.
func attachFloatingIP(client *gophercloud.ServiceClient, serverID string) string {
	// Floating IP assignment
	allIPsPages, err := floatingips.List(client).AllPages()
	if err != nil {
		panic(err)
	}

	allFloatingIPs, err := floatingips.ExtractFloatingIPs(allIPsPages)
	if err != nil {
		panic(err)
	}

	var freeIP string

	for _, fip := range allFloatingIPs {
		if fip.InstanceID == "" {
			freeIP = fip.IP
			break
		}
	}

	if freeIP == "" {
		panic("couldn't find a free IP")
	}

	log.Println("waiting for the instance to come up before attaching an IP")
	time.Sleep(15 * time.Second)
	log.Printf("attaching IP %s to the instance %s", freeIP, serverID)

	associateOpts := floatingips.AssociateOpts{
		FloatingIP: freeIP,
	}

	err = floatingips.AssociateInstance(client, serverID, associateOpts).ExtractErr()
	if err != nil {
		panic(err)
	}

	return freeIP
}
