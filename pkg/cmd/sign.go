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

package cmd

import (
	"github.com/eschercloudai/baski/pkg/cmd/util/data"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

// NewSignCommand creates a command that allows the signing of an image.
func NewSignCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign image",
		Long: `Sign image.
Signing an image allows a user or system to validate that any images being used were indeed built by Baski. 
Using this command a user can generate the keys required to do the signing and then sign an image.
Signing is achieved by taking an image ID and using the hash of that to generate the digest. 
`,
		Run: func(cmd *cobra.Command, args []string) {
			var imgID string
			var err error

			cloudsConfig := ostack.InitOpenstack()

			if len(viper.GetString("sign.image-id")) == 0 {
				imgID, err = data.RetrieveNewImageID()
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				imgID = viper.GetString("sign.image-id")
			}

		},
	}

	cmd.Flags().StringVar(&imageIDFlag, "image-id", "", "The image ID of the image to sign")
	cmd.Flags().StringVar(&privateKeyFlag, "private-key", "", "The path to the private key that will be used to sign the image")
	cmd.Flags().StringVar(&vaultURLFlag, "vault-url", "", "The Vault URL from which you will pull the private key")
	cmd.Flags().StringVar(&vaultTokenFlag, "vault-token", "", "The token used to log into vault")

	cmd.MarkFlagsRequiredTogether("vault-url", "vault-token")
	cmd.MarkFlagsMutuallyExclusive("vault-url", "private-key")

	bindViper(cmd, "sign.image-id", "image-id")
	bindViper(cmd, "sign.vault.url", "vault-url")
	bindViper(cmd, "sign.vault.token", "vault-token")

	return cmd
}
