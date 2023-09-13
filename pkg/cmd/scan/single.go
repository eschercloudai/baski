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

package scan

import (
	"errors"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/eschercloudai/baski/pkg/trivy"
	"github.com/spf13/cobra"
	"strings"
)

// NewScanSingleCommand creates a command that allows the scanning of an image.
func NewScanSingleCommand() *cobra.Command {
	o := &flags.ScanSingleOptions{}

	cmd := &cobra.Command{
		Use:   "single",
		Short: "Scan single image",
		Long: `Scan single image.

Scanning an a single image - useful for when an image has just been built.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SetOptionsFromViper()

			if !trivy.ValidSeverity(strings.ToUpper(o.MaxSeverityType)) {
				return errors.New("severity value passed is invalid. Allowed values are: NONE, LOW, MEDIUM, HIGH, CRITICAL")
			}

			cloudsConfig := ostack.InitOpenstack(o.CloudsPath)
			cloudsConfig.SetOpenstackEnvs(o.CloudName)

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[o.CloudName])

			img, err := osClient.FetchImage(o.ImageID)

			if err != nil {
				return err
			}

			err = runScan(osClient, &o.ScanOptions, img)
			if err != nil {
				return err
			}

			return nil
		},
	}

	o.AddFlags(cmd)

	return cmd
}
