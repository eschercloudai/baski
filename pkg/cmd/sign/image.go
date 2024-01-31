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
	"github.com/drewbernetes/baski/pkg/provisoner"
	"github.com/drewbernetes/baski/pkg/util/flags"
	"github.com/drewbernetes/baski/pkg/util/sign"
	"github.com/spf13/cobra"
	"os"
)

// NewSignImageCommand creates a command that allows the signing of an image.
func NewSignImageCommand() *cobra.Command {
	o := &flags.SignOptions{}
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Sign image",
		Long: `Sign image.
Signing an image allows a user or system to validate that any images being used were indeed built by Baski.
Signing is achieved by taking an image ID and using the hash of that to generate the digest which is then
added as metadata to the image. A Public Key can then be used to validate the metadata.

If using vault, the key should be stored as follows:
* kv/path/to/secret
  * private-key: VALUE
  * public-key: VALUE
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SetOptionsFromViper()

			var key []byte
			var err error

			vaultClient := sign.VaultClient{
				Endpoint: o.VaultURL,
				Token:    o.VaultToken,
			}

			if len(o.PrivateKey) != 0 {
				key, err = os.ReadFile(o.PrivateKey)
				if err != nil {
					return err
				}
			} else if len(o.VaultURL) != 0 {
				key, err = vaultClient.Fetch(o.VaultMountPath, o.VaultSecretPath, "private-key")
				if err != nil {
					return err
				}
			}

			digest, err := sign.Sign(o.ImageID, key)
			if err != nil {
				return err
			}

			signer := provisoner.NewSigner(o)
			if err != nil {
				return err
			}

			err = signer.SignImage(digest)
			if err != nil {
				return err
			}

			return nil
		},
	}
	o.AddFlags(cmd)

	return cmd
}
