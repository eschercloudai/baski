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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/eschercloudai/baski/pkg/util/flags"
	"github.com/eschercloudai/baski/pkg/util/sign"
	"github.com/spf13/cobra"
	"os"
	"path"
)

// NewSignGenerateCommand creates a command that allows the signing of an image.
func NewSignGenerateCommand() *cobra.Command {
	o := &flags.SignGenerateOptions{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate keys",
		Long: `Generate keys for signing images.
Using this command a user can generate the keys required to sign an image.
It will generate a public and private key that can then be stored securely.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SetOptionsFromViper()

			pk, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
			if err != nil {
				return err
			}

			dir := o.Path

			err = os.MkdirAll(dir, os.ModeDir)
			if err != nil {
				return err
			}

			fPriv, err := os.Create(path.Join(dir, "baski.key"))
			if err != nil {
				return err
			}
			defer fPriv.Close()

			fPub, err := os.Create(path.Join(dir, "baski.pub"))
			if err != nil {
				return err
			}
			defer fPub.Close()

			priv, pub := sign.EncodeKeys(pk)
			_, err = fPriv.Write(priv)
			if err != nil {
				return err
			}

			_, err = fPub.Write(pub)
			if err != nil {
				return err
			}

			return nil
		},
	}

	o.AddFlags(cmd)

	return cmd
}
