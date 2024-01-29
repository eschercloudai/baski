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

package sign

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/api"
)

type VaultClient struct {
	Endpoint string
	Token    string
}

func (v *VaultClient) Fetch(mountPath, secretPath, data string) ([]byte, error) {
	config := api.DefaultConfig()
	config.Address = v.Endpoint

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w\n", err)
	}

	client.SetToken(v.Token)

	secret, err := client.KVv2(mountPath).Get(context.Background(), secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w\n", err)
	}

	value, ok := secret.Data[data].(string)
	if !ok {
		return nil, fmt.Errorf("value type assertion failed: %T %#v\n", secret.Data[data], secret.Data[data])
	}

	return []byte(value), nil
}
