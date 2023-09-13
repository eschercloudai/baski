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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"log"
	"strconv"
	"strings"
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
		log.Fatalln(err)
	}
	epOpts := &gophercloud.EndpointOpts{
		Region: client.Cloud.RegionName,
	}
	client.Provider = provider
	client.EndpointOptions = epOpts

	return client
}
