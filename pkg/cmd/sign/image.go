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
	"github.com/eschercloudai/baski/pkg/cmd/util/data"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"github.com/eschercloudai/baski/pkg/cmd/util/sign"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// NewSignImageCommand creates a command that allows the signing of an image.
func NewSignImageCommand() *cobra.Command {

	o := &flags.SignImageOptions{}
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
			o.SetSignImageOptionsFromViper()

			var key []byte
			var err error
			cloudsConfig := ostack.InitOpenstack(o.CloudsPath)
			cloudsConfig.SetOpenstackEnvs(o.CloudName)
			imgID := getImageID(o.ImageID)

			if len(o.PrivateKey) != 0 {
				key, err = os.ReadFile(o.PrivateKey)
				if err != nil {
					log.Fatalln(err)
				}
			} else if len(o.VaultURL) != 0 {
				key, err = sign.FetchPrivateKeyFromVault(o.VaultURL, o.VaultToken)
				if err != nil {
					log.Fatalln(err)
				}
			}
			privKey := sign.DecodePrivateKey(key)

			digest, err := sign.Sign(imgID, privKey)
			if err != nil {
				log.Fatalln(err)
			}

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[o.CloudName])
			_ = osClient.UpdateImageMetadata(imgID, digest)
		},
	}
	o.AddFlags(cmd)

	return cmd
}

func getImageID(imageID string) string {
	var imgID string
	var err error

	if len(imageID) == 0 {
		imgID, err = data.RetrieveNewImageID()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		imgID = imageID
	}

	return imgID
}
