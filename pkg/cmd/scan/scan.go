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

package scan

import (
	"errors"
	"github.com/drewbernetes/baski/pkg/provisoner"
	"github.com/drewbernetes/baski/pkg/trivy"
	"github.com/drewbernetes/baski/pkg/util/flags"
	"github.com/spf13/cobra"
	"strings"
)

// NewScanCommand creates a command that allows the scanning of an image.
func NewScanCommand() *cobra.Command {

	o := &flags.ScanOptions{}
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan image",
		Long: `Scan image.

Scanning an image requires the creation of a new instance in Openstack using the image you want to scan.
Then, Trivy needs downloading and running against the filesystem. Again, this is time consuming.

The scan section of Baski fixes this for you and allows you to drink <enter drink here> whilst it does the hard work for you.

It does the following:
* Creates a new instance using the provided Openstack configuration variables
* Check if Trivy is available already, if not it'll download it
* Scans the rootfs
* Generates a report file that you can read with your eyes or via other means

If the checks for CVE flags/config values are set then it will bail out and generate a report with the CVEs that caused it to do so.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SetOptionsFromViper()
			if !trivy.ValidSeverity(trivy.Severity(strings.ToUpper(o.MaxSeverityType))) {
				return errors.New("severity value passed is invalid. Allowed values are: UNKNOWN, LOW, MEDIUM, HIGH, CRITICAL")
			}

			scan := provisoner.NewScanner(o)

			err := scan.Prepare()
			if err != nil {
				return err
			}

			err = scan.ScanImages()
			if err != nil {
				return err
			}

			return nil
		},
	}

	o.AddFlags(cmd)

	return cmd
}
