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
	"fmt"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"github.com/eschercloudai/baski/pkg/cmd/util/sign"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

type signValidateOptions struct {
	flags.GlobalFlags
	imageID    string
	publicKey  string
	vaultURL   string
	vaultToken string
	digest     string
}

func (o *signValidateOptions) addFlags(cmd *cobra.Command) {
	viperPrefix := "sign"
	viperVaultPrefix := fmt.Sprintf("%s.vault", viperPrefix)

	o.GlobalFlags.AddFlags(cmd)

	flags.StringVarWithViper(cmd, &o.imageID, viperPrefix, "image-id", "", "The image ID of the image to sign")
	flags.StringVarWithViper(cmd, &o.publicKey, viperPrefix, "public-key", "", "The path to the private key that will be used to sign the image")
	flags.StringVarWithViper(cmd, &o.digest, viperPrefix, "digest", "", "The digest to verify")
	flags.StringVarWithViper(cmd, &o.vaultURL, viperVaultPrefix, "url", "", "The Vault URL from which you will pull the private key")
	flags.StringVarWithViper(cmd, &o.vaultToken, viperVaultPrefix, "token", "", "The token used to log into vault")

	cmd.MarkFlagsRequiredTogether("url", "token")
	cmd.MarkFlagsMutuallyExclusive("url", "public-key")
}

// NewSignValidateCommand creates a command that allows the signing of an image.
func NewSignValidateCommand() *cobra.Command {

	o := &signValidateOptions{}
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate digital signature",
		Long: `Validate digital signature.

This just validates a signature. It's useful for verifying a signed image.
`,
		Run: func(cmd *cobra.Command, args []string) {
			var key []byte
			var err error
			imgID := getImageID()

			if len(viper.GetString("sign.public-key")) != 0 {
				key, err = os.ReadFile(viper.GetString("sign.public-key"))
				if err != nil {
					log.Fatalln(err)
				}
			} else if len(viper.GetString("sign.vault.url")) != 0 {
				key, err = sign.FetchPublicKeyFromVault(viper.GetString("sign.vault.url"), viper.GetString("sign.vault.token"))
				if err != nil {
					log.Fatalln(err)
				}
			}
			pubKey := sign.DecodePublicKey(key)

			valid, err := sign.Validate(imgID, pubKey, viper.GetString("sign.digest"))
			if err != nil {
				log.Fatalln(err)
			}

			log.Printf("The validation result was: %t", valid)
		},
	}
	o.addFlags(cmd)

	return cmd
}
