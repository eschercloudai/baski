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

package publish

import (
	"github.com/drew-viles/baskio/pkg/constants"
	ostack "github.com/drew-viles/baskio/pkg/openstack"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	imageIDFlag                                               string
	resultsFileFlag                                           string
	ghUserFlag, ghProjectFlag, ghTokenFlag, ghPagesBranchFlag string
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
			constants.Envs.SetOpenstackEnvs()

			osClient := &ostack.Client{
				Env: constants.Envs,
			}
			osClient.OpenstackInit()

			pagesGitDir, _, err := fetchPagesRepo(ghUserFlag, ghTokenFlag, ghProjectFlag, ghPagesBranchFlag)
			if err != nil {
				log.Fatalln(err)
			}

			resultsFile, err := os.Open(resultsFileFlag)
			if err != nil {
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			img := getImageData(osClient)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = copyResultsFileIntoPages(pagesGitDir, img.Name, resultsFile)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			reports, err := fetchExistingReports(pagesGitDir)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			results, err := parseReports(reports, img)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = buildPages(pagesGitDir, results)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			//err = publishPages(pagesRepo, pagesGitDir)
			//checkErrorPagesWithCleanup(err, pagesGitDir)

			//pagesCleanup(pagesGitDir)
		},
	}

	cmd.Flags().StringVarP(&ghUserFlag, "github-user", "u", "", "The user for the GitHub project to which the pages will be pushed.")
	cmd.Flags().StringVarP(&ghProjectFlag, "github-project", "p", "", "The GitHub project to which the pages will be pushed.")
	cmd.Flags().StringVarP(&ghTokenFlag, "github-token", "t", "", "The token for the GitHub project to which the pages will be pushed.")
	cmd.Flags().StringVarP(&ghPagesBranchFlag, "github-pages-branch", "b", "gh-pages", "The branch name for GitHub project to which the pages will be pushed.")
	cmd.Flags().StringVarP(&imageIDFlag, "image-id", "i", "", "The ID of the image to scan.")
	cmd.Flags().StringVarP(&resultsFileFlag, "results-file", "r", "results.json", "The results file outputted by the scan.")

	requireFlag(cmd, "image-id")
	requireFlag(cmd, "github-user")
	requireFlag(cmd, "github-project")
	requireFlag(cmd, "github-token")

	return cmd

}

// checkErrorPagesWithCleanup takes an error and if it is not nil, will attempt to run a cleanup to ensure no resources are left lying around.
func checkErrorPagesWithCleanup(err error, dir string) {
	if err != nil {
		pagesCleanup(dir)
		log.Fatalln(err)
	}
}

// requireFlag sets flags as required.
func requireFlag(cmd *cobra.Command, flag string) {
	err := cmd.MarkFlagRequired(flag)
	if err != nil {
		log.Fatalln(err)
	}
}
