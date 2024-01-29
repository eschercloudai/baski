/*
Copyright 2024 Drewbernetes.

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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"log"
)

type ComputeClient struct {
	client *gophercloud.ServiceClient
}

func NewComputeClient(provider Provider) (*ComputeClient, error) {
	p, err := provider.Client()
	if err != nil {
		return nil, err
	}
	client, err := openstack.NewComputeV2(p, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}
	return &ComputeClient{
		client: client,
	}, nil
}

// CreateKeypair creates a new KeyPair in Openstack.
func (c *ComputeClient) CreateKeypair(keyNamePrefix string) (*keypairs.KeyPair, error) {
	log.Println("creating keypair")
	c.client.Microversion = "2.2"

	kp, err := keypairs.Create(c.client, keypairs.CreateOpts{
		Name: keyNamePrefix + "-baski-key",
		Type: "ssh",
	}).Extract()
	if err != nil {
		return nil, err
	}

	return kp, nil
}

// RemoveKeypair will delete a Keypair from Openstack.
func (c *ComputeClient) RemoveKeypair(keyName string) error {
	log.Println("removing keypair.")
	res := keypairs.Delete(c.client, keyName, keypairs.DeleteOpts{})
	if res.Err != nil {
		return res.Err
	}
	return nil
}

// CreateServer creates a compute instance in Openstack.
func (c *ComputeClient) CreateServer(keypairName string, flavor, networkID string, attachConfigDrive *bool, userData []byte, imageID string) (*servers.Server, error) {
	log.Println("creating server")
	serverFlavorID, err := c.GetFlavorIDByName(flavor)
	if err != nil {
		return nil, err
	}

	serverOpts := servers.CreateOpts{
		Name:             imageID + "-scanner",
		FlavorRef:        serverFlavorID,
		ImageRef:         imageID,
		SecurityGroups:   []string{"default"},
		UserData:         userData,
		AvailabilityZone: "",
		Networks: []servers.Network{
			{
				UUID: networkID,
			},
		},
		ConfigDrive: attachConfigDrive,
		Min:         1,
		Max:         1,
	}

	createOpts := keypairs.CreateOptsExt{
		CreateOptsBuilder: serverOpts,
		KeyName:           keypairName,
	}

	server, err := servers.Create(c.client, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	return server, nil
}

// GetServerStatus gets the status of a server
func (c *ComputeClient) GetServerStatus(sid string) (bool, error) {
	state, err := servers.Get(c.client, sid).Extract()
	if err != nil {
		return false, err
	}

	if state.Status != "ACTIVE" {
		return false, nil
	}

	return true, nil
}

// AttachIP attaches the provided IP to the provided server.
func (c *ComputeClient) AttachIP(serverID, fip string) error {
	log.Println("attaching IP to server")
	associateOpts := floatingips.AssociateOpts{
		FloatingIP: fip,
	}

	err := floatingips.AssociateInstance(c.client, serverID, associateOpts).ExtractErr()
	if err != nil {
		return err
	}
	return nil
}

// RemoveServer will delete a Server from Openstack.
func (c *ComputeClient) RemoveServer(serverID string) error {
	log.Println("removing scanning server")
	res := servers.Delete(c.client, serverID)

	if res.Err != nil {
		return res.Err
	}
	return nil
}

// GetFlavorIDByName will take a name of a flavor and attempt to find the ID from Openstack.
func (c *ComputeClient) GetFlavorIDByName(name string) (string, error) {
	listOpts := flavors.ListOpts{
		AccessType: flavors.PublicAccess,
	}

	allPages, err := flavors.ListDetail(c.client, listOpts).AllPages()
	if err != nil {
		return "", err
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return "", err
	}

	for _, flavor := range allFlavors {
		if flavor.Name == name {
			return flavor.ID, nil
		}
	}
	return "", nil
}
