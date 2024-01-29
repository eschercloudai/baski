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
	"github.com/eschercloudai/baski/pkg/providers/openstack"
	"github.com/eschercloudai/baski/pkg/providers/scanner"
	"github.com/eschercloudai/baski/pkg/s3"
	"github.com/eschercloudai/baski/pkg/trivy"
	"github.com/eschercloudai/baski/pkg/util/flags"
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
			severity := trivy.Severity(strings.ToUpper(o.MaxSeverityType))
			if !trivy.ValidSeverity(severity) {
				return errors.New("severity value passed is invalid. Allowed values are: UNKNOWN, LOW, MEDIUM, HIGH, CRITICAL")
			}

			cloudProvider := ostack.NewCloudsProvider(o.CloudName)

			i, err := ostack.NewImageClient(cloudProvider)
			if err != nil {
				return err
			}

			c, err := ostack.NewComputeClient(cloudProvider)
			if err != nil {
				return err
			}

			n, err := ostack.NewNetworkClient(cloudProvider)
			if err != nil {
				return err
			}

			img, err := i.FetchImage(o.ImageID)

			if err != nil {
				return err
			}

			s3Conn, err := s3.New(o.Endpoint, o.AccessKey, o.SecretKey, o.ScanBucket, "")
			if err != nil {
				return err
			}

			s := scanner.NewScanner(c, i, n, s3Conn)

			err = s.RunScan(&o.ScanOptions, severity, img)
			if err != nil {
				return err
			}
			err = s.FetchScanResults()
			if err != nil {
				return err
			}
			err = s.CheckResultsTagImageAndUploadToS3(img, o.AutoDeleteImage, o.SkipCVECheck)
			if err != nil {
				return err
			}

			return nil
		},
	}

	o.AddFlags(cmd)

	return cmd
}
