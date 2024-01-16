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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

type Provider interface {
	Client() (*gophercloud.ProviderClient, error)
}

// CloudsProvider creates a client from clouds.yaml.
type CloudsProvider struct {
	// cloud is the key to lookup in clouds.yaml.
	cloud string
}

var _ Provider = &CloudsProvider{}

// NewCloudsProvider creates the initial client for connecting to Openstack.
func NewCloudsProvider(cloud string) *CloudsProvider {
	return &CloudsProvider{
		cloud: cloud,
	}
}

func (c *CloudsProvider) Client() (*gophercloud.ProviderClient, error) {
	clientOps := &clientconfig.ClientOpts{
		Cloud: c.cloud,
	}
	options, err := clientconfig.AuthOptions(clientOps)
	if err != nil {
		return nil, fmt.Errorf("auth error: %s", err.Error())
	}

	return authenticatedClient(*options)
}
func authenticatedClient(opts gophercloud.AuthOptions) (*gophercloud.ProviderClient, error) {
	return openstack.AuthenticatedClient(opts)
}
