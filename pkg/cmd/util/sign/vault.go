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

package sign

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/api"
)

func FetchPrivateKeyFromVault(endpoint, token, mountPath, secretPath string) ([]byte, error) {
	config := api.DefaultConfig()
	config.Address = endpoint

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	client.SetToken(token)

	secret, err := client.KVv2(mountPath).Get(context.Background(), secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	value, ok := secret.Data["private-key"].(string)
	if !ok {
		return nil, fmt.Errorf("value type assertion failed: %T %#v", secret.Data["private-key"], secret.Data["private-key"])
	}

	return []byte(value), nil
}

func FetchPublicKeyFromVault(endpoint, token string) ([]byte, error) {
	config := api.DefaultConfig()
	config.Address = endpoint

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	client.SetToken(token)

	secret, err := client.KVv2("kv/baski").Get(context.Background(), "signing-keys")
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	value, ok := secret.Data["public-key"].(string)
	if !ok {
		return nil, fmt.Errorf("value type assertion failed: %T %#v", secret.Data["public-key"], secret.Data["public-key"])
	}

	return []byte(value), nil
}
