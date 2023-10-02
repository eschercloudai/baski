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
	ostack "github.com/eschercloudai/baski/pkg/providers/openstack"
	"github.com/eschercloudai/baski/pkg/util/data"
	"github.com/eschercloudai/baski/pkg/util/flags"
	"github.com/eschercloudai/baski/pkg/util/sign"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
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
* kv/path/to/secret
  * private-key: VALUE
  * public-key: VALUE
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SetOptionsFromViper()

			var key []byte
			var err error

			imgID := getImageID(o.ImageID)

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

			digest, err := sign.Sign(imgID, key)
			if err != nil {
				return err
			}

			cloudProvider := ostack.NewCloudsProvider(o.CloudName)

			i, err := ostack.NewImageClient(cloudProvider)
			if err != nil {
				return err
			}

			img, err := i.FetchImage(imgID)
			if err != nil {
				return err
			}

			// Default to replace unless the field isn't found below
			operation := images.ReplaceOp

			_, err = data.GetNestedField(img.Properties, "image", "metadata", "digest")
			if err != nil {
				operation = images.AddOp
			}

			log.Println("attaching digest to image metadata")
			_, err = i.ModifyImageMetadata(imgID, "digest", digest, operation)

			if err != nil {
				return err
			}

			return nil
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
