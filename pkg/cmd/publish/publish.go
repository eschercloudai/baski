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

package publish

import (
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// NewPublishCommand creates a command that publishes CVE data to GitHub Pages - this will be deprecated soon
func NewPublishCommand() *cobra.Command {
	o := &flags.PublishOptions{}

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish CVE data",
		Long: `Publish CVE data.

Scanning and image produces a long report in json format. It's not pretty to read.
Sure, you can get a nice json formatter and attempt to do it that way or you can have a website generated for you in your
GitHub Pages and view the report there instead in a slightly nicer format.

The website it generates isn't the prettiest right now but it will be improved on over time.`,

		Run: func(cmd *cobra.Command, args []string) {
			o.SetOptionsFromViper()

			// just setting defaults for account if it's not provided. Presume it's the same as the username.
			if o.GithubAccount == "" {
				o.GithubAccount = o.GithubUser
			}

			cloudsConfig := ostack.InitOpenstack(o.CloudsPath)
			cloudsConfig.SetOpenstackEnvs(o.CloudName)

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[o.CloudName])

			pagesGitDir, pagesRepo, err := FetchPagesRepo(o)
			if err != nil {
				log.Fatalln(err)
			}

			resultsFile, err := os.Open("/tmp/results.json")
			if err != nil {
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			img := GetImageData(osClient, o.ImageID)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = CopyResultsFileIntoPages(pagesGitDir, img.Name, resultsFile)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			reports, err := FetchExistingReports(pagesGitDir)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			results, err := ParseReports(reports, img)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = BuildPages(pagesGitDir, results)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = PublishPages(pagesRepo, pagesGitDir)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			PagesCleanup(pagesGitDir)
		},
	}

	o.AddFlags(cmd)

	return cmd
}

// checkErrorPagesWithCleanup takes an error and if it is not nil, will attempt to run a cleanup to ensure no resources are left lying around.
func checkErrorPagesWithCleanup(err error, dir string) {
	if err != nil {
		PagesCleanup(dir)
		log.Fatalln(err)
	}
}
