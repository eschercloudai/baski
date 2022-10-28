/*
Copyright 2022 EscherCloud.
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
	"github.com/drew-viles/baskio/pkg/constants"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"log"
	"strings"
	"time"
)

// Client contains the Env vars of the program as well as the Provider and any EndpointOptions.
// This is used in gophercloud connections.
type Client struct {
	Env             constants.Env
	Provider        *gophercloud.ProviderClient
	EndpointOptions *gophercloud.EndpointOpts
}

// OpenstackInit creates the initial client for connecting to Openstack.
func (c *Client) OpenstackInit() {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: c.Env.AuthURL + "/" + strings.Join([]string{"v", c.Env.IdentityAPIVersion}, ""),
		Username:         c.Env.Username,
		Password:         c.Env.Password,
		DomainName:       c.Env.UserDomainName,
		TenantName:       c.Env.ProjectName,
	}
	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		panic(err)
	}
	epOpts := &gophercloud.EndpointOpts{
		Region: c.Env.Region,
	}
	c.Provider = provider
	c.EndpointOptions = epOpts
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
func (c *Client) CreateKeypair() *keypairs.KeyPair {
	client := createComputeClient(c)
	client.Microversion = "2.2"

	kp, err := keypairs.Create(client, keypairs.CreateOpts{
		Name: "go-key",
		Type: "ssh",
	}).Extract()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("----DEBUG----")
	fmt.Println(kp.PrivateKey)
	fmt.Println("----DEBUG----")

	return kp
}

// CreateServer creates a compute instance in Openstack.
func (c *Client) CreateServer(keypair *keypairs.KeyPair, imageID, serverFlavorID, networkID string, enableConfigDrive bool) (*servers.Server, string) {
	trivyVersion := "0.31.3"
	client := createComputeClient(c)

	serverOpts := servers.CreateOpts{
		Name:           "Scanner",
		FlavorRef:      serverFlavorID,
		ImageRef:       imageID,
		SecurityGroups: []string{"default"},
		UserData: []byte(fmt.Sprintf(`#!/bin/bash
wget -q -O- "https://github.com/aquasecurity/trivy/releases/download/v%s/trivy_%s_Linux-64bit.tar.gz" | tar xzf -
chmod u+x trivy
sudo ./trivy rootfs -f json -o /tmp/results.json /
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
		c.RemoveKeypair()
		panic(err)
	}

	freeIP := attachFloatingIP(client, server.ID)

	return server, freeIP
}

// RemoveServer will delete a Server from Openstack
func (c *Client) RemoveServer(serverID string) {
	log.Println("removing scanning server.")
	client := createComputeClient(c)
	res := servers.Delete(client, serverID)

	if res.Err != nil {
		log.Println(res.Err)
	}
}

// RemoveKeypair will delete a Keypair from Openstack
func (c *Client) RemoveKeypair() {
	log.Println("removing keypair.")
	client := createComputeClient(c)
	res := keypairs.Delete(client, "go-key", keypairs.DeleteOpts{})
	if res.Err != nil {
		log.Println(res.Err)
	}
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
