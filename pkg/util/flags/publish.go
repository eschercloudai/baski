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

package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type PublishOptions struct {
	OpenStackCoreFlags
	ImageID           string
	GithubUser        string
	GithubAccount     string
	GithubProject     string
	GithubToken       string
	GithubPagesBranch string
	ResultsFile       string
}

func (o *PublishOptions) SetOptionsFromViper() {
	o.OpenStackCoreFlags.SetOptionsFromViper()

	o.ImageID = viper.GetString(fmt.Sprintf("%s.image-id", viperPublishPrefix))
	o.GithubUser = viper.GetString(fmt.Sprintf("%s.user", viperGithubPrefix))
	o.GithubAccount = viper.GetString(fmt.Sprintf("%s.account", viperGithubPrefix))
	o.GithubProject = viper.GetString(fmt.Sprintf("%s.project", viperGithubPrefix))
	o.GithubToken = viper.GetString(fmt.Sprintf("%s.token", viperGithubPrefix))
	o.GithubPagesBranch = viper.GetString(fmt.Sprintf("%s.pages-branch", viperGithubPrefix))
	o.ResultsFile = viper.GetString(fmt.Sprintf("%s.image-id", viperPublishPrefix))
}

func (o *PublishOptions) AddFlags(cmd *cobra.Command) {
	o.OpenStackCoreFlags.AddFlags(cmd, viperOpenStackPrefix)

	StringVarWithViper(cmd, &o.GithubUser, viperGithubPrefix, "user", "", "The user for the GitHub project to which the pages will be pushed")
	StringVarWithViper(cmd, &o.GithubProject, viperGithubPrefix, "project", "", "The GitHub project to which the pages will be pushed")
	StringVarWithViper(cmd, &o.GithubAccount, viperGithubPrefix, "account", "", "The account in which the project is stored. This will default to the user")
	StringVarWithViper(cmd, &o.GithubToken, viperGithubPrefix, "token", "", "The token for the GitHub project to which the pages will be pushed")
	StringVarWithViper(cmd, &o.GithubPagesBranch, viperGithubPrefix, "pages-branch", "gh-pages", "The branch name for GitHub project to which the pages will be pushed")
	StringVarWithViper(cmd, &o.ImageID, viperPublishPrefix, "image-id", "", "The ID of the image to publish the CVE results for")

	//TODO: this is currently not used or implemented in any way
	StringVarWithViper(cmd, &o.ResultsFile, viperPublishPrefix, "results-file", "", "The results file location")

	cmd.MarkFlagsRequiredTogether("user", "project", "token")
}
