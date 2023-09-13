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
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"log"
)

// GetFloatingIP will create a new FIP.
func (c *Client) GetFloatingIP(ipPool string) (*floatingips.FloatingIP, error) {
	log.Println("fetching floating IP")
	client := createComputeClient(c)
	createOpts := floatingips.CreateOpts{
		Pool: ipPool,
	}

	fip, err := floatingips.Create(client, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	return fip, nil
}

// RemoveFIP will delete a Floating IP from Openstack.
func (c *Client) RemoveFIP(fip *floatingips.FloatingIP) {
	log.Println("removing floating IP.")
	client := createComputeClient(c)
	res := floatingips.Delete(client, fip.ID)
	if res.Err != nil {
		log.Println(res.Err)
	}
}
