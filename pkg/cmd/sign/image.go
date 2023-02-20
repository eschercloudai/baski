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
	"github.com/eschercloudai/baski/pkg/cmd/util/data"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"github.com/eschercloudai/baski/pkg/cmd/util/sign"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

type signImageOptions struct {
	flags.GlobalFlags
	imageID     string
	privateKey  string
	vaultURL    string
	vaultRoleID string
	vaultSecret string
}

func (o *signImageOptions) addFlags(cmd *cobra.Command) {
	viperPrefix := "sign"
	viperVaultPrefix := fmt.Sprintf("%s.vault", viperPrefix)

	o.GlobalFlags.AddFlags(cmd)

	flags.StringVarWithViper(cmd, &o.imageID, viperPrefix, "image-id", "", "The image ID of the image to sign")
	flags.StringVarWithViper(cmd, &o.privateKey, viperPrefix, "private-key", "", "The path to the private key that will be used to sign the image")
	flags.StringVarWithViper(cmd, &o.vaultURL, viperVaultPrefix, "url", "", "The Vault URL from which you will pull the private key")
	flags.StringVarWithViper(cmd, &o.vaultSecret, viperVaultPrefix, "token", "", "The token used to log into vault")

	cmd.MarkFlagsRequiredTogether("url", "token")
	cmd.MarkFlagsMutuallyExclusive("url", "private-key")
}

// NewSignImageCommand creates a command that allows the signing of an image.
func NewSignImageCommand() *cobra.Command {

	o := &signImageOptions{}
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Sign image",
		Long: `Sign image.
Signing an image allows a user or system to validate that any images being used were indeed built by Baski.
Signing is achieved by taking an image ID and using the hash of that to generate the digest which is then 
added as metadata to the image. A Public Key can then be used to validate the metadata.

If using vault, the key should be stored as follows:

* kv/baski/signing-keys
* private-key
* public-key
`,
		Run: func(cmd *cobra.Command, args []string) {
			var key []byte
			var err error
			cloudsConfig := ostack.InitOpenstack()
			cloudsConfig.SetOpenstackEnvs()
			imgID := getImageID()

			if len(viper.GetString("sign.private-key")) != 0 {
				key, err = os.ReadFile(viper.GetString("sign.private-key"))
				if err != nil {
					log.Fatalln(err)
				}
			} else if len(viper.GetString("sign.vault.url")) != 0 {
				key, err = sign.FetchPrivateKeyFromVault(viper.GetString("sign.vault.url"), viper.GetString("sign.vault.token"))
				if err != nil {
					log.Fatalln(err)
				}
			}
			privKey := sign.DecodePrivateKey(key)

			digest, err := sign.Sign(imgID, privKey)
			if err != nil {
				log.Fatalln(err)
			}

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[viper.GetString("cloud-name")])
			_ = osClient.UpdateImageMetadata(imgID, digest)
		},
	}
	o.addFlags(cmd)

	return cmd
}

func getImageID() string {
	var imgID string
	var err error

	if len(viper.GetString("sign.image-id")) == 0 {
		imgID, err = data.RetrieveNewImageID()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		imgID = viper.GetString("sign.image-id")
	}

	return imgID
}
