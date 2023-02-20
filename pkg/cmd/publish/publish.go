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
	"fmt"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

type publishOptions struct {
	flags.GlobalFlags
	imageID       string
	ghUser        string
	ghAccount     string
	ghProject     string
	ghToken       string
	ghPagesBranch string
	ResultsFile   string
}

func (o *publishOptions) addFlags(cmd *cobra.Command) {
	viperPrefix := "publish"
	viperGithubPrefix := fmt.Sprintf("%s.github", viperPrefix)

	o.GlobalFlags.AddFlags(cmd)

	flags.StringVarWithViper(cmd, &o.ghUser, viperGithubPrefix, "user", "", "The user for the GitHub project to which the pages will be pushed")
	flags.StringVarWithViper(cmd, &o.ghProject, viperGithubPrefix, "project", "", "The GitHub project to which the pages will be pushed")
	flags.StringVarWithViper(cmd, &o.ghAccount, viperGithubPrefix, "account", "", "The account in which the project is stored. This will default to the user")
	flags.StringVarWithViper(cmd, &o.ghToken, viperGithubPrefix, "token", "", "The token for the GitHub project to which the pages will be pushed")
	flags.StringVarWithViper(cmd, &o.ghPagesBranch, viperGithubPrefix, "pages-branch", "gh-pages", "The branch name for GitHub project to which the pages will be pushed")
	flags.StringVarWithViper(cmd, &o.imageID, viperPrefix, "image-id", "", "The ID of the image to scan")

	//TODO: this is currently not used or implemented in any way
	flags.StringVarWithViper(cmd, &o.ResultsFile, viperPrefix, "results-file", "", "The results file location")

	cmd.MarkFlagsRequiredTogether("user", "project", "token")
}

// NewPublishCommand creates a command that publishes CVE data to GitHub Pages.
func NewPublishCommand() *cobra.Command {
	o := &publishOptions{}

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish CVE data",
		Long: `Publish CVE data.

Scanning and image produces a long report in json format. It's not pretty to read.
Sure, you can get a nice json formatter and attempt to do it that way or you can have a website generated for you in your 
GitHub Pages and view the report there instead in a slightly nicer format.

The website it generates isn't the prettiest right now but it will be improved on over time.`,

		Run: func(cmd *cobra.Command, args []string) {
			// just setting defaults for account if it's not provided. Presume it's the same as the username.
			if viper.GetString("publish.github.account") == "" {
				viper.Set("publish.github.account", viper.GetString("publish.github.user"))
			}

			cloudsConfig := ostack.InitOpenstack()
			cloudsConfig.SetOpenstackEnvs()

			osClient := ostack.NewOpenstackClient(cloudsConfig.Clouds[viper.GetString("cloud-name")])

			pagesGitDir, pagesRepo, err := FetchPagesRepo(viper.GetString("publish.github.user"), viper.GetString("publish.github.token"), viper.GetString("publish.github.account"), viper.GetString("publish.github.project"), viper.GetString("publish.github.pages-branch"))
			if err != nil {
				log.Fatalln(err)
			}

			resultsFile, err := os.Open("/tmp/results.json")
			if err != nil {
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			img := GetImageData(osClient, viper.GetString("publish.image-id"))
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

	o.addFlags(cmd)

	return cmd
}

// checkErrorPagesWithCleanup takes an error and if it is not nil, will attempt to run a cleanup to ensure no resources are left lying around.
func checkErrorPagesWithCleanup(err error, dir string) {
	if err != nil {
		PagesCleanup(dir)
		log.Fatalln(err)
	}
}
