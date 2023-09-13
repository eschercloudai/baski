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
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"sync"
)

// NewScanExistingCommand creates a command that allows the scanning of an image.
func NewScanExistingCommand() *cobra.Command {
	o := &flags.ScanMultipleOptions{}

	cmd := &cobra.Command{
		Use:   "existing",
		Short: "Scan multiple existing images",
		Long: `Scan multiple existing images.

Retrospectively scanning images is required to make sure images stay secure or are taken out of circulation
as soon as possible when they are no longer secure. 
If the image fails the scan it will be tagged with metadata to mark it as insecure.
It will looks for any images starting with the prefix as defined in the config and scan all of those images.
Depending on how many images there are, this could take some time.
to prevent every single image being launched for a scan, the concurrency is limited to 5, this can be overridden in the config.'
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SetOptionsFromViper()

			if !trivy.ValidSeverity(strings.ToUpper(o.MaxSeverityType)) {
				return errors.New("severity value passed is invalid. Allowed values are: NONE, LOW, MEDIUM, HIGH, CRITICAL")
			}

			cloudsConfig := ostack.InitOpenstack(o.CloudsPath)
			cloudsConfig.SetOpenstackEnvs(o.CloudName)

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[o.CloudName])

			imgs, err := osClient.FetchAllImages(o.ImageSearch)
			if err != nil {
				return err
			}

			semaphore := make(chan struct{}, o.Concurrency)
			var wg sync.WaitGroup

			for i, img := range imgs {
				wg.Add(1)
				semaphore <- struct{}{}
				go func(image images.Image, id int) {
					defer func() {
						<-semaphore // Release the slot in the semaphore
					}()
					err = scanServer(osClient, o.ScanOptions, image, id, &wg)
					if err != nil {
						log.Println(err)
					}

				}(img, i)
			}
			wg.Wait()

			close(semaphore)

			return nil
		},
	}

	o.AddFlags(cmd)

	return cmd
}

func scanServer(osClient *ostack.Client, opts flags.ScanOptions, img images.Image, id int, wg *sync.WaitGroup) error {
	defer wg.Done()

	log.Printf("Processing Image %d with ID: %s\n", id, img.ID)
	err := runScan(osClient, &opts, &img)
	if err != nil {
		return err
	}

	log.Printf("Finished processing Image %d with ID: %s\n", id, img.ID)
	return nil
}
