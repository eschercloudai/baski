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
	"github.com/spf13/cobra"
)

// NewSignCommand creates a command that allows the signing of an image.
func NewSignCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign",
		Long: `Sign
Signing an image allows a user or system to validate that any images being used were indeed built by Baski. 
Using this command a user can generate the keys required to do the signing and then sign an image.
Signing is achieved by taking an image ID and using the hash of that to generate the digest. 
`,
	}
	commands := []*cobra.Command{
		NewSignGenerateCommand(),
		NewSignImageCommand(),
		NewSignValidateCommand(),
	}
	cmd.AddCommand(commands...)

	return cmd
}
