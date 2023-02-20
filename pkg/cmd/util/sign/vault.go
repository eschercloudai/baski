package sign

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/api"
)

func FetchPrivateKeyFromVault(endpoint, token string) ([]byte, error) {
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
