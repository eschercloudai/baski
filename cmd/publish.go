/*
Copyright 2022 EscherCloud.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/drew-viles/baskio/cmd/publish"
	ostack "github.com/drew-viles/baskio/pkg/openstack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

// NewPublishCommand creates a command that publishes CVE data to GitHub Pages.
func NewPublishCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish CVE data",
		Long: `Publish CVE data.

Scanning and image produces a long report in json format. It's not pretty to read.
Sure, you can get a nice json formatter and attempt to do it that way or you can have a website generated for you in your 
GitHub Pages and view the report there instead in a slightly nicer format.

The website it generates isn't the prettiest right now but it will be improved on over time.`,

		Run: func(cmd *cobra.Command, args []string) {
			cloudsConfig := ostack.InitOpenstack()
			cloudsConfig.SetOpenstackEnvs()

			osClient := &ostack.Client{
				Cloud: cloudsConfig.Clouds[viper.GetString("cloud-name")],
			}
			osClient.OpenstackInit()

			pagesGitDir, pagesRepo, err := publish.FetchPagesRepo(viper.GetString("publish.github.user"), viper.GetString("publish.github.token"), viper.GetString("publish.github.project"), viper.GetString("publish.github.branch"))
			if err != nil {
				log.Fatalln(err)
			}

			resultsFile, err := os.Open(viper.GetString("publish.results-file"))
			if err != nil {
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			img := publish.GetImageData(osClient, viper.GetString("publish.image-id"))
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = publish.CopyResultsFileIntoPages(pagesGitDir, img.Name, resultsFile)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			reports, err := publish.FetchExistingReports(pagesGitDir)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			results, err := publish.ParseReports(reports, img)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = publish.BuildPages(pagesGitDir, results)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = publish.PublishPages(pagesRepo, pagesGitDir)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			publish.PagesCleanup(pagesGitDir)
		},
	}

	cmd.Flags().StringVar(&ghUserFlag, "github-user", "", "The user for the GitHub project to which the pages will be pushed")
	cmd.Flags().StringVar(&ghProjectFlag, "github-project", "", "The GitHub project to which the pages will be pushed")
	cmd.Flags().StringVar(&ghTokenFlag, "github-token", "", "The token for the GitHub project to which the pages will be pushed")
	cmd.Flags().StringVar(&ghPagesBranchFlag, "github-pages-branch", "gh-pages", "The branch name for GitHub project to which the pages will be pushed")
	cmd.Flags().StringVar(&imageIDFlag, "image-id", "", "The ID of the image to scan")
	cmd.Flags().StringVar(&resultsFileFlag, "results-file", "results.json", "The results file outputted by the scan")

	//requireFlag(cmd, "image-id")
	//requireFlag(cmd, "github-user")
	//requireFlag(cmd, "github-project")
	//requireFlag(cmd, "github-token")

	//cmd.MarkFlagsRequiredTogether("image-id", "github-user", "github-project", "github-token")

	bindViper(cmd, "publish.image-id", "image-id")
	bindViper(cmd, "publish.github.user", "github-user")
	bindViper(cmd, "publish.github.project", "github-project")
	bindViper(cmd, "publish.github.token", "github-token")
	bindViper(cmd, "publish.github.pages-branch", "github-pages-branch")
	bindViper(cmd, "publish.results-file", "results-file")

	return cmd

}

// checkErrorPagesWithCleanup takes an error and if it is not nil, will attempt to run a cleanup to ensure no resources are left lying around.
func checkErrorPagesWithCleanup(err error, dir string) {
	if err != nil {
		publish.PagesCleanup(dir)
		log.Fatalln(err)
	}
}
