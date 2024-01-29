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
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"log"
)

type NetworkClient struct {
	client *gophercloud.ServiceClient
}

func NewNetworkClient(provider Provider) (*NetworkClient, error) {
	p, err := provider.Client()
	if err != nil {
		return nil, err
	}
	client, err := openstack.NewNetworkV2(p, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}
	return &NetworkClient{
		client: client,
	}, nil
}

// getNetwork will fetch a network by name
func (n *NetworkClient) getNetwork(networkName string) (*networks.Network, error) {
	log.Println("fetching floating IP")

	affirmative := true
	page, err := networks.List(n.client, &external.ListOptsExt{ListOptsBuilder: &networks.ListOpts{}, External: &affirmative}).AllPages()
	if err != nil {
		return nil, err
	}

	var results []networks.Network
	var net networks.Network

	if err = networks.ExtractNetworksInto(page, &results); err != nil {
		return nil, err
	}

	for _, network := range results {
		if network.Name == networkName {
			net = network
			break
		}
	}

	return &net, nil
}

// GetFloatingIP will create a new FIP.
func (n *NetworkClient) GetFloatingIP(networkName string) (*floatingips.FloatingIP, error) {
	log.Println("creating floating IP")

	network, err := n.getNetwork(networkName)
	if err != nil {
		return nil, err
	}
	createOpts := floatingips.CreateOpts{
		FloatingNetworkID: network.ID,
	}

	fip, err := floatingips.Create(n.client, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	return fip, nil
}

// RemoveFIP will delete a Floating IP from Openstack.
func (n *NetworkClient) RemoveFIP(fipID string) error {
	log.Println("removing floating IP.")
	res := floatingips.Delete(n.client, fipID)
	if res.Err != nil {
		return res.Err
	}
	return nil
}
