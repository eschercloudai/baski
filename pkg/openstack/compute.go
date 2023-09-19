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
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"log"
)

// createComputeClient will generate the compute client required for creating Servers and KeyPairs in Openstack.
func createComputeClient(client *Client) *gophercloud.ServiceClient {
	c, err := openstack.NewComputeV2(client.Provider, *client.EndpointOptions)
	if err != nil {
		panic(err)
	}

	return c
}

// CreateKeypair creates a new KeyPair in Openstack.
func (c *Client) CreateKeypair(keyNamePrefix string) (*keypairs.KeyPair, error) {
	log.Println("creating keypair")
	client := createComputeClient(c)
	client.Microversion = "2.2"

	kp, err := keypairs.Create(client, keypairs.CreateOpts{
		Name: keyNamePrefix + "-baski-key",
		Type: "ssh",
	}).Extract()
	if err != nil {
		return nil, err
	}

	return kp, nil
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

// CreateServer creates a compute instance in Openstack.
func (c *Client) CreateServer(keypair *keypairs.KeyPair, o *flags.ScanOptions, userData []byte, imageID string) (*servers.Server, error) {
	log.Println("creating server")
	client := createComputeClient(c)

	serverFlavorID := c.GetFlavorIDByName(o.FlavorName)

	serverOpts := servers.CreateOpts{
		Name:             imageID + "-scanner",
		FlavorRef:        serverFlavorID,
		ImageRef:         imageID,
		SecurityGroups:   []string{"default"},
		UserData:         userData,
		AvailabilityZone: "",
		Networks: []servers.Network{
			{
				UUID: o.NetworkID,
			},
		},
		ConfigDrive: &o.AttachConfigDrive,
		Min:         1,
		Max:         1,
	}

	createOpts := keypairs.CreateOptsExt{
		CreateOptsBuilder: serverOpts,
		KeyName:           keypair.Name,
	}

	server, err := servers.Create(client, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	return server, nil
}

// GetServerStatus gets the status of a server
func (c *Client) GetServerStatus(sid string) bool {
	client := createComputeClient(c)

	state, err := servers.Get(client, sid).Extract()
	if err != nil {
		log.Println(err)
		return false
	}
	if state.Status != "ACTIVE" {
		return false
	}

	return true
}

// AttachIP attaches the provided IP to the provided server.
func (c *Client) AttachIP(serverID, fip string) error {
	log.Println("attaching IP to server")
	client := createComputeClient(c)

	associateOpts := floatingips.AssociateOpts{
		FloatingIP: fip,
	}

	err := floatingips.AssociateInstance(client, serverID, associateOpts).ExtractErr()
	if err != nil {
		return err
	}
	return nil
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